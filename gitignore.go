package wildcat

import (
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

type Ignore interface {
	IsIgnore(path string) bool
	Filter(targets []string) []string
}

type noIgnore struct {
	parent Ignore
}

func (ni *noIgnore) Filter(slice []string) []string {
	return slice
}

func (ni *noIgnore) IsIgnore(path string) bool {
	if ni.parent != nil {
		return ni.parent.IsIgnore(path)
	}
	return false
}

type gitIgnore struct {
	ignore *ignore.GitIgnore
	parent Ignore
}

func (gi *gitIgnore) Filter(slice []string) []string {
	results := []string{}
	for _, item := range slice {
		if !gi.IsIgnore(item) && !strings.HasSuffix(item, "/.gitignore") {
			results = append(results, item)
		}
	}
	return results
}

func (gi *gitIgnore) IsIgnore(path string) bool {
	if !gi.ignore.MatchesPath(path) {
		if gi.parent != nil {
			return gi.parent.IsIgnore(path)
		}
		return false
	}
	return true
}

func newIgnoreWithParent(dirPath string, parent Ignore) Ignore {
	gitIgnoreFile := filepath.Join(dirPath, ".gitignore")
	if ExistFile(gitIgnoreFile) {
		return newGitIgnore(gitIgnoreFile, parent)
	}
	return &noIgnore{parent: parent}
}

func newIgnore(dirPath string) Ignore {
	return newIgnoreWithParent(dirPath, nil)
}

func newGitIgnore(gitIgnoreFilePath string, parent Ignore) Ignore {
	gi, err := ignore.CompileIgnoreFile(gitIgnoreFilePath)
	if err != nil {
		return &noIgnore{parent: parent}
	}
	return &gitIgnore{ignore: gi, parent: parent}
}
