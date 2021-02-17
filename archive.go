package wildcat

import "strings"

// IsArchiveFile checks the given fileName shows archive file.
// This function examines by the suffix of the fileName.
func IsArchiveFile(fileName string) bool {
	suffixes := []string{".zip", ".tar", ".tar.gz", ".tar.bz2", ".jar"}
	for _, suffix := range suffixes {
		if strings.HasSuffix(fileName, suffix) {
			return true
		}
	}
	return false
}
