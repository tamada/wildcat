package wildcat

import (
	"bufio"
	"io"
	"path/filepath"
	"strings"
)

// ReadOptions represents the set of options about reading file.
type ReadOptions struct {
	FileList  bool
	NoIgnore  bool
	NoExtract bool
	AllFiles  bool
}

type RuntimeOptions struct {
	ShowProgress bool
	ThreadNumber int64
	StoreContent bool
}

// Argf shows the command line arguments and stdin (if no command line arguments).
type Argf struct {
	Options     *ReadOptions
	RuntimeOpts *RuntimeOptions
	Arguments   []*Arg
}

// Arg represents the one of command line arguments and its index.
type Arg struct {
	name  string
	index *Order
}

// NewArg creates an instance of Arg with the given name.
func NewArg(name string) *Arg {
	return NewArgWithIndex(NewOrder(), name)
}

// NewArgWithIndex creates an instance of Arg with given parameters.
func NewArgWithIndex(index *Order, name string) *Arg {
	return &Arg{index: index, name: name}
}

// Name returns the name of receiver Arg object.
func (arg *Arg) Name() string {
	return arg.name
}

// Index returns the index of receiver Arg object.
func (arg *Arg) Index() *Order {
	return arg.index
}

// NewArgf creates an instance of Argf for treating command line arguments.
func NewArgf(arguments []string, opts *ReadOptions, runtimeOpts *RuntimeOptions) *Argf {
	entries := []*Arg{}
	for index, arg := range arguments {
		entries = append(entries, NewArgWithIndex(NewOrderWithIndex(index), arg))
	}
	return &Argf{Arguments: entries, Options: opts, RuntimeOpts: runtimeOpts}
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
		if err != nil {
			return err
		}
	}
	return nil
}

func ignores(dir string, withIgnoreFile bool, parent Ignore) Ignore {
	if withIgnoreFile {
		return newIgnoreWithParent(dir, parent)
	}
	return &noIgnore{parent: parent}
}

func isIgnore(opts *ReadOptions, ignore Ignore, name string) bool {
	base := filepath.Base(name)
	ignoreFlag := !opts.AllFiles && strings.HasPrefix(base, ".")
	if !opts.NoIgnore {
		return ignoreFlag || ignore.IsIgnore(name)
	}
	return ignoreFlag
}
