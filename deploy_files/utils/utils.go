package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func CheckError(err error) error {
	if err != nil {
		return err
	}

	return nil
}

// GetPathMap searches the contents of a directory and returns a map of the files.
// The keys in the map are all lowercase and the extension is removed.
func GetFileMap(dirPath string) map[string]bool {
	pathContent := make(map[string]bool)

	dirEntries, dirErr := os.ReadDir(dirPath)
	if dirErr != nil {
		// FIXME: add logging
		// this is a critical error, if ignored it will always download pkgs no matter what.
		panic(dirErr)
	}

	for _, file := range dirEntries {
		fileName := strings.ToLower(file.Name())
		ext := filepath.Ext(fileName)
		if ext != "" {
			fileName = fileName[0 : len(fileName)-len(ext)]
		}

		pathContent[fileName] = true
	}

	return pathContent
}
