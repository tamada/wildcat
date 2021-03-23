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

// Wildcat is the struct treating to count the specified files, directories, and urls.
type Wildcat struct {
	config     *Config
	eitherChan chan *Either
	generator  Generator
	group      *sync.WaitGroup
}

// NewWildcat creates an instance of Wildcat.
func NewWildcat(opts *ReadOptions, generator Generator) *Wildcat {
	channel := make(chan *Either)
	return &Wildcat{
		config:     NewConfig(ignores(".", !opts.NoIgnore, nil), opts, errors.New()),
		eitherChan: channel,
		generator:  generator,
		group:      new(sync.WaitGroup),
	}
}

func (wc *Wildcat) run(f func(Generator, *Config) *Either) {
	wc.group.Add(1)
	go func() {
		defer wc.group.Done()
		either := f(wc.generator, wc.config)
		wc.eitherChan <- either
	}()
}

func (wc *Wildcat) CountEntries(entries []Entry) (*ResultSet, *errors.Center) {
	for _, entry := range entries {
		e := entry
		wc.run(func(generator Generator, config *Config) *Either {
			return wc.handleItem(e)
		})
	}
	go func() {
		wc.group.Wait()
		wc.Close()
	}()
	return wc.receiveImpl()
}

// CountAll counts the arguments in the given Argf.
func (wc *Wildcat) CountAll(argf *Argf) (*ResultSet, *errors.Center) {
	for _, arg := range argf.Arguments {
		arg := arg
		wc.run(func(generator Generator, config *Config) *Either {
			return wc.handleItem(arg)
		})
	}
	if len(argf.Arguments) == 0 {
		wc.countStdin()
	}
	go func() {
		wc.group.Wait()
		wc.Close()
	}()
	return wc.receiveImpl()
}

func (wc *Wildcat) receiveImpl() (*ResultSet, *errors.Center) {
	rs := NewResultSet()
	ec := errors.New()
	for either := range wc.eitherChan {
		receiveEither(either, rs, ec)
	}
	return rs, ec
}

// Close finishes the receiver object.
func (wc *Wildcat) Close() {
	close(wc.eitherChan)
}

// ReadFileListFromReader reads data from the given reader as the file list.
func (wc *Wildcat) ReadFileListFromReader(in io.Reader, index *Order) {
	reader := bufio.NewReader(in)
	order := index.Sub()
	newOpts := *wc.config.opts
	newOpts.FileList = false
	newWc := wc.updateOpts(&newOpts)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" && !newWc.config.IsIgnore(line) {
			order := order
			newWc.run(func(generator Generator, config *Config) *Either {
				return newWc.handleItem(NewArgWithIndex(order, line))
			})
		}
		if err == io.EOF {
			break
		}
		order = order.Next()
	}
}

func (wc *Wildcat) handleDir(arg NameAndIndex) *Either {
	currentIgnore := ignores(arg.Name(), !wc.config.opts.NoIgnore, wc.config.ignore)
	fileInfos, err := ioutil.ReadDir(arg.Name())
	if err != nil {
		wc.config.ec.Push(err)
		return &Either{Err: wc.config.ec}
	}
	index := arg.Index().Sub()
	for _, info := range fileInfos {
		newName := filepath.Join(arg.Name(), info.Name())
		if !isIgnore(wc.config.opts, currentIgnore, newName) {
			newWc := wc.updateIgnore(currentIgnore)
			index := index
			newWc.run(func(generator Generator, config *Config) *Either {
				return newWc.handleItem(NewArgWithIndex(index, newName))
			})
			index = index.Next()
		}
	}
	return &Either{Results: []*Result{}}
}

func (wc *Wildcat) handleEntryAsFileList(entry Entry) *Either {
	reader, err := entry.Open()
	defer reader.Close()
	if err != nil {
		return &Either{Err: err}
	}
	wc.ReadFileListFromReader(reader, entry.Index().Sub())
	return &Either{Results: []*Result{}}
}

func (wc *Wildcat) handleEntry(entry Entry) *Either {
	targetEntry := entry
	if !wc.config.opts.NoExtract {
		newEntry, _ := ConvertToArchiveEntry(entry)
		targetEntry = newEntry
	}
	if wc.config.opts.FileList {
		return wc.handleEntryAsFileList(targetEntry)
	}
	return targetEntry.Count(wc.generator)
}

func (wc *Wildcat) handleItem(arg NameAndIndex) *Either {
	// fmt.Fprintf(os.Stderr, "%s (%s)\n", arg.Name(), arg.Index())
	name := arg.Name()
	entry, ok := arg.(Entry)
	switch {
	case ok:
		return wc.handleEntry(entry)
	case IsURL(name):
		return wc.handleEntry(toURLEntry(arg, wc.config.opts))
	case ExistDir(name):
		return wc.handleDir(arg)
	case ExistFile(name):
		return wc.handleEntry(&FileEntry{nai: arg})
	default:
		wc.config.ec.Push(fmt.Errorf("%s: file or directory not found", name))
		return &Either{Err: wc.config.ec}
	}
}

func (wc *Wildcat) updateIgnore(newIgnore Ignore) *Wildcat {
	return &Wildcat{
		config:     wc.config.updateIgnore(newIgnore),
		eitherChan: wc.eitherChan,
		generator:  wc.generator,
		group:      wc.group,
	}
}

func (wc *Wildcat) updateOpts(newOpts *ReadOptions) *Wildcat {
	return &Wildcat{
		config:     wc.config.updateOpts(newOpts),
		eitherChan: wc.eitherChan,
		generator:  wc.generator,
		group:      wc.group,
	}
}

func (wc *Wildcat) countStdin() {
	if wc.config.opts.FileList {
		newOpts := *wc.config.opts
		newOpts.FileList = false
		wc.updateOpts(&newOpts).ReadFileListFromReader(os.Stdin, NewOrder())
	} else {
		wc.run(func(generator Generator, config *Config) *Either {
			return CountDefault(&stdinEntry{index: NewOrder()}, generator())
		})
	}
}

func receiveEither(either *Either, rs *ResultSet, ec *errors.Center) {
	if either.Err != nil {
		ec.Push(either.Err)
	} else {
		for _, result := range either.Results {
			rs.Push(result)
		}
	}
}
