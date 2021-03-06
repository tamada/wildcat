package wildcat

import (
	"net/url"
	"os"
	"strings"
)

func NormalizePath(arg NameAndIndex) NameAndIndex {
	path := arg.Name()
	if strings.HasSuffix(path, "\"") && strings.Index(path, "\"") == len(path)-1 {
		newPath := strings.TrimRight(path, "\"")
		if ExistDir(newPath) {
			return NewArgWithIndex(arg.Index(), newPath)
		}
	}
	return arg
}

// ExistFile examines the given path is the regular file.
// If given path is not found or is not a file, this function returns false.
func ExistFile(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.Mode().IsRegular()
}

// ExistDir examines the given path is the directory.
// If given path is not found or is not a directory, this function returns false.
func ExistDir(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

// IsURL checks the given path is the form of url.
func IsURL(path string) bool {
	u, err := url.Parse(path)
	if err != nil {
		return false
	}
	return u.Host != "" && u.Scheme != ""
}
