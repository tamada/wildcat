package wildcat

import "testing"

func TestGitIgnore(t *testing.T) {
	testdata := []struct {
		givePath  string
		giveSlice []string
		wontSlice []string
	}{
		{"testdata/ignores", []string{"ignore.test", "notIgnore.txt"}, []string{"notIgnore.txt"}},
		{"testdata/archives", []string{"ignore.test", "notIgnore.txt"}, []string{"notIgnore.txt", "ignore.test"}}, // no ignore
	}
	for _, td := range testdata {
		ig := newIgnore(td.givePath)
		gotSlice := ig.Filter(td.giveSlice)
		if len(gotSlice) != len(td.wontSlice) {
			t.Errorf("got slice length did not match: wont %d, got %d", len(td.wontSlice), len(gotSlice))
		}
		for _, gotItem := range gotSlice {
			found := false
			for _, wontItem := range td.wontSlice {
				if gotItem == wontItem {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("gotItem: %s not found in %v", gotItem, td.wontSlice)
				break
			}
		}
	}
}
