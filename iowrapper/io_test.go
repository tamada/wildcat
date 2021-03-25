package iowrapper

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestRead(t *testing.T) {
	testdata := []struct {
		givePath    string
		wontStrings []string
	}{
		{"../testdata/filelist.txt", []string{"testdata/wc/humpty_dumpty.txt", "testdata/wc/ja/sakura_sakura.txt", "testdata/wc/london_bridge_is_broken_down.txt", "https://www.apache.org/licenses/LICENSE-2.0.txt"}},
	}
	for _, td := range testdata {
		in, _ := os.Open(td.givePath)
		reader := NewReader(in)
		data, _ := io.ReadAll(reader)
		str := string(data)
		lines := strings.Split(str, "\n")
		for index, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && line != td.wontStrings[index] {
				t.Errorf("%s[%d]: did not match, wont %s, got %s", td.givePath, index, td.wontStrings[index], line)
			}
		}
	}
}

func TestBasic(t *testing.T) {
	testdata := []struct {
		givePath string
		wontType string
	}{
		{"../testdata/archives/wc.jar", "zip"},
	}
	for _, td := range testdata {
		in, _ := os.Open(td.givePath)
		reader := NewReader(in)
		defer reader.Close()

		ft, err := reader.ParseFileType()
		if err != nil {
			t.Errorf(err.Error())
			break
		}
		ext := ft.Extension
		if ext != td.wontType {
			t.Errorf("%s: parsed type did not match, wont %s, got %s", td.givePath, td.wontType, ext)
		}
	}
}
