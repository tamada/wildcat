package wildcat

import (
	"testing"
)

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
		{"testdata/archives/wc.war", 4, 78, 312, 1781},
	}

	for _, td := range testdata {
		argf := NewArgf([]string{td.giveFileName}, &ReadOptions{FileList: false, NoIgnore: true, NoExtract: false})
		targets, _ := argf.CollectTargets()
		rs, _ := targets.CountAll(func() Counter { return NewCounter(All) })
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
