package wildcat

import (
	"bufio"
	"io"
	"strings"
)

// ReadOptions represents the set of options about reading file.
type ReadOptions struct {
	FileList     bool
	NoIgnore     bool
	NoExtract    bool
	StoreContent bool
}

// Argf shows the command line arguments and stdin (if no command line arguments).
type Argf struct {
	Options *ReadOptions
	Entries []Entry
}

// NewArgf creates an instance of Argf for treating command line arguments.
func NewArgf(arguments []string, opts *ReadOptions) *Argf {
	entries := []Entry{}
	for index, arg := range arguments {
		entries = append(entries, &defaultEntry{fileName: arg, index: index})
	}
	return &Argf{Entries: entries, Options: opts}
}

// Generator is the type for generating Counter object.
type Generator func() Counter

// DefaultGenerator is the default generator for counting all (bytes, characters, words, and lines).
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
