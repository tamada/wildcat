package wildcat

import "testing"

func TestExistFile(t *testing.T) {
	testdata := []struct {
		givePath string
		wontFlag bool
	}{
		{"testdata/humpty_dumpty.txt", true},
		{"testdata/london_bridge_is_broken_down.txt", true},
		{"testdata/ja/sakura_sakura.txt", true},
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
		{"testdata/humpty_dumpty.txt", false},
		{"testdata/london_bridge_is_broken_down.txt", false},
		{"testdata/ja/sakura_sakura.txt", false},
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
