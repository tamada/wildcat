package wildcat

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Target interface {
	Count(counterGenerator func() Counter) *ResultSet
}

type file interface {
	Name() string
	Count(counter Counter)
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

type defaultFile struct {
	name   string
	reader io.ReadCloser
	err    error
}

func newDefaultFile(name string) file {
	return &defaultFile{name: name, reader: nil, err: nil}
}

func (df *defaultFile) Name() string {
	return df.name
}

func (df *defaultFile) Count(counter Counter) {
	reader, err := os.Open(df.Name())
	if err != nil {
		return
	}
	defer reader.Close()
	drainDataFromReader(reader, counter)
}

type sliceTarget struct {
	targets []string
}

func (st *sliceTarget) Count(counterGenerator func() Counter) *ResultSet {
	rs := NewResultSet()
	for _, t := range st.targets {
		counter := counterGenerator()
		file := newDefaultFile(t)
		file.Count(counter)
		rs.Push(t, counter)
	}
	return rs
}

type stdinTarget struct {
}

func (stdinT *stdinTarget) Count(counterGenerator func() Counter) *ResultSet {
	rs := NewResultSet()
	counter := counterGenerator()
	drainDataFromReader(os.Stdin, counter)
	rs.Push("<stdin>", counter)
	return rs
}

func (stdinT *stdinTarget) Size() int {
	return 1
}

func readFilesInDir(dirName string, ec *ErrorCenter, withIgnoreFile bool, ignore Ignore) []string {
	fileInfos, err := ioutil.ReadDir(dirName)
	if err != nil {
		ec.Push(err)
		return []string{}
	}
	targets := []string{}
	for _, fileInfo := range fileInfos {
		newName := filepath.Join(dirName, fileInfo.Name())
		targets = appendTargets(targets, newName, ec, withIgnoreFile, ignore)
		targets = ignore.Filter(targets)
	}
	return targets
}

func ignores(dir string, withIgnoreFile bool, parent Ignore) Ignore {
	if withIgnoreFile {
		return newIgnore(dir)
	}
	return &noIgnore{parent: parent}
}

func appendTargets(targets []string, arg string, ec *ErrorCenter, withIgnoreFile bool, ignore Ignore) []string {
	if ExistDir(arg) {
		ignore := ignores(arg, withIgnoreFile, ignore)
		files := readFilesInDir(arg, ec, withIgnoreFile, ignore)
		targets = append(targets, files...)
	} else if ExistFile(arg) {
		targets = append(targets, arg)
	}
	return targets
}

func findTargets(arg string, ec *ErrorCenter, withIgnoreFile bool, ignore Ignore) []string {
	if ExistDir(arg) {
		ignore := ignores(arg, withIgnoreFile, ignore)
		return readFilesInDir(arg, ec, withIgnoreFile, ignore)
	} else if ExistFile(arg) {
		return []string{arg}
	}
	ec.Push(fmt.Errorf("%s: file or directory not found", arg))
	return []string{}
}

func NewTargetWithIgnoreFile(arg string, ec *ErrorCenter) Target {
	return newTarget(arg, ec, true)
}

func NewTarget(arg string, ec *ErrorCenter) Target {
	return newTarget(arg, ec, false)
}

func newTarget(fileName string, ec *ErrorCenter, withIgnoreFile bool) Target {
	if IsArchiveFile(fileName) {
		return newArchiveTarget(fileName, ec)
	}
	return &sliceTarget{targets: findTargets(fileName, ec, withIgnoreFile, &noIgnore{})}
}

func readFileList(fileName string, ec *ErrorCenter) []string {
	file, err := os.Open(fileName)
	if err != nil {
		ec.Push(err)
		return []string{}
	}
	defer file.Close()
	return readFileListImpl(file, ec)
}

func readFileListImpl(plainReader io.Reader, ec *ErrorCenter) []string {
	targets := []string{}
	reader := bufio.NewReader(plainReader)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		fileName := string(line)
		if ExistFile(fileName) {
			targets = append(targets, fileName)
		}
	}
	return targets
}

func NewTargetFromFileList(args []string, ec *ErrorCenter) Target {
	targets := []string{}
	for _, file := range args {
		list := readFileList(file, ec)
		targets = append(targets, list...)
	}
	if len(args) == 0 {
		return &sliceTarget{targets: readFileListImpl(os.Stdin, ec)}
	}
	return &sliceTarget{targets: targets}
}

func NewStdinTarget() Target {
	return &stdinTarget{}
}
