package wildcat

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tamada/wildcat/errors"
)

// Targets shows the targets for counting.
type Targets struct {
	entries []Entry
}

// Config is the configuration object for counting.
type Config struct {
	ignore Ignore
	opts   *ReadOptions
	ec     *errors.Center
}

// NewConfig creates an instance of Config.
func NewConfig(ignore Ignore, opts *ReadOptions, ec *errors.Center) *Config {
	return &Config{ignore: ignore, opts: opts, ec: ec}
}

func (config *Config) updateOpts(newOpts *ReadOptions) *Config {
	return NewConfig(config.ignore, newOpts, config.ec)
}

func (config *Config) updateIgnore(newIgnore Ignore) *Config {
	return NewConfig(newIgnore, config.opts, config.ec)
}

// IsIgnore checks given line is the ignored file or not.
func (config *Config) IsIgnore(line string) bool {
	return config.ignore != nil && config.ignore.IsIgnore(line)
}

// Push adds the given entry to the receiver targets object.
func (targets *Targets) Push(entry Entry) {
	targets.entries = append(targets.entries, entry)
}

func (targets *Targets) urlTargets(entry Entry, opts *ReadOptions) {
	if !opts.NoExtract && IsArchiveFile(entry.Name()) {
		targets.Push(&archiveEntry{entry: entry})
	} else {
		targets.Push(entry)
	}
}

func (targets *Targets) readFileListFromFile(entry Entry, config *Config) {
	file, err := os.Open(entry.Name())
	if err != nil {
		config.ec.Push(err)
		return
	}
	defer file.Close()
	newOpts := *config.opts
	newOpts.FileList = false
	targets.ReadFileListFromReader(file, config.updateOpts(&newOpts))
}

func (targets *Targets) handleFile(entry Entry) {
	if IsArchiveFile(entry.Name()) {
		targets.Push(&archiveEntry{entry: entry})
	} else {
		targets.Push(entry)
	}
}

func (targets *Targets) fileTargets(entry Entry, config *Config) {
	if config.opts.FileList {
		targets.readFileListFromFile(entry, config)
	} else if !config.IsIgnore(entry.Name()) {
		targets.handleFile(entry)
	}
}

func (targets *Targets) dirTargets(entry Entry, config *Config) {
	currentIgnore := ignores(entry.Name(), !config.opts.NoIgnore, config.ignore)
	fileInfos, err := ioutil.ReadDir(entry.Name())
	if err != nil {
		config.ec.Push(err)
		return
	}
	for _, fileInfo := range fileInfos {
		newName := filepath.Join(entry.Name(), fileInfo.Name())
		if !isIgnore(config.opts, currentIgnore, newName) {
			targets.handleItem(&defaultEntry{fileName: newName}, config)
		}
	}
}

func (targets *Targets) handleItem(entry Entry, config *Config) {
	if IsURL(entry.Name()) {
		targets.urlTargets(toURLEntry(entry, config.opts), config.opts)
	} else if ExistDir(entry.Name()) {
		targets.dirTargets(entry, config)
	} else if ExistFile(entry.Name()) {
		targets.fileTargets(entry, config)
	} else {
		config.ec.Push(fmt.Errorf("%s: file or directory not found", entry.Name()))
	}
}

func (argf *Argf) pushEach(targets *Targets, config *Config) {
	for _, entry := range argf.Entries {
		targets.handleItem(entry, config)
	}
}

// ReadFileListFromReader reads from the given reader as the file list.
func (targets *Targets) ReadFileListFromReader(in io.Reader, config *Config) {
	reader := bufio.NewReader(in)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" && !config.IsIgnore(line) {
			targets.handleItem(&defaultEntry{fileName: line}, config)
		}
		if err == io.EOF {
			break
		}
	}
	targets.reindex()
}

func buildTargetsFromStdin(targets *Targets, config *Config) {
	if config.opts.FileList {
		newOpts := *config.opts
		newOpts.FileList = false
		targets.ReadFileListFromReader(os.Stdin, config.updateOpts(&newOpts))
	} else {
		targets.Push(&stdinEntry{index: 0})
	}
}

func (targets *Targets) reindex() {
	for index, entry := range targets.entries {
		switch entry := entry.(type) {
		case *defaultEntry:
			entry.index = index
		case *urlEntry:
			entry.index = index
		case *stdinEntry:
			entry.index = index
		case *archiveEntry:
			index := index
			entry.index = &index
		}
	}
}

// CollectTargets collects targets from Argf.
func (argf *Argf) CollectTargets() (*Targets, *errors.Center) {
	config := NewConfig(ignores(".", !argf.Options.NoIgnore, nil), argf.Options, errors.New())
	targets := &Targets{entries: []Entry{}}
	if len(argf.Entries) == 0 {
		buildTargetsFromStdin(targets, config)
	} else {
		argf.pushEach(targets, config)
	}
	targets.reindex()
	return targets, config.ec
}

// CountAll counts the bytes, characters, words, and lines of targets.
func (targets *Targets) CountAll(generator func() Counter) (*ResultSet, *errors.Center) {
	eitherChan := make(chan *Either)
	ec := errors.New()
	targets.countAllImpl(generator, eitherChan)
	return receive(eitherChan, ec)
}

func receiveItem(either *Either, rs *ResultSet, ec *errors.Center) {
	if either.Err != nil {
		ec.Push(either.Err)
	} else {
		for _, result := range either.Results {
			rs.Push(result)
		}
	}
}

func receive(eitherChan <-chan *Either, ec *errors.Center) (*ResultSet, *errors.Center) {
	rs := NewResultSet()
	for either := range eitherChan {
		receiveItem(either, rs, ec)
	}
	return rs, ec
}

func (targets *Targets) countAllImpl(generator func() Counter, eitherChan chan<- *Either) {
	wg := new(sync.WaitGroup)
	for _, entry := range targets.entries {
		entry := entry
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter := entry.Count(generator)
			eitherChan <- counter
		}()
	}

	go func() {
		wg.Wait()
		close(eitherChan)
	}()
}
