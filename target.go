package wildcat

import (
	"fmt"
	"io"

	"github.com/h2non/filetype"
	"gitlab.com/osaki-lab/iowrapper"
)

// Target shows a target for counting, typically a file.
type Target interface {
	Open() (io.ReadSeekCloser, error)
	Reset() bool
	ParseType() (string, error)
	Close() error
}

type myTarget struct {
	entry  Entry
	reader io.ReadSeekCloser
	kind   string
}

// NewTarget creates an instance of Target for the counting target.
func NewTarget(entry Entry) Target {
	return &myTarget{entry: entry}
}

func (target *myTarget) Close() error {
	if target.reader != nil {
		return target.reader.Close()
	}
	return fmt.Errorf("%s: not opened", target.entry.Name())
}

func (target *myTarget) Reset() bool {
	if target.reader == nil {
		return false
	}
	_, err := target.reader.Seek(0, 0)
	return err == nil
}

func (target *myTarget) Open() (io.ReadSeekCloser, error) {
	if target.reader != nil {
		target.Reset()
		return target.reader, nil
	}
	reader, err := target.entry.Open()
	if err != nil {
		return nil, err
	}
	target.reader = NewReadSeekCloser(reader)
	return target.reader, nil
}

func (target *myTarget) ParseType() (string, error) {
	if target.kind != "" {
		return target.kind, nil
	}
	kind, err := parseImpl(target)
	if err != nil {
		return kind, err
	}
	if kind == "gz" || kind == "bz2" {
		kind = wrapAndTryAgain(target, kind)
	}
	target.kind = kind
	return kind, err
}

func wrapAndTryAgain(target *myTarget, kind string) string {
	target.reader = NewReadSeekCloser(wrapReader(target.reader, kind))
	newKind, _ := target.ParseType()
	if newKind != "unknown" {
		return newKind + "." + kind
	}
	return kind
}

func parseImpl(target *myTarget) (string, error) {
	reader, err := target.Open()
	if err != nil {
		return "", err
	}
	kind, _ := filetype.MatchReader(reader)
	target.Reset()
	return kind.Extension, err
}

type myReader struct {
	reader io.ReadSeeker
	closer io.Closer
}

// NewReadSeekCloser creates an instance of ReadSeekCloser from ReadCloser.
func NewReadSeekCloser(reader io.ReadCloser) io.ReadSeekCloser {
	return &myReader{reader: iowrapper.NewSeeker(reader), closer: reader}
}

func (mr *myReader) Read(p []byte) (n int, err error) {
	return mr.reader.Read(p)
}

func (mr *myReader) Seek(offset int64, whence int) (int64, error) {
	return mr.reader.Seek(offset, whence)
}

func (mr *myReader) Close() error {
	return mr.closer.Close()
}
