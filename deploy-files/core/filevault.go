package core

import (
	"fmt"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"os/exec"
	"strings"
)

// EnableFileVault enables FileVault and returns the key generated from the command.
// If it fails then an empty string is returned.
func EnableFileVault(adminUser string, adminPassword string) string {
	cmd := "fdesetup isactive"
	out, _ := exec.Command("sudo", "bash", "-c", cmd).Output()
	fileVaultStatus := strings.TrimSpace(strings.ToLower(string(out)))

	// either some fail happened or this is ran on a non-mac OS
	if fileVaultStatus == "" {
		fileVaultStatus = "unknown"
	}

	fileVaultMsg := fmt.Sprintf("FileVault status: %s", fileVaultStatus)
	logger.Log(fileVaultMsg, 6)

	if fileVaultStatus == "false" {
		logger.Log("Enabling FileVault", 6)

		out, err := exec.Command("sudo", "bash", "-c", scripts.EnableFileVaultScript,
			adminUser, adminPassword).CombinedOutput()

		outText := string(out)
		logMsg := strings.TrimSpace(fmt.Sprintf("Output: %s", outText))
		logger.Log(logMsg, 7)

		if err != nil {
			logger.Log("Failed to execute FileVault script", 3)
			return ""
		}

		// output is <name> = '<key>'
		outArr := strings.Split(outText, "'")
		// TIL: in Go an empty string is added to the array if the delimiter is at the end!
		// also println is not the same as fmt.Println...
		key := outArr[1]

		logger.Log("FileVault enabled", 6)

		return key
	} else if fileVaultStatus == "true" {
		logger.Log("FileVault is already enabled", 6)
	} else {
		logger.Log("FileVault failed to execute", 3)
	}

	return ""
}
