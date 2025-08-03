package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

// GetPathMap searches the contents of a directory and returns a map of the files.
// The keys in the map are all lowercase and the extension is removed.
func GetFileMap(dirPath string) (map[string]bool, error) {
	pathContent := make(map[string]bool)

	dirEntries, dirErr := os.ReadDir(dirPath)
	if dirErr != nil {
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

// FormatName formats a name string as First.Last or F.Last.
//
// This function expects the string to consist of only two names, and does not
// handle any suffxies, bad characters, or special characters outside of "." and " ".
func FormatName(name string) string {
	//caser := cases.Title(language.AmericanEnglish)
	//newStr := caser.String(str)
	newStr := strings.Trim(name, " ")
	strBytes := []byte(newStr)

	replaceMap := map[string]bool{".": true, " ": true}

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
	nameArrLen := len(strArr)

	firstName := strArr[0]
	lastName := strArr[nameArrLen-1]

	return fmt.Sprintf("%s.%s", firstName, lastName)
}

// GetSerialTag retrieves the serial tag for the device.
//
// This only works on macOS devices.
func GetSerialTag() (string, error) {
	cmd := "ioreg -l | grep IOPlatformSerialNumber"
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", errors.New("cannot find serial tag of the device")
	}

	serialTagArr := strings.Split(string(out), "\"")
	serialTag := serialTagArr[len(serialTagArr)-2]

	return serialTag, nil
}
