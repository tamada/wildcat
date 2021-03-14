package wildcat

import (
	"bufio"
	"io"
	"strings"
)

type ReadOptions struct {
	FileList     bool
	NoIgnore     bool
	NoExtract    bool
	StoreContent bool
}

type Argf struct {
	Options *ReadOptions
	Entries []Entry
}

func NewArgf(arguments []string, opts *ReadOptions) *Argf {
	entries := []Entry{}
	for index, arg := range arguments {
		entries = append(entries, &defaultEntry{fileName: arg, index: index})
	}
	return &Argf{Entries: entries, Options: opts}
}

type Generator func() Counter

var DefaultGenerator Generator = func() Counter { return NewCounter(All) }

func drainDataFromReader(in io.Reader, counter Counter) error {
	reader := bufio.NewReader(in)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			counter.update(line)
			break
		}
		counter.update(line)
	}
	return nil
}

func ignores(dir string, withIgnoreFile bool, parent Ignore) Ignore {
	if withIgnoreFile {
		return newIgnore(dir)
	}
	return &noIgnore{parent: parent}
}

func isIgnore(opts *ReadOptions, ignore Ignore, name string) bool {
	if !opts.NoIgnore {
		return ignore.IsIgnore(name) || strings.HasSuffix(name, ".gitignore")
	}
	return false
}
