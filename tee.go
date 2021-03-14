package wildcat

import "io"

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
