package wildcat

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
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

type archiveTraverser interface {
	traverse(fileName string, rs *ResultSet, counterGenerator func() Counter)
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

func (tt *tarTraverser) traverse(fileName string, rs *ResultSet, generator func() Counter) {
	reader, err := os.Open(fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return
	}
	defer reader.Close()
	in := wrapReader(reader, fileName)
	tar := tar.NewReader(in)
	for {
		header, err := tar.Next()
		if err == io.EOF {
			break
		}
		name := fmt.Sprintf("%s!%s", fileName, header.Name)
		countEach(&tarFile{tar: tar, name: name}, rs, generator)
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

func (zt *zipTraverser) traverse(fileName string, rs *ResultSet, generator func() Counter) {
	r, err := zip.OpenReader(fileName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	defer r.Close()
	for _, f := range r.File {
		countEach(&zipFile{zipFileName: fileName, file: f}, rs, generator)
	}
}

func countEach(f file, rs *ResultSet, generator func() Counter) {
	counter := generator()
	f.Count(counter)
	rs.Push(f.Name(), counter)
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

func newArchiveTarget(fileName string, ec *ErrorCenter) Target {
	if hasSuffix(fileName, ".jar", ".zip") {
		return &archiveTarget{fileName: fileName, traverser: &zipTraverser{}}
	}
	if hasSuffix(fileName, ".tar", ".tar.gz", ".tar.bz2") {
		return &archiveTarget{fileName: fileName, traverser: &tarTraverser{}}
	}
	ec.Push(fmt.Errorf("%s: unsupported archive file", fileName)) // never reach here!
	return nil
}

type archiveTarget struct {
	fileName  string
	traverser archiveTraverser
}

func (at *archiveTarget) Count(counterGenerator func() Counter) *ResultSet {
	rs := NewResultSet()
	at.traverser.traverse(at.fileName, rs, counterGenerator)
	return rs
}
