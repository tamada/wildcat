package wildcat

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/tamada/wildcat/iowrapper"
)

// Entry shows the input for the each line in the results.
type Entry interface {
	NameAndIndex
	Count(generator Generator) *Either
	Open() (iowrapper.ReadCloseTypeParser, error)
}

// NameAndIndex means that the implemented object has the name and index.
type NameAndIndex interface {
	Name() string
	Index() *Order
}

type CompressedEntry struct {
	entry  Entry
	reader iowrapper.ReadCloseTypeParser
}

func (ce *CompressedEntry) Name() string {
	return ce.entry.Name()
}

func (ce *CompressedEntry) Index() *Order {
	return ce.entry.Index()
}

func (ce *CompressedEntry) Open() (iowrapper.ReadCloseTypeParser, error) {
	if ce.reader != nil {
		return ce.reader, nil
	}
	reader, err := ce.openImpl()
	if err == nil {
		ce.reader = iowrapper.NewReader(reader)
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
	return wrapReader(iowrapper.NewReader(reader)), nil
}

type FileEntry struct {
	nai    NameAndIndex
	reader iowrapper.ReadCloseTypeParser
}

func NewFileEntry(fileName string) *FileEntry {
	return NewFileEntryWithIndex(NewArgWithIndex(NewOrder(), fileName))
}

func NewFileEntryWithIndex(nai NameAndIndex) *FileEntry {
	return &FileEntry{nai: nai}
}

func (fe *FileEntry) Name() string {
	return fe.nai.Name()
}

func (fe *FileEntry) Index() *Order {
	return fe.nai.Index()
}

func (fe *FileEntry) Open() (iowrapper.ReadCloseTypeParser, error) {
	if fe.reader != nil {
		return fe.reader, nil
	}
	reader, err := os.Open(fe.Name())
	if err != nil {
		return nil, err
	}
	fe.reader = iowrapper.NewReader(reader)
	return fe.reader, nil
}

func (fe *FileEntry) Count(generator Generator) *Either {
	return CountDefault(fe, generator())
}

type URLEntry struct {
	nai    NameAndIndex
	reader iowrapper.ReadCloseTypeParser
}

func (ue *URLEntry) Name() string {
	return ue.nai.Name()
}

func (ue *URLEntry) Index() *Order {
	return ue.nai.Index()
}

func (ue *URLEntry) Open() (iowrapper.ReadCloseTypeParser, error) {
	if ue.reader != nil {
		return ue.reader, nil
	}
	return ue.openImpl()
}

func (ue *URLEntry) openImpl() (iowrapper.ReadCloseTypeParser, error) {
	response, err := http.Get(ue.Name())
	if err != nil {
		return nil, fmt.Errorf("%s: http error: %w", ue.Name(), err)
	}
	if response.StatusCode == 404 {
		defer response.Body.Close()
		return nil, fmt.Errorf("%s: file not found", ue.Name())
	}
	ue.reader = iowrapper.NewReader(response.Body)
	return ue.reader, nil
}

func (ue *URLEntry) Count(generator Generator) *Either {
	return CountDefault(ue, generator())
}

type stdinEntry struct {
	index  *Order
	reader iowrapper.ReadCloseTypeParser
}

// CountDefault is the default routine for counting.
func CountDefault(entry Entry, counter Counter) *Either {
	reader, err := entry.Open()
	if err != nil {
		return &Either{Err: err}
	}
	defer reader.Close()
	if err := drainDataFromReader(reader, counter); err != nil {
		return &Either{Err: err}
	}
	return &Either{Results: []*Result{newResult(entry, counter)}}
}

func (se *stdinEntry) Name() string {
	return "<stdin>"
}

func (se *stdinEntry) Count(generator Generator) *Either {
	return CountDefault(se, generator())
}

func (se *stdinEntry) Open() (iowrapper.ReadCloseTypeParser, error) {
	if se.reader == nil {
		se.reader = iowrapper.NewReader(os.Stdin)
	}
	return se.reader, nil
}

func (se *stdinEntry) Index() *Order {
	return se.index
}

type downloadURLEntry struct {
	entry  *URLEntry
	reader iowrapper.ReadCloseTypeParser
}

func (due *downloadURLEntry) Index() *Order {
	return due.entry.Index()
}

func (due *downloadURLEntry) Name() string {
	return due.entry.Name()
}

func (due *downloadURLEntry) Count(generator Generator) *Either {
	return CountDefault(due, generator())
}

func (due *downloadURLEntry) Open() (iowrapper.ReadCloseTypeParser, error) {
	if due.reader != nil {
		return due.reader, nil
	}
	reader, err := due.entry.Open()
	if err != nil {
		return nil, err
	}
	in, err := createTeeReader(reader, due.Name())
	due.reader = in
	return in, err
}

func createTeeReader(reader io.ReadCloser, name string) (iowrapper.ReadCloseTypeParser, error) {
	u, err := url.Parse(name)
	if err != nil {
		return nil, fmt.Errorf("url.Parse failed: %w", err)
	}
	newName := path.Base(u.Path)
	writer, err := os.Create(newName)
	if err != nil {
		return nil, fmt.Errorf("%s: file creation error (%w)", newName, err)
	}
	return iowrapper.NewReader(iowrapper.NewTeeReader(reader, writer)), nil
}

func toURLEntry(arg NameAndIndex, opts *ReadOptions) Entry {
	newEntry := &URLEntry{nai: arg}
	if opts.StoreContent {
		return &downloadURLEntry{entry: newEntry}
	}
	return newEntry
}
