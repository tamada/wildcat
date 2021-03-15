package wildcat

import "testing"

func TestFileType(t *testing.T) {
	testdata := []struct {
		givePath string
		wontExt  string
	}{
		{"testdata/archives/wc.jar", "zip"},
		{"testdata/archives/wc.zip", "zip"},
		{"testdata/archives/wc.tar", "tar"},
		{"testdata/archives/wc.tar.gz", "gz"},
		{"testdata/archives/wc.tar.bz2", "bz2"},
		{"testdata/wc/humpty_dumpty.txt", "unknown"},
	}
	for _, td := range testdata {
		entry := &defaultEntry{fileName: td.givePath}
		target := NewTarget(entry)
		gotType, err := target.ParseType()
		if err != nil {
			t.Errorf("%s: parseType got error: %s", td.givePath, err.Error())
		}
		if gotType != td.wontExt {
			t.Errorf("%s: parsed file type did not match, wont %s, got %s", td.givePath, td.wontExt, gotType)
		}
	}
}
