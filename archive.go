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
)

// IsArchiveFile checks the given fileName shows archive file.
// This function examines by the suffix of the fileName.
func IsArchiveFile(fileName string) bool {
	return hasSuffix(fileName, ".zip", ".tar", ".tar.gz", ".tar.bz2", ".jar")
}

func hasSuffix(fileName string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(fileName, suffix) {
			return true
		}
	}
	return false
}

type tarTraverser struct {
}

func wrapReader(reader io.Reader, fileName string) io.Reader {
	if hasSuffix(fileName, ".tar.bz2") {
		return bzip2.NewReader(reader)
	}
	if hasSuffix(fileName, ".tar.gz") {
		r, _ := gzip.NewReader(reader)
		return r
	}
	return reader
}

type archiver interface {
	nameIndex
	traverse(generator func() Counter) *Either
}

type archiveItem interface {
	nameIndex
	Count(counter Counter) error
}

type tarArchiver struct {
	entry Entry
}

type tarItem struct {
	nameIndex nameIndex
	tar       *tar.Reader
}

func countArchiveItem(counter Counter, item archiveItem) (*Result, error) {
	item.Count(counter)
	return &Result{nameIndex: item, counter: counter}, nil
}

func (tt *tarArchiver) Name() string {
	return tt.entry.Name()
}

func (tt *tarArchiver) Index() int {
	return tt.entry.Index()
}

func (tt *tarArchiver) traverse(generator func() Counter) *Either {
	plainIn, err := tt.entry.Open()
	if err != nil {
		return &Either{Err: err}
	}
	in := wrapReader(plainIn, tt.entry.Name())
	return tt.traverseTarImpl(generator, tar.NewReader(in))
}

func (tt *tarArchiver) traverseTarImpl(generator func() Counter, tar *tar.Reader) *Either {
	results := []*Result{}
	for {
		header, err := tar.Next()
		if err == io.EOF {
			break
		}
		name := fmt.Sprintf("%s!%s", tt.entry.Name(), header.Name)
		result, err := countArchiveItem(generator(), &tarItem{tar: tar, nameIndex: &indexString{index: tt.entry.Index(), value: name}})
		if err != nil {
			return &Either{Err: err}
		}
		results = append(results, result)
	}
	return &Either{Results: results}
}

func (tf *tarItem) Count(counter Counter) error {
	return drainDataFromReader(tf.tar, counter)
}

func (tf *tarItem) Name() string {
	return tf.nameIndex.Name()
}

func (tf *tarItem) Index() int {
	return tf.nameIndex.Index()
}

type zipArchiver struct {
	entry Entry
}

func (zt *zipArchiver) Name() string {
	return zt.entry.Name()
}

func (zt *zipArchiver) Index() int {
	return zt.entry.Index()
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

func (zt *zipArchiver) traverse(generator func() Counter) *Either {
	in, err := zt.entry.Open()
	if err != nil {
		return &Either{Err: err}
	}
	rr, err := createZipReader(in)
	if err != nil {
		return &Either{Err: err}
	}
	results := []*Result{}
	for _, f := range rr.File {
		r, err := countArchiveItem(generator(), &zipItem{file: f, nameIndex: &indexString{index: zt.entry.Index(), value: zt.entry.Name()}})
		if err != nil {
			return &Either{Err: err}
		}
		results = append(results, r)
	}
	return &Either{Results: results}
}

type zipItem struct {
	nameIndex nameIndex
	file      *zip.File
}

func (zf *zipItem) Index() int {
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

func newArchiver(entry *archiveEntry) archiver {
	fileName := entry.Name()
	if hasSuffix(fileName, ".jar", ".zip") {
		return &zipArchiver{entry: entry}
	}
	if hasSuffix(fileName, ".tar", ".tar.gz", ".tar.bz2") {
		return &tarArchiver{entry: entry}
	}
	return nil
}
