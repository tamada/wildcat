package wildcat

import ignore "github.com/sabhiram/go-gitignore"

type Ignore interface {
	IgnoreFile(path string) bool
}

type noIgnore struct {
}

func (ni *noIgnore) IgnoreFile(path string) bool {
	return false
}

type gitIgnore struct {
	ignore *ignore.GitIgnore
}
