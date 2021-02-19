package wildcat

import "testing"

func TestIsUrl(t *testing.T) {
	testdata := []struct {
		giveString string
		wontUrl    bool
	}{
		{"http://github.com/tamada", true},
		{"github.com/tamada", false},
		{"/hoge/path", false},
		{"https://github.com/tamada", true},
	}
	for _, td := range testdata {
		gotFlag := IsUrl(td.giveString)
		if gotFlag != td.wontUrl {
			t.Errorf(`IsUrl("%s") did not match, wont %v, got %v`, td.giveString, td.wontUrl, gotFlag)
		}
	}
}

func TestExistFile(t *testing.T) {
	testdata := []struct {
		givePath string
		wontFlag bool
	}{
		{"testdata/wc/humpty_dumpty.txt", true},
		{"testdata/wc/london_bridge_is_broken_down.txt", true},
		{"testdata/wc/ja/sakura_sakura.txt", true},
		{"testdata", false},
		{"no_file_or_directory", false},
	}

	for _, td := range testdata {
		gotFlag := ExistFile(td.givePath)
		if gotFlag != td.wontFlag {
			t.Errorf("ExistFile(%s) did not match, wont %v, got %v", td.givePath, td.wontFlag, gotFlag)
		}
	}
}

func TestExistDir(t *testing.T) {
	testdata := []struct {
		givePath string
		wontFlag bool
	}{
		{"testdata/wc/humpty_dumpty.txt", false},
		{"testdata/wc/london_bridge_is_broken_down.txt", false},
		{"testdata/wc/ja/sakura_sakura.txt", false},
		{"testdata", true},
		{"no_file_or_directory", false},
	}

	for _, td := range testdata {
		gotFlag := ExistDir(td.givePath)
		if gotFlag != td.wontFlag {
			t.Errorf("ExistFile(%s) did not match, wont %v, got %v", td.givePath, td.wontFlag, gotFlag)
		}
	}
}
