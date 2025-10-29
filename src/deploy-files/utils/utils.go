package utils

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

// GetFiles reads the directory and returns a map of the files.
// The keys in the map are all in lowercase.
// This does not traverse recursively.
func GetFiles(dirPath string) ([]string, error) {
	pathContent := make([]string, 0)

	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, file := range dirEntries {
		fileName := strings.ToLower(file.Name())

		pathContent = append(pathContent, fileName)
	}

	return pathContent, nil
}

// FormatUsername returns a formatted username: lowercase and made into one word.
// It will remove all invalid characters.
//
// This follows the Apple's naming convention for MacBooks and is used for the internal
// username of the device for the specified user.
func FormatUsername(value string) string {
	newName := strings.ToLower(value)
	newName = strings.TrimSpace(newName)

	var newNameRunes []rune
	// allowed characters: -_.
	// however, the period (.) cannot be at the beginning of the name
	invalidString := "`~!@#$%^&*()=+[]{}\\|;:'\",<>/?\n\t"

	hasAlpha := false

	for _, charRune := range newName {
		charStr := string(charRune)
		if strings.ContainsRune(invalidString, charRune) || charStr == " " {
			continue
		}

		if !hasAlpha {
			if unicode.IsLetter(charRune) {
				hasAlpha = true
			}
		}

		newNameRunes = append(newNameRunes, charRune)
	}

	// if the string has no alphabets, then append a to the front and do not
	// trim the leading dots if they exist.
	if !hasAlpha {
		noAlphaRunes := []rune("a")
		noAlphaRunes = append(noAlphaRunes, newNameRunes...)

		return string(noAlphaRunes)
	} else {
		// removing all leading periods
		newName = strings.TrimLeft(string(newNameRunes), ".")
	}

	return newName
}

// GetSerialTag retrieves the serial tag for the device.
//
// An error will return if the serial tag cannot be retrieved.
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

// RemoveFiles removes the files based on a given map. It searches for the files in the map
// of the directory the execution process started in.
func RemoveFiles[T any](filesToRemove map[string]T, files []os.DirEntry) {
	for _, file := range files {
		fileName := strings.ToLower(file.Name())

		if _, ok := filesToRemove[fileName]; ok {
			if file.IsDir() {
				err := os.RemoveAll(fileName)
				if err != nil {
					fmt.Printf("Error removing directory %v\n", err)
					continue
				}
			} else {
				err := os.Remove(fileName)
				if err != nil {
					fmt.Printf("Error removing file %v\n", err)
					continue
				}
			}

			fmt.Printf("Removed file %s\n", fileName)
		}
	}
}

// FormatBannerString returns a formatted string between two lines of stars (*)
// with padding the inner text. It creates the banner text based on the slice of strings given.
func FormatBannerString(lineArr []string, padding int) string {
	longestLen := len(lineArr[0])
	for i := 1; i < len(lineArr); i++ {
		if len(lineArr[i]) > longestLen {
			longestLen = len(lineArr[i])
		}
	}
	// right padding
	longestLen += padding * 2
	starLineArr := []string{}
	for range longestLen {
		starLineArr = append(starLineArr, "*")
	}

	starLine := strings.Join(starLineArr, "")
	outString := []string{}
	endLineArr := []string{}
	for range padding / 2 {
		endLineArr = append(endLineArr, "\n")
		starLine += "\n"
	}

	outString = append(outString, starLine)

	for _, line := range lineArr {
		padArr := []string{}

		// left padding
		for range padding {
			padArr = append(padArr, " ")
		}

		padArr = append(padArr, line)

		outString = append(outString, strings.Join(padArr, ""))
	}

	endLine := strings.Join(endLineArr, "") + starLine
	outString = append(outString, endLine)

	return strings.Join(outString, "\n")
}
