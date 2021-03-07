package wildcat

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/tamada/wildcat/errors"
)

type ReadOptions struct {
	FileList     bool
	NoIgnore     bool
	NoExtract    bool
	StoreContent bool
}

type Entry interface {
	Name() string
	Open() (io.ReadCloser, error)
}

type defaultEntry struct {
	fileName string
}

func (de *defaultEntry) Name() string {
	return de.fileName
}

func (de *defaultEntry) Open() (io.ReadCloser, error) {
	return os.Open(de.fileName)
}

type Argf struct {
	Options *ReadOptions
	Entries []Entry
}

func NewArgf(arguments []string, opts *ReadOptions) *Argf {
	entries := []Entry{}
	for _, arg := range arguments {
		entries = append(entries, &defaultEntry{fileName: arg})
	}
	return &Argf{Entries: entries, Options: opts}
}

type Generator func() Counter

var DefaultGenerator Generator = func() Counter { return NewCounter(All) }

type Source struct {
	in   io.Reader
	name string
}

func NewSource(name string, in io.Reader) *Source {
	return &Source{name: name, in: in}
}

type DataSink struct {
	ec  *errors.Center
	gen Generator
	rs  *ResultSet
}

func (ds *DataSink) Dump(printerType string, sizer Sizer) []byte {
	buffer := bytes.NewBuffer([]byte{})
	printer := NewPrinter(buffer, printerType, sizer)
	ds.rs.Print(printer)
	return buffer.Bytes()
}

func (ds *DataSink) ResultSet() *ResultSet {
	return ds.rs
}

func (ds *DataSink) Error() error {
	if ds.ec.IsEmpty() {
		return nil
	}
	return ds.ec
}

