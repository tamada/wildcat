package wildcat

import "testing"

func opts(fileList, noIgnore, noExtract, storeContent bool) *ReadOptions {
	return &ReadOptions{FileList: fileList, NoIgnore: noIgnore, NoExtract: noExtract, StoreContent: storeContent}
}

func TestBasic(t *testing.T) {
	testdata := []struct {
		giveStrings    []string
		opts           *ReadOptions // filelist, noignore, noextract, storecontent
		wontResultSize int
		wontError      bool
	}{
		{[]string{"testdata/wc"}, opts(false, false, false, false), 3, false},
		{[]string{"testdata/filelist.txt"}, opts(true, false, false, false), 4, false},
		{[]string{"docs/public/images/demo.gif"}, opts(false, false, false, false), 1, false},
	}
	for _, td := range testdata {
		argf := NewArgf(td.giveStrings, nil)
		wildcat := NewWildcat(td.opts, DefaultGenerator)
		rs, err := wildcat.CountAll(argf)
		if td.wontError == (err == nil || err.IsEmpty()) {
			t.Errorf("%v: wont error %v, but got %v", td.giveStrings, td.wontError, !td.wontError)
		}
		if rs == nil {
			return
		}
		if rs.Size() != td.wontResultSize {
			t.Errorf("%v: result size did not match, wont %d, got %d", td.giveStrings, td.wontResultSize, rs.Size())
		}
	}
}
