package wildcat

import "testing"

func opts(fileList, noIgnore, noExtract, storeContent bool) *testOpts {
	return &testOpts{
		readOpts:    &ReadOptions{FileList: fileList, NoIgnore: noIgnore, NoExtract: noExtract},
		runtimeOpts: &RuntimeOptions{ShowProgress: false, StoreContent: storeContent, ThreadNumber: 10},
	}
}

type testOpts struct {
	readOpts    *ReadOptions // filelist, noignore, noextract, storecontent
	runtimeOpts *RuntimeOptions
}

func TestBasic(t *testing.T) {
	testdata := []struct {
		giveStrings    []string
		opts           *testOpts
		wontResultSize int
		wontError      bool
	}{
		{[]string{"testdata/wc"}, opts(false, false, false, false), 3, false},
		{[]string{"testdata/filelist.txt"}, opts(true, false, false, false), 4, false},
		{[]string{"docs/static/images/demo.gif"}, opts(false, false, false, false), 1, false},
	}
	for _, td := range testdata {
		argf := NewArgfWithOptions(td.giveStrings, td.opts.readOpts, td.opts.runtimeOpts)
		wildcat := newWildcatImpl(td.opts.readOpts, td.opts.runtimeOpts, DefaultGenerator)
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
