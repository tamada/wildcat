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
	if !opts.NoExtract {
		updatedEntry, ok := ConvertToArchiveEntry(entry)
		if ok {
			entry = updatedEntry
		}
	}
	targets.Push(entry)
}

func (targets *Targets) readFileListFromFile(arg *Arg, config *Config) {
	file, err := os.Open(arg.Name())
	if err != nil {
		config.ec.Push(err)
		return
	}
	defer file.Close()
	newOpts := *config.opts
	newOpts.FileList = false
	targets.ReadFileListFromReader(file, arg.Index(), config.updateOpts(&newOpts))
}

func (targets *Targets) handleFile(arg *Arg, config *Config) {
	var entry Entry = &FileEntry{nai: arg}
	if !config.opts.NoExtract {
		newEntry, _ := ConvertToArchiveEntry(entry)
		entry = newEntry
	}
	targets.Push(entry)
}

func (targets *Targets) fileTargets(arg *Arg, config *Config) {
	if config.opts.FileList {
		targets.readFileListFromFile(arg, config)
	} else if !config.IsIgnore(arg.Name()) {
		targets.handleFile(arg, config)
	}
}

func (targets *Targets) dirTargets(arg *Arg, config *Config) {
	currentIgnore := ignores(arg.Name(), !config.opts.NoIgnore, config.ignore)
	fileInfos, err := ioutil.ReadDir(arg.Name())
	if err != nil {
		config.ec.Push(err)
		return
	}
	index := arg.Index().Sub()
	for _, fileInfo := range fileInfos {
		newName := filepath.Join(arg.Name(), fileInfo.Name())
		if !isIgnore(config.opts, currentIgnore, newName) {
			targets.handleItem(NewArgWithIndex(index, newName), config)
			index = index.Next()
		}
	}
}

func (targets *Targets) handleItem(arg *Arg, config *Config) {
	if IsURL(arg.Name()) {
		targets.urlTargets(toURLEntry(arg, config.opts), config.opts)
	} else if ExistDir(arg.Name()) {
		targets.dirTargets(arg, config)
	} else if ExistFile(arg.Name()) {
		targets.fileTargets(arg, config)
	} else {
		config.ec.Push(fmt.Errorf("%s: file or directory not found", arg.Name()))
	}
}

func (argf *Argf) pushEach(targets *Targets, config *Config) {
	for _, arg := range argf.Arguments {
		targets.handleItem(arg, config)
	}
}

// ReadFileListFromReader reads from the given reader as the file list.
func (targets *Targets) ReadFileListFromReader(in io.Reader, index *Order, config *Config) {
	reader := bufio.NewReader(in)
	order := index.Sub()
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" && !config.IsIgnore(line) {
			targets.handleItem(NewArgWithIndex(order, line), config)
		}
		if err == io.EOF {
			break
		}
		order = order.Next()
	}
	targets.reindex()
}

func buildTargetsFromStdin(targets *Targets, config *Config) {
	if config.opts.FileList {
		newOpts := *config.opts
		newOpts.FileList = false
		targets.ReadFileListFromReader(os.Stdin, NewOrder(), config.updateOpts(&newOpts))
	} else {
		targets.Push(&stdinEntry{index: NewOrder()})
	}
}

func (targets *Targets) reindex() {
	for index, entry := range targets.entries {
		entry.Reindex(index)
	}
}

// CollectTargets collects targets from Argf.
func (argf *Argf) CollectTargets() (*Targets, *errors.Center) {
	config := NewConfig(ignores(".", !argf.Options.NoIgnore, nil), argf.Options, errors.New())
	targets := &Targets{entries: []Entry{}}
	if len(argf.Arguments) == 0 {
		buildTargetsFromStdin(targets, config)
	} else {
		argf.pushEach(targets, config)
	}
	targets.reindex()
	return targets, config.ec
}

// CountAll counts the bytes, characters, words, and lines of targets.
func (targets *Targets) CountAll(generator Generator) (*ResultSet, *errors.Center) {
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
