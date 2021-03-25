package iowrapper

import (
	"io"
	"strings"
	"testing"
)

type nopWriteCloser struct {
	writer io.Writer
}

func (nwc *nopWriteCloser) Close() error {
	return nil
}

func (nwc *nopWriteCloser) Write(p []byte) (int, error) {
	return nwc.writer.Write(p)
}

func NopWriteCloser(writer io.Writer) io.WriteCloser {
	return &nopWriteCloser{writer: writer}
}

func TestTeeReader(t *testing.T) {
	content := "This is test content for TeeReader."
	writer := new(strings.Builder)
	tee := NewTeeReader(io.NopCloser(strings.NewReader(content)), NopWriteCloser(writer))
	defer tee.Close()
	data, err := io.ReadAll(tee)
	if err != nil {
		t.Errorf("TeeReader read error: %s", err.Error())
	}
	gotContent := string(data)
	if gotContent != content {
		t.Errorf("tee read data did not match, wont %s, got \"%s\"", content, gotContent)
	}
	teeString := writer.String()
	if teeString != content {
		t.Errorf("tee written data did not match, wont \"%s\", got \"%s\"", content, teeString)
	}
}
