package wildcat

import (
	"path/filepath"
	"testing"
)

func TestReadFilesInDir(t *testing.T) {
	ec := NewErrorCenter()
	targets := readFilesInDir("testdata/wc", ec, false, &noIgnore{})
	if len(targets) != 3 {
		t.Errorf("readFilesInDir(\"testdata/wc\") size did not match, wont %d, got %d", 3, len(targets))
	}
	testdata := []string{"ja/sakura_sakura.txt", "humpty_dumpty.txt", "london_bridge_is_broken_down.txt"}
	for _, td := range testdata {
		if !Contains(targets, filepath.Join("testdata/wc/", td)) {
			t.Errorf("readFilesInDir did not contains %s", td)
		}
	}
}

func TestReadFilesInDirWithIgnore(t *testing.T) {
	ec := NewErrorCenter()
	targets := readFilesInDir("testdata/ignores", ec, true, newGitIgnore("testdata/ignores/.gitignore", nil))
	if len(targets) != 2 {
		t.Errorf("readFilesInDir(\"testdata/ignores\") size did not match, wont %d, got %d", 2, len(targets))
	}
	testdata := []string{"notIgnore.txt", "subdir/notIgnore.txt"}
	for _, td := range testdata {
		if !Contains(targets, filepath.Join("testdata/ignores/", td)) {
			t.Errorf("readFilesInDir did not contains %s", td)
		}
	}
}

func TestReadFilesInDirWithoutIgnore(t *testing.T) {
	ec := NewErrorCenter()
	targets := readFilesInDir("testdata/ignores", ec, false, &noIgnore{})
	if len(targets) != 7 {
		t.Errorf("readFilesInDir(\"testdata/ignores\") size did not match, wont %d, got %d", 7, len(targets))
	}
	testdata := []string{"notIgnore.txt", "subdir/notIgnore.txt"}
	for _, td := range testdata {
		if !Contains(targets, filepath.Join("testdata/ignores/", td)) {
			t.Errorf("readFilesInDir did not contains %s", td)
		}
	}
}
