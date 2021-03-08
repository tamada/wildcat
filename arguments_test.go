package wildcat

import (
	"os"
	"strings"
	"testing"

	"github.com/tamada/wildcat/errors"
)

func match(list []*indexString, wonts []string) bool {
	for _, wont := range wonts {
		found := false
		for _, item := range list {
			if strings.Contains(item.value, wont) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestStdin(t *testing.T) {
	testdata := []struct {
		stdinFile     string
		opts          *ReadOptions
		listSize      int
		wontFileNames []string
		wontErrorSize int
	}{
		{"testdata/wc/london_bridge_is_broken_down.txt", &ReadOptions{FileList: false, NoIgnore: true, NoExtract: true}, 1, []string{"<stdin>"}, 0},
		{"testdata/filelist.txt", &ReadOptions{FileList: true, NoIgnore: false, NoExtract: false}, 3, []string{"humpty_dumpty.txt", "sakura_sakura.txt", "london_bridge_is_broken_down.txt"}, 0},
	}
	for _, td := range testdata {
		file, _ := os.Open(td.stdinFile)
		origStdin := os.Stdin
		os.Stdin = file
		defer func() {
			os.Stdin = origStdin
			file.Close()
		}()
		argf := NewArgf([]string{}, td.opts)
		ec := errors.New()
		rs, _ := argf.CountAll(func() Counter { return NewCounter(All) }, ec)

		if len(rs.list) != td.listSize {
			t.Errorf("ResultSet size did not match, wont %d, got %d (%v)", td.listSize, len(rs.list), rs.list)
		}
		if !match(rs.list, td.wontFileNames) {
			t.Errorf("ResultSet files did not match, wont %v, got %v", td.wontFileNames, rs.list)
		}
		if ec.Size() != td.wontErrorSize {
			t.Errorf("ErrorSize did not match, wont %d, got %d (%v)", td.wontErrorSize, ec.Size(), ec.Error())
		}
	}
}

func TestCountAll(t *testing.T) {
	testdata := []struct {
		args          []string
		opts          *ReadOptions // FileList, NoIgnore, NoExtract
		listSize      int
		wontFileNames []string
		wontErrorSize int
	}{
		{[]string{"testdata/wc"}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false}, 3, []string{"humpty_dumpty.txt", "sakura_sakura.txt", "london_bridge_is_broken_down.txt"}, 0},
		{[]string{"https://www.apache.org/licenses/LICENSE-2.0.txt"}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false}, 1, []string{"https://www.apache.org/licenses/LICENSE-2.0.txt"}, 0},
		{[]string{"testdata/ignores"}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false}, 2, []string{"notIgnore.txt", "notIgnore_sub.txt"}, 0},
		{[]string{"testdata/ignores"}, &ReadOptions{FileList: false, NoIgnore: true, NoExtract: false}, 7, []string{"ignore.test", "ignore.test2", "notIgnore.txt", "notIgnore_sub.txt", "ignore_sub.test"}, 0},
		{[]string{"testdata/filelist.txt"}, &ReadOptions{FileList: true, NoIgnore: false, NoExtract: false}, 3, []string{"humpty_dumpty.txt", "sakura_sakura.txt", "london_bridge_is_broken_down.txt"}, 0},
		{[]string{"testdata/not_found.txt"}, &ReadOptions{FileList: true, NoIgnore: false, NoExtract: false}, 0, []string{}, 1},
		{[]string{"https://example.com/not_found"}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false}, 0, []string{}, 1},
		{[]string{"https://github.com/tamada/wildcat/raw/main/testdata/archives/wc.jar"}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false}, 4, []string{"humpty_dumpty.txt", "sakura_sakura.txt", "london_bridge_is_broken_down.txt"}, 0},
	}
	for _, td := range testdata {
		argf := NewArgf(td.args, td.opts)
		ec := errors.New()
		rs, _ := argf.CountAll(func() Counter { return NewCounter(All) }, ec)

		if len(rs.list) != td.listSize {
			t.Errorf("ResultSet size did not match, wont %d, got %d (%v)", td.listSize, len(rs.list), rs.list)
		}
		if !match(rs.list, td.wontFileNames) {
			t.Errorf("ResultSet files did not match, wont %v, got %v", td.wontFileNames, rs.list)
		}
		if ec.Size() != td.wontErrorSize {
			t.Errorf("ErrorSize did not match, wont %d, got %d (%v)", td.wontErrorSize, ec.Size(), ec.Error())
		}
	}
}

func TestStoreFile(t *testing.T) {
	testdata := []struct {
		url          string
		wontFileName string
	}{
		{"https://github.com/tamada/wildcat/raw/main/testdata/archives/wc.jar", "wc.jar"},
	}
	for _, td := range testdata {
		argf := NewArgf([]string{td.url}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false, StoreContent: true})
		ec := errors.New()
		argf.CountAll(func() Counter { return NewCounter(All) }, ec)

		stat, err := os.Stat(td.wontFileName)
		if err != nil {
			t.Errorf("%s: file not found", td.wontFileName)
		}
		if !stat.Mode().IsRegular() {
			t.Errorf("%s: not regular file", td.wontFileName)
		}
		defer os.Remove(td.wontFileName)
	}
}
