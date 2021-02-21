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

type file interface {
	Name() string
	Count(counter Counter, sink *DataSink)
}

type archiveTraverser interface {
	traverseSource(s *Source, r *DataSink)
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

func (tt *tarTraverser) traverseSource(s *Source, r *DataSink) {
	in := wrapReader(s.in, s.name)
	traverseTarImpl(tar.NewReader(in), s.name, r)
}

func traverseTarImpl(tar *tar.Reader, fileName string, r *DataSink) {
	for {
		header, err := tar.Next()
		if err == io.EOF {
			break
		}
		name := fmt.Sprintf("%s!%s", fileName, header.Name)
		countEach(&tarFile{tar: tar, name: name}, r)
	}
}

type tarFile struct {
	tar  *tar.Reader
	name string
}

func (tf *tarFile) Count(counter Counter, sink *DataSink) {
	drainDataFromReader(tf.tar, counter)
}

func (tf *tarFile) Name() string {
	return tf.name
}

type zipTraverser struct {
}

type myReaderAt struct {
	r io.Reader
	n int64
}

func copyDataFromSource(s *Source) (io.ReaderAt, int64, error) {
	buff := bytes.NewBuffer([]byte{})
	size, err := io.Copy(buff, s.in)
	if err != nil {
		return nil, 0, err
	}
	return bytes.NewReader(buff.Bytes()), size, nil
}

func createZipReader(s *Source) (*zip.Reader, error) {
	reader, size, err := copyDataFromSource(s)
	if err != nil {
		return nil, err
	}
	return zip.NewReader(reader, size)
}

func (zt *zipTraverser) traverseSource(s *Source, r *DataSink) {
	rr, err := createZipReader(s)
	if err != nil {
		r.ec.Push(err)
		return
	}
	for _, f := range rr.File {
		countEach(&zipFile{zipFileName: s.name, file: f}, r)
	}
}

func countEach(f file, r *DataSink) {
	counter := r.gen()
	f.Count(counter, r)
	r.rs.Push(f.Name(), counter)
}

type zipFile struct {
	zipFileName string
	file        *zip.File
}

func (zf *zipFile) Name() string {
	return zf.zipFileName + "!" + zf.file.Name
}

func (zf *zipFile) Count(counter Counter, sink *DataSink) {
	reader, err := zf.file.Open()
	if err != nil {
		sink.ec.Push(err)
		return
	}
	defer reader.Close()
	drainDataFromReader(reader, counter)
}

func newArchiveTraverser(fileName string) archiveTraverser {
	if hasSuffix(fileName, ".jar", ".zip") {
		return &zipTraverser{}
	}
	if hasSuffix(fileName, ".tar", ".tar.gz", ".tar.bz2") {
		return &tarTraverser{}
	}
	return nil
}
