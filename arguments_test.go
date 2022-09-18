package wildcat

import (
	"os"
	"strings"
	"testing"

	"github.com/tamada/wildcat/errors"
)

func toStr(list []NameAndIndex) []string {
	result := []string{}
	for _, item := range list {
		result = append(result, item.Name())
	}
	return result
}

func match(list []NameAndIndex, wonts []string) bool {
	for _, wont := range wonts {
		found := false
		for _, item := range list {
			if strings.Contains(item.Name(), wont) {
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
		{"testdata/filelist.txt", &ReadOptions{FileList: true, NoIgnore: false, NoExtract: false}, 4, []string{"humpty_dumpty.txt", "sakura_sakura.txt", "london_bridge_is_broken_down.txt", "https://www.apache.org/licenses/LICENSE-2.0.txt"}, 0},
	}
	runtimeOpts := &RuntimeOptions{ShowProgress: false, ThreadNumber: 10, StoreContent: false}
	for _, td := range testdata {
		file, _ := os.Open(td.stdinFile)
		origStdin := os.Stdin
		os.Stdin = file
		defer func() {
			os.Stdin = origStdin
			file.Close()
		}()
		argf := NewArgfWithOptions([]string{}, td.opts, runtimeOpts)
		ec := errors.New()
		wc := newWildcatImpl(td.opts, runtimeOpts, DefaultGenerator)
		rs, err := wc.CountAll(argf)
		ec.Push(err)

		if len(rs.list) != td.listSize {
			t.Errorf("%v: ResultSet size did not match, wont %d, got %d (%v)", td.stdinFile, td.listSize, len(rs.list), rs.list)
		}
		if !match(rs.list, td.wontFileNames) {
			t.Errorf("%v: ResultSet files did not match, wont %v, got %v", td.stdinFile, td.wontFileNames, rs.list)
		}
		if ec.Size() != td.wontErrorSize {
			t.Errorf("%v: ErrorSize did not match, wont %d, got %d (%v)", td.stdinFile, td.wontErrorSize, ec.Size(), ec.Error())
		}
	}
}

func TestCountAll(t *testing.T) {
	testdata := []struct {
		args          []string
		opts          *ReadOptions // FileList, NoIgnore, NoExtract, AllFiles
		listSize      int
		wontFileNames []string
		wontErrorSize int
	}{
		{[]string{"testdata/wc"}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false}, 3, []string{"humpty_dumpty.txt", "sakura_sakura.txt", "london_bridge_is_broken_down.txt"}, 0},
		{[]string{"https://www.apache.org/licenses/LICENSE-2.0.txt"}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false}, 1, []string{"https://www.apache.org/licenses/LICENSE-2.0.txt"}, 0},
		{[]string{"testdata/ignores"}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false, AllFiles: true}, 4, []string{"notIgnore.txt", "notIgnore_sub.txt", ".gitignore"}, 0},
		{[]string{"testdata/ignores"}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false, AllFiles: false}, 2, []string{"notIgnore.txt", "notIgnore_sub.txt"}, 0},
		{[]string{"testdata/ignores"}, &ReadOptions{FileList: false, NoIgnore: true, NoExtract: false, AllFiles: false}, 5, []string{"ignore.test", "ignore.test2", "notIgnore.txt", "notIgnore_sub.txt", "ignore_sub.test"}, 0},
		{[]string{"testdata/filelist.txt"}, &ReadOptions{FileList: true, NoIgnore: false, NoExtract: false}, 4, []string{"humpty_dumpty.txt", "sakura_sakura.txt", "london_bridge_is_broken_down.txt", "https://www.apache.org/licenses/LICENSE-2.0.txt"}, 0},
		{[]string{"testdata/not_found.txt"}, &ReadOptions{FileList: true, NoIgnore: false, NoExtract: false}, 0, []string{}, 1},
		{[]string{"https://example.com/not_found"}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false}, 0, []string{}, 1},
		{[]string{"https://github.com/tamada/wildcat/raw/main/testdata/archives/wc.jar"}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false}, 4, []string{"humpty_dumpty.txt", "sakura_sakura.txt", "london_bridge_is_broken_down.txt"}, 0},
	}
	runtimeOpts := &RuntimeOptions{ShowProgress: false, ThreadNumber: 10, StoreContent: false}
	for _, td := range testdata {
		argf := NewArgfWithOptions(td.args, td.opts, runtimeOpts)
		wc := newWildcatImpl(td.opts, runtimeOpts, DefaultGenerator)
		rs, ec := wc.CountAll(argf)

		if len(rs.list) != td.listSize {
			t.Errorf("%v: ResultSet size did not match, wont %d, got %d (%v)", td.args, td.listSize, len(rs.list), toStr(rs.list))
		}
		if !match(rs.list, td.wontFileNames) {
			t.Errorf("%v: ResultSet files did not match, wont %v, got %v", td.args, td.wontFileNames, toStr(rs.list))
		}
		if ec.Size() != td.wontErrorSize {
			t.Errorf("%v: ErrorSize did not match, wont %d, got %d (%v)", td.args, td.wontErrorSize, ec.Size(), ec.Error())
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
	runtimeOpts := &RuntimeOptions{ShowProgress: false, ThreadNumber: 10, StoreContent: true}
	for _, td := range testdata {
		argf := NewArgfWithOptions([]string{td.url}, &ReadOptions{FileList: false, NoIgnore: false, NoExtract: false}, runtimeOpts)
		wc := NewWildcat(argf, DefaultGenerator)
		_, err := wc.CountAll(argf)
		if !err.IsEmpty() {
			t.Errorf("some error: %s", err.Error())
		}

		stat, err2 := os.Stat(td.wontFileName)
		if err2 != nil {
			t.Errorf("%s: file not found", td.wontFileName)
		}
		if stat != nil && !stat.Mode().IsRegular() {
			t.Errorf("%s: not regular file", td.wontFileName)
		}
		defer os.Remove(td.wontFileName)
	}
}
