package iowrapper

import (
	"bytes"
	"io"

	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

const DefaultBufferSize = 368

type ReadCloseTypeParser interface {
	io.ReadCloser
	ParseFileType() (types.Type, error)
}

type readSeekCloser struct {
	reader  io.ReadCloser
	wrapper io.Reader
	buffer  []byte
}

func (rsc *readSeekCloser) ParseFileType() (types.Type, error) {
	return filetype.Match(rsc.buffer)
}

func (rsc *readSeekCloser) Close() error {
	return rsc.reader.Close()
}

func (rsc *readSeekCloser) Read(p []byte) (int, error) {
	return rsc.wrapper.Read(p)
}

func resizeBuffer(buffer []byte, length int, err error) []byte {
	if (err == nil || err == io.EOF) && length < len(buffer) {
		buf := make([]byte, length)
		copy(buf, buffer)
		return buf
	}
	return buffer
}

func NewReaderWithBufferSize(in io.ReadCloser, bufferSize int64) ReadCloseTypeParser {
	buffer := make([]byte, bufferSize)
	len, err := in.Read(buffer)
	buffer = resizeBuffer(buffer, len, err)
	return &readSeekCloser{
		reader:  in,
		buffer:  buffer,
		wrapper: io.MultiReader(bytes.NewReader(buffer), in),
	}
}

func NewReader(in io.ReadCloser) ReadCloseTypeParser {
	return NewReaderWithBufferSize(in, DefaultBufferSize)
}
