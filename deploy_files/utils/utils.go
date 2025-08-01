package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func CheckError(err error) error {
	if err != nil {
		return err
	}

	return nil
}

// GetPathMap searches the contents of a directory and returns a map of the files.
// The keys in the map are all lowercase and the extension is removed.
func GetFileMap(dirPath string) (map[string]bool, error) {
	pathContent := make(map[string]bool)

	dirEntries, dirErr := os.ReadDir(dirPath)
	if dirErr != nil {
		// FIXME: add logging
		// this is a critical error, if ignored it will always download pkgs no matter what.
		return nil, dirErr
	}

	for _, file := range dirEntries {
		fileName := strings.ToLower(file.Name())
		ext := filepath.Ext(fileName)
		if ext != "" {
			fileName = fileName[0 : len(fileName)-len(ext)]
		}

		pathContent[fileName] = true
	}

	return pathContent, nil
}

// FormatName
func FormatName(name string) string {
	//caser := cases.Title(language.AmericanEnglish)
	//newStr := caser.String(str)
	newStr := strings.Trim(name, " ")
	strBytes := []byte(newStr)

	replaceMap := map[string]bool{".": true, " ": true, ",": true}

	var delimiterIndex int = 0

	for i := 1; i < len(strBytes); i++ {
		if _, found := replaceMap[string(name[i])]; found {
			// update the next value from the replaceMap characters
			if delimiterIndex == 0 {
				delimiterIndex = i
			}

			strBytes[i+1] = byte(unicode.ToUpper(rune(name[i+1])))
		}
	}

	strBytes[0] = byte(unicode.ToUpper(rune(name[0])))

	newStr = string(strBytes)
	strArr := strings.Split(newStr, string(name[delimiterIndex]))

	return fmt.Sprintf("%s.%s", strArr[0], strArr[len(strArr)-1])
}
