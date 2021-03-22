package wildcat

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
)

// Entry shows the input for the each line in the results.
type Entry interface {
	NameAndIndex
	Count(generator Generator) *Either
	Open() (io.ReadCloser, error)
}

// NameAndIndex means that the implemented object has the name and index.
type NameAndIndex interface {
	Name() string
	Index() *Order
	Reindex(newIndex int)
}

type CompressedEntry struct {
	entry  Entry
	kind   string
	reader io.ReadSeekCloser
}

func (ce *CompressedEntry) Name() string {
	return ce.entry.Name()
}

func (ce *CompressedEntry) Index() *Order {
	return ce.entry.Index()
}

func (ce *CompressedEntry) Reindex(newIndex int) {
	ce.entry.Reindex(newIndex)
}

func (ce *CompressedEntry) Open() (io.ReadCloser, error) {
	if ce.reader != nil {
		ce.reader.Seek(0, 0)
		return ce.reader, nil
	}
	reader, err := ce.openImpl()
	if err == nil {
		ce.reader = NewReadSeekCloser(reader)
	}
	return ce.reader, err
}

func (ce *CompressedEntry) Count(generator Generator) *Either {
	return CountDefault(ce, generator())
}

func (ce *CompressedEntry) openImpl() (io.ReadCloser, error) {
	reader, err := ce.entry.Open()
	if err != nil {
		return nil, err
	}
	return wrapReader(reader, ce.kind), nil
}

type ArchiveEntry struct {
	entry Entry
}

func (ae *ArchiveEntry) Name() string {
	return ae.entry.Name()
}

func (ae *ArchiveEntry) Index() *Order {
	return ae.entry.Index()
}

func (ae *ArchiveEntry) Reindex(newIndex int) {
	ae.entry.Reindex(newIndex)
}

func (ae *ArchiveEntry) Open() (io.ReadCloser, error) {
	return ae.entry.Open()
}

func (ae *ArchiveEntry) Count(generator Generator) *Either {
	return &Either{Err: fmt.Errorf("not implement yet.")}
}

type FileEntry struct {
	nai    NameAndIndex
	reader io.ReadSeekCloser
}

func NewFileEntry(fileName string) *FileEntry {
	return &FileEntry{nai: NewArg(fileName)}
}

func NewFileEntryWithIndex(fileName string, index int) *FileEntry {
	return &FileEntry{nai: NewArgWithIndex(NewOrderWithIndex(index), fileName)}
}

func (fe *FileEntry) Name() string {
	return fe.nai.Name()
}

func (fe *FileEntry) Index() *Order {
	return fe.nai.Index()
}

func (fe *FileEntry) Reindex(newIndex int) {
	fe.nai.Reindex(newIndex)
}

func (fe *FileEntry) Open() (io.ReadCloser, error) {
	if fe.reader != nil {
		fe.reader.Seek(0, 0)
		return fe.reader, nil
	}
	reader, err := os.Open(fe.Name())
	if err != nil {
		return nil, err
	}
	fe.reader = NewReadSeekCloser(reader)
	return fe.reader, nil
}

func (fe *FileEntry) Count(generator Generator) *Either {
	return CountDefault(fe, generator())
}

type URLEntry struct {
	nai    NameAndIndex
	reader io.ReadSeekCloser
}

func (ue *URLEntry) Name() string {
	return ue.nai.Name()
}

func (ue *URLEntry) Index() *Order {
	return ue.nai.Index()
}

func (ue *URLEntry) Reindex(newIndex int) {
	ue.nai.Reindex(newIndex)
}

func (ue *URLEntry) Open() (io.ReadCloser, error) {
	if ue.reader != nil {
		ue.reader.Seek(0, 0)
		return ue.reader, nil
	}
	return ue.openImpl()
}

func (ue *URLEntry) openImpl() (io.ReadCloser, error) {
	response, err := http.Get(ue.Name())
	if err != nil {
		return nil, fmt.Errorf("%s: http error: %w", ue.Name(), err)
	}
	if response.StatusCode == 404 {
		defer response.Body.Close()
		return nil, fmt.Errorf("%s: file not found", ue.Name())
	}
	ue.reader = NewReadSeekCloser(response.Body)
	return ue.reader, nil
}

func (ue *URLEntry) Count(generator Generator) *Either {
	return CountDefault(ue, generator())
}

type stdinEntry struct {
	index *Order
}

// CountDefault is the default routine for counting.
func CountDefault(entry Entry, counter Counter) *Either {
	reader, err := entry.Open()
	if err != nil {
		return &Either{Err: err}
	}
	defer reader.Close()
	drainDataFromReader(reader, counter)
	return &Either{Results: []*Result{newResult(entry, counter)}}
}

func (se *stdinEntry) Name() string {
	return "<stdin>"
}

func (se *stdinEntry) Count(generator Generator) *Either {
	return CountDefault(se, generator())
}

func (se *stdinEntry) Open() (io.ReadCloser, error) {
	return os.Stdin, nil
}

func (se *stdinEntry) Index() *Order {
	return se.index
}

func (se *stdinEntry) Reindex(newIndex int) {
	// se.index = newIndex
}

type downloadURLEntry struct {
	entry *URLEntry
}

func (due *downloadURLEntry) Index() *Order {
	return due.entry.Index()
}

func (due *downloadURLEntry) Reindex(newIndex int) {
	due.entry.Reindex(newIndex)
}

func (due *downloadURLEntry) Name() string {
	return due.entry.Name()
}

func (due *downloadURLEntry) Count(generator Generator) *Either {
	return CountDefault(due, generator())
}

func (due *downloadURLEntry) Open() (io.ReadCloser, error) {
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

func toURLEntry(arg *Arg, opts *ReadOptions) Entry {
	newEntry := &URLEntry{nai: arg}
	if opts.StoreContent {
		return &downloadURLEntry{entry: newEntry}
	}
	return newEntry
}
