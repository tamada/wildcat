package wildcat

import (
	"io"

	"gitlab.com/osaki-lab/iowrapper"
)

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
