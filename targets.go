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

type Targets struct {
	entries []Entry
}

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

func (targets *Targets) readFileListFromFile(entry Entry, ignore Ignore, opts *ReadOptions, ec *errors.Center) {
	file, err := os.Open(entry.Name())
	if err != nil {
		ec.Push(err)
		return
	}
	defer file.Close()
	newOpts := *opts
	newOpts.FileList = false
	targets.ReadFileListFromReader(file, entry.Index(), ignore, &newOpts, ec)
}

func (targets *Targets) handleFile(entry Entry, opts *ReadOptions, ec *errors.Center) {
	if IsArchiveFile(entry.Name()) {
		targets.Push(&archiveEntry{entry: entry})
	} else {
		targets.Push(entry)
	}
}

func (targets *Targets) fileTargets(entry Entry, ignore Ignore, opts *ReadOptions, ec *errors.Center) {
	if opts.FileList {
		targets.readFileListFromFile(entry, ignore, opts, ec)
	} else if ignore == nil || !ignore.IsIgnore(entry.Name()) {
		targets.handleFile(entry, opts, ec)
	}
}

func (targets *Targets) dirTargets(entry Entry, ignore Ignore, opts *ReadOptions, ec *errors.Center) {
	currentIgnore := ignores(entry.Name(), !opts.NoIgnore, ignore)
	fileInfos, err := ioutil.ReadDir(entry.Name())
	if err != nil {
		ec.Push(err)
		return
	}
	for _, fileInfo := range fileInfos {
		newName := filepath.Join(entry.Name(), fileInfo.Name())
		if !isIgnore(opts, currentIgnore, newName) {
			targets.handleItem(&defaultEntry{fileName: newName}, ignore, opts, ec)
		}
	}
}

func (targets *Targets) handleItem(entry Entry, ignore Ignore, opts *ReadOptions, ec *errors.Center) {
	if IsUrl(entry.Name()) {
		targets.urlTargets(toURLEntry(entry, opts), opts)
	} else if ExistDir(entry.Name()) {
		targets.dirTargets(entry, ignore, opts, ec)
	} else if ExistFile(entry.Name()) {
		targets.fileTargets(entry, ignore, opts, ec)
	} else {
		ec.Push(fmt.Errorf("%s: file or directory not found", entry.Name()))
	}
}

func (argf *Argf) pushEach(targets *Targets, ignore Ignore, ec *errors.Center) {
	opts := argf.Options
	for _, entry := range argf.Entries {
		targets.handleItem(entry, ignore, opts, ec)
	}
}

func (targets *Targets) ReadFileListFromReader(in io.Reader, index int, ignore Ignore, opts *ReadOptions, ec *errors.Center) {
	reader := bufio.NewReader(in)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			if !ignore.IsIgnore(line) {
				targets.handleItem(&defaultEntry{fileName: line, index: index}, ignore, opts, ec)
			}
		}
		if err == io.EOF {
			break
		}
	}
	targets.reindex()
}

func (opts *ReadOptions) buildTargetsFromStdin(targets *Targets, ignore Ignore, ec *errors.Center) {
	if opts.FileList {
		newOpts := *opts
		newOpts.FileList = false
		targets.ReadFileListFromReader(os.Stdin, 0, ignore, &newOpts, ec)
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

func (targets *Targets) maxIndex() int {
	max := 0
	for _, entry := range targets.entries {
		if max < entry.Index() {
			max = entry.Index()
		}
	}
	return max
}

func (argf *Argf) CollectTargets() (*Targets, *errors.Center) {
	ec := errors.New()
	ignore := ignores(".", !argf.Options.NoIgnore, nil)
	targets := &Targets{entries: []Entry{}}
	if len(argf.Entries) == 0 {
		argf.Options.buildTargetsFromStdin(targets, ignore, ec)
	} else {
		argf.pushEach(targets, ignore, ec)
	}
	targets.reindex()
	return targets, ec
}

func (targets *Targets) CountAll(generator func() Counter) (*ResultSet, *errors.Center) {
	eitherChan := make(chan *Either)
	ec := errors.New()
	targets.countAllImpl(generator, eitherChan)
	return receive(eitherChan, ec)
}

func receive(eitherChan <-chan *Either, ec *errors.Center) (*ResultSet, *errors.Center) {
	rs := NewResultSet()
	for either := range eitherChan {
		if either.Err != nil {
			ec.Push(either.Err)
		} else {
			for _, result := range either.Results {
				rs.Push(result)
			}
		}
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
