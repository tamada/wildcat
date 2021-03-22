package wildcat

import (
	"testing"
)

func TestBuildTargets(t *testing.T) {
	testdata := []struct {
		givePaths      []string
		giveOpts       *ReadOptions // filelist, noignore, noextract, storecontent
		wontEntryCount int
		wontErrorCount int
	}{
		{[]string{"testdata/wc"}, &ReadOptions{FileList: false, NoIgnore: false}, 3, 0},
		{[]string{"testdata/ignores"}, &ReadOptions{FileList: false, NoIgnore: false}, 2, 0},
		{[]string{"testdata/ignores"}, &ReadOptions{FileList: false, NoIgnore: true}, 7, 0},
		{[]string{"testdata/filelist.txt"}, &ReadOptions{FileList: false, NoIgnore: true}, 1, 0},
		{[]string{"testdata/filelist.txt"}, &ReadOptions{FileList: true, NoIgnore: true}, 4, 0},
		{[]string{}, &ReadOptions{FileList: false, NoIgnore: false}, 1, 0},
		{[]string{"testdata/notfound"}, &ReadOptions{FileList: false, NoIgnore: false}, 0, 1},
	}

	for _, td := range testdata {
		argf := NewArgf(td.givePaths, td.giveOpts)
		targets, ec := argf.CollectTargets()
		if len(targets.entries) != td.wontEntryCount {
			t.Errorf("entry size of target did not match, wont %d, got %d", td.wontEntryCount, len(targets.entries))
		}
		if ec.Size() != td.wontErrorCount {
			t.Errorf("error size did not match, wont %d, got %d", td.wontErrorCount, ec.Size())
		}
		order := NewOrderWithIndex(-1)
		for _, entry := range targets.entries {
			// targets.entries are arranged in order, therefore, comparing before and current index should be 1.
			if order.Compare(entry.Index()) >= 0 {
				t.Errorf("%s: index did not match, wont %s, got %s", entry.Name(), order.String(), entry.Index().String())
			}
			order = entry.Index()
		}
	}
}
