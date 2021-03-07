package wildcat

import (
	"testing"

	"github.com/tamada/wildcat/errors"
)

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

func TestArchives(t *testing.T) {
	testdata := []struct {
		giveFileName   string
		wontSize       int
		wontTotalLines int64
		wontTotalWords int64
		wontTotalBytes int64
	}{
		{"testdata/archives/wc.jar", 4, 78, 312, 1781},
		{"testdata/archives/wc.tar", 4, 78, 312, 1781},
		{"testdata/archives/wc.tar.gz", 4, 78, 312, 1781},
		{"testdata/archives/wc.tar.bz2", 4, 78, 312, 1781},
		// not supported archive format. Therefore, read as binary file.
		{"testdata/archives/wc.war", 1, 5, 62, 1080},
	}

	for _, td := range testdata {
		ec := errors.New()
		argf := NewArgf([]string{td.giveFileName}, &ReadOptions{FileList: false, NoIgnore: true, NoExtract: false})
		rs, _ := argf.CountAll(func() Counter { return NewCounter(All) }, ec)
		if rs.Size() != td.wontSize {
			t.Errorf("archive (%s) size did not match, wont %d, got %d", td.giveFileName, td.wontSize, rs.Size())
		}
		if rs.total.Count(Lines) != td.wontTotalLines {
			t.Errorf("archive (%s) total lines did not match, wont %d, got %d", td.giveFileName, td.wontTotalLines, rs.total.Count(Lines))
		}
		if rs.total.Count(Words) != td.wontTotalWords {
			t.Errorf("archive (%s) total words did not match, wont %d, got %d", td.giveFileName, td.wontTotalWords, rs.total.Count(Words))
		}
		if rs.total.Count(Bytes) != td.wontTotalBytes {
			t.Errorf("archive (%s) total bytes did not match, wont %d, got %d", td.giveFileName, td.wontTotalBytes, rs.total.Count(Bytes))
		}
	}
}
