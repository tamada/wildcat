package wildcat

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/tamada/wildcat/errors"
)

// Wildcat is the struct treating to count the specified files, directories, and urls.
type Wildcat struct {
	config     *Config
	eitherChan chan *Either
	generator  Generator
	progress   Progress
}

// NewWildcat creates an instance of Wildcat.
func NewWildcat(opts *ReadOptions, runtimeOpts *RuntimeOptions, generator Generator) *Wildcat {
	channel := make(chan *Either)
	return &Wildcat{
		config:     NewConfig(ignores(".", !opts.NoIgnore, nil), opts, runtimeOpts, errors.New()),
		eitherChan: channel,
		generator:  generator,
		progress:   NewProgress(runtimeOpts.ShowProgress, runtimeOpts.ThreadNumber),
	}
}

func (wc *Wildcat) run(f func(Generator, *Config) *Either) {
	wc.progress.UpdateTarget()
	go func() {
		defer wc.progress.Done()
		either := f(wc.generator, wc.config)
		wc.eitherChan <- either
	}()
}

func (wc *Wildcat) CountEntries(entries []Entry) (*ResultSet, *errors.Center) {
	for _, entry := range entries {
		e := entry
		err := wc.handleItem(e)
		wc.config.ec.Push(err)
	}
	go func() {
		wc.progress.Wait()
		wc.Close()
	}()
	return wc.receiveImpl()
}

// CountAll counts the arguments in the given Argf.
func (wc *Wildcat) CountAll(argf *Argf) (*ResultSet, *errors.Center) {
	wc.progress.UpdateTarget()
	go func() {
		for _, arg := range argf.Arguments {
			err := wc.handleItem(arg)
			wc.config.ec.Push(err)
		}
		if len(argf.Arguments) == 0 {
			wc.handleEntry(&stdinEntry{index: NewOrder()})
		}
		wc.progress.Done()
	}()
	go func() {
		wc.progress.Wait()
		wc.Close()
	}()
	return wc.receiveImpl()
}

func (wc *Wildcat) receiveImpl() (*ResultSet, *errors.Center) {
	rs := NewResultSet()
	for either := range wc.eitherChan {
		receiveEither(either, rs, wc.config.ec)
	}
	return rs, wc.config.ec
}

// Close finishes the receiver object.
func (wc *Wildcat) Close() {
	close(wc.eitherChan)
}

func (wc *Wildcat) updateFileList(fileList bool) *Wildcat {
	newOpts := *wc.config.readOpts
	newOpts.FileList = fileList
	return wc.updateOpts(&newOpts)
}

// ReadFileListFromReader reads data from the given reader as the file list.
func (wc *Wildcat) ReadFileListFromReader(in io.Reader, index *Order) {
	reader := bufio.NewReader(in)
	order := index.Sub()
	newWc := wc.updateFileList(false)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" && !newWc.config.IsIgnore(line) {
			err := newWc.handleItem(NewArgWithIndex(order, line))
			newWc.config.ec.Push(err)
		}
		if err == io.EOF {
			break
		}
		order = order.Next()
	}
}

func (wc *Wildcat) handleDir(arg NameAndIndex) *Either {
	currentIgnore := ignores(arg.Name(), !wc.config.readOpts.NoIgnore, wc.config.ignore)
	fileInfos, err := ioutil.ReadDir(arg.Name())
	if err != nil {
		return &Either{Err: err}
	}
	index := arg.Index().Sub()
	for _, info := range fileInfos {
		newName := filepath.Join(arg.Name(), info.Name())
		if !isIgnore(wc.config.readOpts, currentIgnore, newName) {
			newWc := wc.updateIgnore(currentIgnore)
			err := newWc.handleItem(NewArgWithIndex(index, newName))
			newWc.config.ec.Push(err)
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
	wc.ReadFileListFromReader(reader, entry.Index())
	return &Either{Results: []*Result{}}
}

func (wc *Wildcat) handleEntry(entry Entry) *Either {
	targetEntry := entry
	if !wc.config.readOpts.NoExtract {
		newEntry, _ := ConvertToArchiveEntry(entry)
		targetEntry = newEntry
	}
	if wc.config.readOpts.FileList {
		return wc.handleEntryAsFileList(targetEntry)
	}
	wc.run(func(arg1 Generator, arg2 *Config) *Either {
		return targetEntry.Count(wc.generator)
	})
	return &Either{Results: []*Result{}}
}

func (wc *Wildcat) handleItem(arg NameAndIndex) error {
	name := NormalizePath(arg.Name())
	entry, ok := arg.(Entry)
	switch {
	case ok:
		wc.handleEntry(entry)
	case IsURL(name):
		wc.handleEntry(toURLEntry(arg, wc.config.runtimeOpts))
	case ExistDir(name):
		wc.handleDir(arg)
	case ExistFile(name):
		wc.handleEntry(NewFileEntryWithIndex(arg))
	default:
		return fmt.Errorf("%s: file or directory not found", name)
	}
	return nil
}

func (wc *Wildcat) updateIgnore(newIgnore Ignore) *Wildcat {
	return &Wildcat{
		config:     wc.config.updateIgnore(newIgnore),
		eitherChan: wc.eitherChan,
		generator:  wc.generator,
		progress:   wc.progress,
	}
}

func (wc *Wildcat) updateOpts(newOpts *ReadOptions) *Wildcat {
	return &Wildcat{
		config:     wc.config.updateOpts(newOpts),
		eitherChan: wc.eitherChan,
		generator:  wc.generator,
		progress:   wc.progress,
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
