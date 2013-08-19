package store

import (
	"os"
	"strings"
)

// ToBytes convert strings passed as argument to a slice of bytes
func ToBytes(args...string) [][]byte {
	res := make([][]byte, len(args))
	for i, v := range args {
		res[i] = []byte(v)
	}
	return res
}

func dirExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	} // if file doesn't exists, throws here

	return fileInfo.IsDir(), nil
}

func isFilePath(str string) bool {
	startsWithDot := strings.HasPrefix(str, ".")
	containsSlash := strings.Contains(str, "/")

	if startsWithDot == true || containsSlash == true {
		return true
	}

	return false
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
