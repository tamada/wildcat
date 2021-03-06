package wildcat

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	"github.com/tamada/wildcat/iowrapper"
)

func ConvertToArchiveEntry(entry Entry) (Entry, bool) {
	reader, err := entry.Open()
	if err != nil {
		return entry, false
	}
	gotKind, _ := reader.ParseFileType()
	ext := gotKind.Extension
	return createArchiveEntry(entry, ext)
}

func createArchiveEntry(entry Entry, ext string) (Entry, bool) {
	switch ext {
	case "gz", "bz2":
		return wrapReaderAndTryAgain(entry, ext)
	case "jar", "zip":
		return &ZipEntry{entry: entry}, true
	case "tar":
		return &TarEntry{entry: entry}, true
	default:
		return entry, false
	}
}

func wrapReaderAndTryAgain(entry Entry, gotKind string) (Entry, bool) {
	newEntry := &CompressedEntry{entry: entry}
	return ConvertToArchiveEntry(newEntry)
}

func hasSuffix(fileName string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(fileName, suffix) {
			return true
		}
	}
	return false
}

type myReadCloser struct {
	reader io.Reader
	closer io.Closer
}

func (mrc *myReadCloser) Read(p []byte) (int, error) {
	return mrc.reader.Read(p)
}

func (mrc *myReadCloser) Close() error {
	return mrc.closer.Close()
}

func wrapReader(reader iowrapper.ReadCloseTypeParser) io.ReadCloser {
	ft, err := reader.ParseFileType()
	if err != nil {
		return reader
	}
	if hasSuffix(ft.Extension, "bz2") {
		return &myReadCloser{reader: bzip2.NewReader(reader), closer: reader}
	}
	if hasSuffix(ft.Extension, "gz") {
		r, _ := gzip.NewReader(reader)
		return r
	}
	return reader
}

type archiveItem interface {
	NameAndIndex
	Count(counter Counter) error
}

type tarItem struct {
	nameIndex NameAndIndex
	tar       *tar.Reader
}

type TarEntry struct {
	entry Entry
}

func (te *TarEntry) Name() string {
	return te.entry.Name()
}

func (te *TarEntry) Index() *Order {
	return te.entry.Index()
}

func (te *TarEntry) Open() (iowrapper.ReadCloseTypeParser, error) {
	return te.entry.Open()
}

func (te *TarEntry) Count(generator Generator) *Either {
	reader, err := te.Open()
	if err != nil {
		return &Either{Err: err}
	}
	return countTarEntries(te, generator, tar.NewReader(reader))
}

func countTarEntries(entry Entry, generator Generator, tar *tar.Reader) *Either {
	results := []*Result{}
	index := entry.Index().Sub()
	for {
		header, err := tar.Next()
		if err == io.EOF {
			break
		}
		name := fmt.Sprintf("%s!%s", entry.Name(), header.Name)
		result, err := countArchiveItem(generator(), &tarItem{tar: tar, nameIndex: NewArgWithIndex(index, name)})
		if err != nil {
			return &Either{Err: err}
		}
		results = append(results, result)
		index = index.Next()
	}
	return &Either{Results: results}
}

func countArchiveItem(counter Counter, item archiveItem) (*Result, error) {
	item.Count(counter)
	return &Result{nameIndex: item, counter: counter}, nil
}

func (tf *tarItem) Count(counter Counter) error {
	return drainDataFromReader(tf.tar, counter)
}

func (tf *tarItem) Name() string {
	return tf.nameIndex.Name()
}

func (tf *tarItem) Index() *Order {
	return tf.nameIndex.Index()
}

func copyDataFromSource(in io.Reader) (io.ReaderAt, int64, error) {
	buff := bytes.NewBuffer([]byte{})
	size, err := io.Copy(buff, in)
	if err != nil {
		return nil, 0, err
	}
	return bytes.NewReader(buff.Bytes()), size, nil
}

func createZipReader(in io.Reader) (*zip.Reader, error) {
	reader, size, err := copyDataFromSource(in)
	if err != nil {
		return nil, fmt.Errorf("failed to read all zip data from Reader: %w", err)
	}
	return zip.NewReader(reader, size)
}

type zipItem struct {
	nameIndex NameAndIndex
	file      *zip.File
}

func (zf *zipItem) Index() *Order {
	return zf.nameIndex.Index()
}

func (zf *zipItem) Name() string {
	return zf.nameIndex.Name() + "!" + zf.file.Name
}

func (zf *zipItem) Count(counter Counter) error {
	reader, err := zf.file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()
	return drainDataFromReader(reader, counter)
}

type ZipEntry struct {
	entry Entry
}

func (ze *ZipEntry) Index() *Order {
	return ze.entry.Index()
}

func (ze *ZipEntry) Name() string {
	return ze.entry.Name()
}

func (ze *ZipEntry) Open() (iowrapper.ReadCloseTypeParser, error) {
	return ze.entry.Open()
}

func (ze *ZipEntry) Count(generator Generator) *Either {
	in, err := ze.Open()
	if err != nil {
		return &Either{Err: err}
	}
	rr, err := createZipReader(in)
	if err != nil {
		return &Either{Err: err}
	}
	return countZipEntries(ze, rr, generator)
}

func countZipEntries(entry Entry, rr *zip.Reader, generator Generator) *Either {
	results := []*Result{}
	index := entry.Index().Sub()
	for _, f := range rr.File {
		r, err := countArchiveItem(generator(), &zipItem{file: f, nameIndex: NewArgWithIndex(index, entry.Name())})
		if err != nil {
			return &Either{Err: err}
		}
		results = append(results, r)
		index = index.Next()
	}
	return &Either{Results: results}
}