func NewDataSink(gen Generator, ec *errors.Center) *DataSink {
	return &DataSink{gen: gen, ec: ec, rs: NewResultSet()}
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

func countFromReader(s *Source, r *DataSink) {
	counter := r.gen()
	drainDataFromReader(s.in, counter)
	r.rs.Push(s.name, counter)
}

func (opts *ReadOptions) handleReader(s *Source, r *DataSink, ignore Ignore) *DataSink {
	if opts.FileList {
		return opts.readFileList(s.in, r, ignore)
	}
	countFromReader(s, r)
	return r
}

func (opts *ReadOptions) handleStdin(r *DataSink, ignore Ignore) *DataSink {
	return opts.handleReader(NewSource("<stdin>", os.Stdin), r, ignore)
}

func handleArchiveFile(item Entry, r *DataSink) {
	traverser := newArchiveTraverser(item.Name())
	file, err := item.Open()
	if err != nil {
		r.ec.Push(err)
		return
	}
	defer file.Close()
	traverser.traverseSource(NewSource(item.Name(), file), r)
}

func countFile(entry Entry, r *DataSink) {
	file, err := entry.Open()
	if err != nil {
		r.ec.Push(err)
		return
	}
	defer file.Close()
	countFromReader(NewSource(entry.Name(), file), r)
}

func (opts *ReadOptions) HandleFile(item Entry, r *DataSink, ignore Ignore) {
	if ignore != nil && ignore.IsIgnore(item.Name()) {
		return
	}
	if IsArchiveFile(item.Name()) && !opts.NoExtract {
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

func (opts *ReadOptions) handleDir(dirName Entry, r *DataSink, ignore Ignore) {
	currentIgnore := ignores(dirName.Name(), !opts.NoIgnore, ignore)
	fileInfos, err := ioutil.ReadDir(dirName.Name())
	if err != nil {
		r.ec.Push(err)
		return
	}
	for _, fileInfo := range fileInfos {
		newName := filepath.Join(dirName.Name(), fileInfo.Name())
		if !isIgnore(opts, ignore, newName) {
			opts.handleItem(&defaultEntry{fileName: newName}, r, currentIgnore)
		}
	}
}

func (opts *ReadOptions) handleURL(item Entry, r *DataSink) {
	execFunc := countFromReader
	if !opts.NoExtract && IsArchiveFile(item.Name()) {
		execFunc = countArchiveFromReader
	}
	opts.handleURLContent(item, r, execFunc)
}

func countArchiveFromReader(s *Source, r *DataSink) {
	traverser := newArchiveTraverser(s.name)
	traverser.traverseSource(s, r)
}

func createTeeReader(reader io.ReadCloser, name string) (io.ReadCloser, error) {
	u, err := url.Parse(name)
	if err != nil {
		return nil, fmt.Errorf("url.Parse failed: %w", err)
	}
	newName := path.Base(u.Path)
	writer, err := os.Create(newName)
	if err != nil {
		return nil, fmt.Errorf("%s: file not found (%w)", newName, err)
	}
	return newMyTeeReader(reader, writer), nil
}

func (opts *ReadOptions) openReader(item Entry) (io.ReadCloser, error) {
	reader, err := item.Open()
	if err != nil {
		return nil, fmt.Errorf("%s: open error (%w)", item.Name(), err)
	}
	if opts.StoreContent {
		return createTeeReader(reader, item.Name())
	}
	return reader, nil
}

func (opts *ReadOptions) handleURLContent(item Entry, r *DataSink, execFunc func(*Source, *DataSink)) {
	reader, err := opts.openReader(item)
	if err != nil {
		r.ec.Push(err)
		return
	}
	defer reader.Close()
	source := NewSource(item.Name(), reader)
	execFunc(source, r)
}

func (mtr *myTeeReader) Read(p []byte) (n int, err error) {
	return mtr.tee.Read(p)
}

func (mtr *myTeeReader) Close() error {
	err1 := mtr.reader.Close()
	err2 := mtr.writer.Close()
	if err1 != nil {
		return err1
	}
	if err2 != nil {
		return err2
	}
	return nil
}

func newMyTeeReader(reader io.ReadCloser, writer io.WriteCloser) *myTeeReader {
	tee := &myTeeReader{reader: reader, writer: writer}
	tee.tee = io.TeeReader(reader, writer)
	return tee
}

type myTeeReader struct {
	reader io.ReadCloser
	writer io.WriteCloser
	tee    io.Reader
}

type urlEntry struct {
	url string
}

func (ue *urlEntry) Name() string {
	return ue.url
}

func (ue *urlEntry) Open() (io.ReadCloser, error) {
	response, err := http.Get(ue.url)
	if err != nil {
		return nil, fmt.Errorf("%s: http error: %w", ue.url, err)
	}
	if response.StatusCode == 404 {
		defer response.Body.Close()
		return nil, fmt.Errorf("%s: file not found", ue.url)
	}
	return response.Body, nil
}

func toURLEntry(entry Entry) Entry {
	return &urlEntry{url: entry.Name()}
}

func (opts *ReadOptions) handleItem(item Entry, r *DataSink, ignore Ignore) {
	if IsUrl(item.Name()) {
		opts.handleURL(toURLEntry(item), r)
	} else if ExistDir(item.Name()) {
		opts.handleDir(item, r, ignore)
	} else {
		opts.HandleFile(item, r, ignore)
	}
}

func (opts *ReadOptions) readFileList(in io.Reader, r *DataSink, ignore Ignore) *DataSink {
	reader := bufio.NewReader(in)
	for {
		line, err := reader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line != "" {
			opts.handleItem(&defaultEntry{fileName: line}, r, ignore)
		}
		if err == io.EOF {
			break
		}
	}
	return r
}

func (opts *ReadOptions) openFileAndReadFileList(item Entry, r *DataSink, ignore Ignore) *DataSink {
	file, err := item.Open()
	if err != nil {
		r.ec.Push(fmt.Errorf("%s: file not found (%w)", item, err))
		return r
	}
	defer file.Close()
	opts.readFileList(file, r, ignore)
	return r
}

func (opts *ReadOptions) HandleArg(item Entry, r *DataSink, ignore Ignore) {
	if opts.FileList {
		opts.openFileAndReadFileList(item, r, ignore)
	} else {
		opts.handleItem(item, r, ignore)
	}
}

func (argf *Argf) handleArgs(r *DataSink, ignore Ignore) *DataSink {
	for _, item := range argf.Entries {
		argf.Options.HandleArg(item, r, ignore)
	}
	return r
}

func (argf *Argf) CountAll(generator func() Counter, ec *errors.Center) (*ResultSet, error) {
	r := NewDataSink(generator, ec)
	ignore := ignores(".", !argf.Options.NoIgnore, nil)
	if len(argf.Entries) == 0 {
		argf.Options.handleStdin(r, ignore)
	} else {
		argf.handleArgs(r, ignore)
	}
	return r.rs, r.Error()
}
