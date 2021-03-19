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
	Options   *ReadOptions
	Arguments []*Arg
}

// Arg represents the one of command line arguments and its index.
type Arg struct {
	name  string
	index int
}

// NewArg creates an instance of Arg with given parameters.
func NewArg(index int, name string) *Arg {
	return &Arg{index: index, name: name}
}

// Name returns the name of receiver Arg object.
func (arg *Arg) Name() string {
	return arg.name
}

// Index returns the index of receiver Arg object.
func (arg *Arg) Index() int {
	return arg.index
}

func (arg *Arg) Reindex(newIndex int) {
	arg.index = newIndex
}

// NewArgf creates an instance of Argf for treating command line arguments.
func NewArgf(arguments []string, opts *ReadOptions) *Argf {
	entries := []*Arg{}
	for index, arg := range arguments {
		entries = append(entries, NewArg(index, arg))
	}
	return &Argf{Arguments: entries, Options: opts}
}

// Generator is the type for generating Counter object.
type Generator func() Counter

// DefaultGenerator is the default generator for counting all (bytes, characters, words, and lines).
var DefaultGenerator Generator = func() Counter { return NewCounter(All) }

func drainDataFromReader(in io.Reader, counter Counter) error {
	reader := bufio.NewReader(in)
	for {
		line, err := reader.ReadBytes('\n')
		counter.update(line)
		if err == io.EOF {
			break
		}
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
