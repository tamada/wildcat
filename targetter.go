package wildcat

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Target interface {
	Iter() <-chan file
	Count(counterGenerator func() Counter) *ResultSet
	Size() int
}

type file interface {
	Name() string
	Count(counter Counter)
}

type stdinFile struct {
}

func (stdin *stdinFile) Name() string {
	return "<stdin>"
}

func (stdin *stdinFile) Count(counter Counter) {
	reader := bufio.NewReader(os.Stdin)
	for {
		line, err := reader.ReadBytes('\n')
		if err == io.EOF {
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
	in := bufio.NewReader(reader)
	for {
		data, err := in.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		counter.update(data)
	}
}

type sliceTarget struct {
	targets []string
}

func (st *sliceTarget) Size() int {
	return len(st.targets)
}

func (st *sliceTarget) Count(counterGenerator func() Counter) *ResultSet {
	return countImpl(st, counterGenerator)
}

func (st *sliceTarget) Iter() <-chan file {
	ch := make(chan file)
	go func() {
		for _, t := range st.targets {
			ch <- newDefaultFile(t)
		}
		close(ch)
	}()
	return ch
}

func countImpl(targets Target, counterGenerator func() Counter) *ResultSet {
	rs := NewResultSet()
	for file := range targets.Iter() {
		counter := counterGenerator()
		file.Count(counter)
		rs.Push(file, counter)
	}
	return rs
}

type stdinTarget struct {
}

func (stdinT *stdinTarget) Count(counterGenerator func() Counter) *ResultSet {
	return countImpl(stdinT, counterGenerator)
}

func (stdinT *stdinTarget) Size() int {
	return 1
}

func (stdinT *stdinTarget) Iter() <-chan file {
	ch := make(chan file)
	go func() {
		ch <- &stdinFile{}
		close(ch)
	}()
	return ch
}

func readFilesInDir(dirName string, ec *ErrorCenter) []string {
	fileInfos, err := ioutil.ReadDir(dirName)
	if err != nil {
		ec.Push(err)
		return []string{}
	}
	targets := []string{}
	for _, fileInfo := range fileInfos {
		newName := filepath.Join(dirName, fileInfo.Name())
		targets = appendTargets(targets, newName, ec)
	}
	return targets
}

func appendTargets(targets []string, arg string, ec *ErrorCenter) []string {
	if ExistDir(arg) {
		files := readFilesInDir(arg, ec)
		targets = append(targets, files...)
	} else if ExistFile(arg) {
		targets = append(targets, arg)
	}
	return targets
}

func NewTarget(args []string, ec *ErrorCenter) Target {
	targets := []string{}
	for _, arg := range args {
		targets = appendTargets(targets, arg, ec)
	}
	return &sliceTarget{targets: targets}
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
