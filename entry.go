package wildcat

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

type nameIndex interface {
	Name() string
	Index() int
}

type indexString struct {
	value string
	index int
}

func (is *indexString) Name() string {
	return is.value
}

func (is *indexString) Index() int {
	return is.index
}

type Entry interface {
	nameIndex
	Open() (io.ReadCloser, error)
	Count(generator func() Counter) *Either
}

type Either struct {
	Err     error
	Results []*Result
}

type Result struct {
	nameIndex nameIndex
	counter   Counter
}

func newResult(entry Entry, counter Counter) *Result {
	return &Result{
		nameIndex: entry,
		counter:   counter,
	}
}

type stdinEntry struct {
	index int
}

func CountDefault(entry Entry, counter Counter) *Either {
	reader, err := entry.Open()
	if err != nil {
		return &Either{Err: err}
	}
	drainDataFromReader(reader, counter)
	return &Either{Results: []*Result{newResult(entry, counter)}}
}

func (se *stdinEntry) Name() string {
	return "<stdin>"
}

func (se *stdinEntry) Count(generator func() Counter) *Either {
	return CountDefault(se, generator())
}

func (se *stdinEntry) Open() (io.ReadCloser, error) {
	return os.Stdin, nil
}

func (se *stdinEntry) Index() int {
	return se.index
}

func NewArchiveEntry(entry Entry) Entry {
	return &archiveEntry{entry: entry}
}

type archiveEntry struct {
	entry Entry
	index *int
}

func (ae *archiveEntry) Name() string {
	return ae.entry.Name()
}

func (ae *archiveEntry) Open() (io.ReadCloser, error) {
	return ae.entry.Open()
}

func (ae *archiveEntry) Count(generator func() Counter) *Either {
	archiver := newArchiver(ae)
	file, err := ae.Open()
	if err != nil {
		return &Either{Err: err}
	}
	defer file.Close()
	return archiver.traverse(generator)
}

func (ae *archiveEntry) Index() int {
	if ae.index != nil {
		return *ae.index
	}
	return ae.entry.Index()
}

type defaultEntry struct {
	fileName string
	index    int
}

func (de *defaultEntry) Name() string {
	return de.fileName
}

func (de *defaultEntry) Open() (io.ReadCloser, error) {
	return os.Open(de.fileName)
}

func (de *defaultEntry) Count(generator func() Counter) *Either {
	return CountDefault(de, generator())
}

func (de *defaultEntry) Index() int {
	return de.index
}

type downloadUrlEntry struct {
	entry *urlEntry
}

func (due *downloadUrlEntry) Index() int {
	return due.entry.Index()
}

func (due *downloadUrlEntry) Name() string {
	return due.entry.Name()
}

func (due *downloadUrlEntry) Count(generator func() Counter) *Either {
	return CountDefault(due, generator())
}

func (due *downloadUrlEntry) Open() (io.ReadCloser, error) {
	reader, err := due.entry.Open()
	if err != nil {
		return nil, err
	}
	return createTeeReader(reader, due.Name())
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

type urlEntry struct {
	url   string
	index int
}

func (ue *urlEntry) Index() int {
	return ue.index
}

func (ue *urlEntry) Name() string {
	return ue.url
}

func (ue *urlEntry) Count(generator func() Counter) *Either {
	return CountDefault(ue, generator())
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

func toURLEntry(entry Entry, opts *ReadOptions) Entry {
	newEntry := &urlEntry{url: entry.Name(), index: entry.Index()}
	if opts.StoreContent {
		return &downloadUrlEntry{entry: newEntry}
	}
	return newEntry
}
