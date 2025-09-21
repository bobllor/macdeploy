package utils

import (
	"errors"
	"fmt"
	"macos-deployment/deploy-files/logger"
	"os"
	"os/exec"
	"strings"
)

// InitializeSudo starts a sudo session without the need of manual input.
// This can be called multiple times to refresh the sudo timer.
func InitializeSudo(adminPassword string) error {
	initSudoCmd := fmt.Sprintf("sudo -S echo <<< '%s'", adminPassword)
	err := exec.Command("bash", "-c", initSudoCmd).Run()

	if err != nil {
		// not a major issue, just requires manual interaction instead.
		return err
	}

	return nil
}

// GetPathMap searches the contents of a directory and returns a map of the files.
// The keys in the map are all lowercase.
func GetSearchFiles(dirPath string) ([]string, error) {
	pathContent := make([]string, 0)

	dirEntries, dirErr := os.ReadDir(dirPath)
	if dirErr != nil {
		return nil, dirErr
	}

	for _, file := range dirEntries {
		fileName := strings.ToLower(file.Name())

		pathContent = append(pathContent, fileName)
	}

	return pathContent, nil
}

// FormatFullName returns a formatted name: lowercase and replacement of spaces with periods.
// It will remove all invalid characters.
//
// This follows the Apple's naming convention for macOS.
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
func RemoveFiles[T any](filesToRemove map[string]T, projectDirectory string) error {
	currDir, err := os.Getwd()
	// i am unsure what errors can happen here
	if err != nil {
		logger.Log(fmt.Sprintf("Error getting working directory: %v", err), 3)
		return err
	}

	if strings.Contains(currDir, projectDirectory) {
		errMsg := fmt.Sprintf("project directory is forbidden, clean up aborted %v", currDir)
		err := errors.New(errMsg)
		logger.Log(err.Error(), 3)

		return err
	}

	files, err := os.ReadDir(currDir)
	if err != nil {
		logger.Log(fmt.Sprintf("Unable to read directory %v", err), 4)
		return err
	}

	// yes this is only for logging.
	currFileNames := make([]string, 0)
	for _, file := range files {
		currFileNames = append(currFileNames, strings.ToLower(file.Name()))
	}
	logger.Log(fmt.Sprintf("Files in working directory: %v", currFileNames), 7)

	for _, file := range files {
		fileName := strings.ToLower(file.Name())

		if _, ok := filesToRemove[fileName]; ok {
			if file.IsDir() {
				// could be unneeded but want to be extra safe
				if Globals.DistDirName == fileName {
					err := os.RemoveAll(fileName)
					if err != nil {
						logger.Log(fmt.Sprintf("Error removing directory %v", err), 3)
						continue
					}
				}
			} else {
				err := os.Remove(fileName)
				if err != nil {
					logger.Log(fmt.Sprintf("Error removing file %v", err), 3)
					continue
				}
			}

			logger.Log(fmt.Sprintf("Removed file %s", fileName), 6)
		}
	}

	return nil
}
