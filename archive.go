package wildcat

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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
	Count(counter Counter)
}

type archiveTraverser interface {
	traverse(fileName string, r *result)
	traverseSource(s *source, r *result)
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

func (tt *tarTraverser) traverseSource(s *source, r *result) {
	in := wrapReader(s.in, s.name)
	traverseTarImpl(tar.NewReader(in), s.name, r)
}

func (tt *tarTraverser) traverse(fileName string, r *result) {
	reader, err := os.Open(fileName)
	if err != nil {
		r.ec.Push(err)
		return
	}
	defer reader.Close()
	in := wrapReader(reader, fileName)
	traverseTarImpl(tar.NewReader(in), fileName, r)
}

func traverseTarImpl(tar *tar.Reader, fileName string, r *result) {
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

func (tf *tarFile) Count(counter Counter) {
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

func newMyReaderAt(r io.Reader) io.ReaderAt {
	return &myReaderAt{r: r}
}

func (u *myReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off < u.n {
		return 0, errors.New("invalid offset")
	}
	diff := off - u.n
	written, err := io.CopyN(ioutil.Discard, u.r, diff)
	u.n += written
	if err != nil {
		return 0, err
	}

	n, err = u.r.Read(p)
	u.n += int64(n)
	return
}

func copyDataFromSource(s *source) (io.ReaderAt, int64, error) {
	buff := bytes.NewBuffer([]byte{})
	size, err := io.Copy(buff, s.in)
	if err != nil {
		return nil, 0, err
	}
	return bytes.NewReader(buff.Bytes()), size, nil
}

func createZipReader(s *source) (*zip.Reader, error) {
	reader, size, err := copyDataFromSource(s)
	if err != nil {
		return nil, err
	}
	return zip.NewReader(reader, size)
}

func (zt *zipTraverser) traverseSource(s *source, r *result) {
	rr, err := createZipReader(s)
	if err != nil {
		r.ec.Push(err)
		return
	}
	for _, f := range rr.File {
		countEach(&zipFile{zipFileName: s.name, file: f}, r)
	}
}

func (zt *zipTraverser) traverse(fileName string, r *result) {
	rr, err := zip.OpenReader(fileName)
	if err != nil {
		r.ec.Push(err)
		return
	}
	defer rr.Close()
	for _, f := range rr.File {
		countEach(&zipFile{zipFileName: fileName, file: f}, r)
	}
}

func countEach(f file, r *result) {
	counter := r.gen()
	f.Count(counter)
	r.rs.Push(f.Name(), counter)
}

type zipFile struct {
	zipFileName string
	file        *zip.File
}

func (zf *zipFile) Name() string {
	return zf.zipFileName + "!" + zf.file.Name
}

func (zf *zipFile) Count(counter Counter) {
	reader, err := zf.file.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
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
