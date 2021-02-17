package wildcat

import "testing"

func TestIsArchiveFile(t *testing.T) {
	testdata := []struct {
		giveFileName  string
		wontIsArchive bool
	}{
		{"file.zip", true},
		{"file.txt", false},
		{"file.jar", true},
		{"file.tar", true},
		{"file.tar.gz", true},
		{"file.tar.bz2", true},
		{"file.war", false}, // not support
	}

	for _, td := range testdata {
		gotIsArchive := IsArchiveFile(td.giveFileName)
		if gotIsArchive != td.wontIsArchive {
			t.Errorf(`IsArchiveFile("%s") did not match, wont %v, got %v`, td.giveFileName, td.wontIsArchive, gotIsArchive)
		}
	}
}
