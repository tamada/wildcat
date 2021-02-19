package wildcat

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type ReadOptions struct {
	FileList  bool
	NoIgnore  bool
	NoExtract bool
}

type Arguments struct {
	Options *ReadOptions
	Args    []string
}

type generator func() Counter

type source struct {
	in   io.Reader
	name string
}

type result struct {
	ec  *ErrorCenter
	gen generator
	rs  *ResultSet
}

func NewArguments() *Arguments {
	return &Arguments{Args: []string{}, Options: &ReadOptions{}}
}

func drainDataFromReader(in io.Reader, counter Counter) {
	reader := bufio.NewReader(in)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
			counter.update(line)
			break
		}
		counter.update(line)
	}
}

func countFromReader(s *source, r *result) {
	counter := r.gen()
	drainDataFromReader(s.in, counter)
	r.rs.Push(s.name, counter)
}

func (arg *Arguments) handleStdin(r *result, ignore Ignore) *result {
	if arg.Options.FileList {
		return arg.readFileList(os.Stdin, r, ignore)
	}
	countFromReader(&source{in: os.Stdin, name: "<stdin>"}, r)
	return r
}

func handleArchiveFile(item string, r *result) {
	traverser := newArchiveTraverser(item)
	file, err := os.Open(item)
	if err != nil {
		r.ec.Push(err)
		return
	}
	defer file.Close()
	traverser.traverseSource(&source{in: file, name: item}, r)
}

func countFile(fileName string, r *result) {
	file, err := os.Open(fileName)
	if err != nil {
		r.ec.Push(err)
		return
	}
	defer file.Close()
	countFromReader(&source{in: file, name: fileName}, r)
}

func (arg *Arguments) handleFile(item string, r *result, ignore Ignore) {
	if ignore != nil && ignore.IsIgnore(item) {
		return
	}
	if IsArchiveFile(item) && !arg.Options.NoExtract {
		handleArchiveFile(item, r)
	} else {
		countFile(item, r)
	}
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

func (arg *Arguments) handleDir(dirName string, r *result, ignore Ignore) {
	currentIgnore := ignores(dirName, !arg.Options.NoIgnore, ignore)
	fileInfos, err := ioutil.ReadDir(dirName)
	if err != nil {
		r.ec.Push(err)
		return
	}
	for _, fileInfo := range fileInfos {
		newName := filepath.Join(dirName, fileInfo.Name())
		if !isIgnore(arg.Options, ignore, newName) {
			arg.handleItem(newName, r, currentIgnore)
		}
	}
}

func (arg *Arguments) handleURL(item string, r *result) {
	if !arg.Options.NoExtract && IsArchiveFile(item) {
		handleArchiveURLFile(item, r)
	} else {
		arg.handleURLContent(item, r)
	}
}

func handleArchiveURLFile(item string, r *result) {

}

func (arg *Arguments) handleURLContent(item string, r *result) {
	response, err := http.Get(item)
	if err != nil {
		r.ec.Push(err)
		return
	}
	defer response.Body.Close()
	if response.StatusCode == 404 {
		r.ec.Push(fmt.Errorf("%s: not found", item))
		return
	}
	countFromReader(&source{name: item, in: response.Body}, r)
}

func (arg *Arguments) handleItem(item string, r *result, ignore Ignore) {
	if IsUrl(item) {
		arg.handleURL(item, r)
	} else if ExistDir(item) {
		arg.handleDir(item, r, ignore)
	} else if ExistFile(item) {
		arg.handleFile(item, r, ignore)
	} else {
		r.ec.Push(fmt.Errorf("%s: file or directory not found", item))
	}
}

func (arg *Arguments) readFileList(in io.Reader, r *result, ignore Ignore) *result {
	reader := bufio.NewReader(in)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			arg.handleItem(line, r, ignore)
		}
		if err == io.EOF {
			break
		}
	}
	return r
}

func (arg *Arguments) openFileAndReadFileList(item string, r *result, ignore Ignore) *result {
	file, err := os.Open(item)
	if err != nil {
		r.ec.Push(fmt.Errorf("%s: file not found (%s)", item, err.Error()))
		return r
	}
	defer file.Close()
	arg.readFileList(file, r, ignore)
	return r
}

func (arg *Arguments) handleArg(item string, r *result, ignore Ignore) {
	if arg.Options.FileList {
		arg.openFileAndReadFileList(item, r, ignore)
	} else {
		arg.handleItem(item, r, ignore)
	}
}

func (arg *Arguments) handleArgs(r *result, ignore Ignore) *result {
	for _, item := range arg.Args {
		arg.handleArg(item, r, ignore)
	}
	return r
}

func (arg *Arguments) CountAll(generator func() Counter, ec *ErrorCenter) *ResultSet {
	r := &result{rs: NewResultSet(), gen: generator, ec: ec}
	ignore := ignores(".", !arg.Options.NoIgnore, nil)
	if len(arg.Args) == 0 {
		arg.handleStdin(r, ignore)
	} else {
		arg.handleArgs(r, ignore)
	}
	return r.rs
}
