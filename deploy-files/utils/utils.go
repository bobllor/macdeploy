package utils

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// FormatFullName returns a formatted name: lowercase and replacement of spaces with periods.
// It will remove all invalid characters.
//
// This follows the same rule of macOS' naming convention.
func FormatFullName(value string) string {
	newName := strings.ToLower(value)
	newName = strings.TrimSpace(newName)

	var newNameBytes []rune
	invalidCharacters := map[string]struct{}{
		"/":  {},
		";":  {},
		",":  {},
		"\\": {},
		"=":  {},
		"%":  {},
		"\n": {},
	}

	spaceFound := true

	for _, strBytes := range newName {
		char := string(strBytes)

		if _, ok := invalidCharacters[char]; ok {
			continue
		}

		// multiple spaces, keep the first occurrence and skip the rest
		// boundary check is not needed because newName is already trimmed
		if char == " " {
			if !spaceFound {
				spaceFound = true
			} else {
				continue
			}
		} else if spaceFound {
			spaceFound = false
		}

		newNameBytes = append(newNameBytes, strBytes)
	}

	return strings.ReplaceAll(string(newNameBytes), " ", ".")
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

// RemoveFile removes a given file.
// If it is a directory, it will recursively remove the files.
func RemoveFile(filePath string) {

}
