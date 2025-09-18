package core

import (
	"fmt"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"macos-deployment/deploy-files/yaml"
	"os/exec"
	"strings"
)

// EnableFileVault enables FileVault and returns the key generated from the command.
// If it fails then an empty string is returned.
func EnableFileVault(adminUser string, adminPassword string) string {
	cmd := fmt.Sprintf("sudo -S fdesetup isactive <<< '%s'", adminPassword)
	// turns out if isactive == false the exit status is 1. ignoring the error here!
	out, _ := exec.Command("bash", "-c", cmd).Output()

	// instead of a boolean it must be in a string due to the subprocess.
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

// addSecureToken adds the user to the SecureToken list for FileVault.
// This is required on every new user.
//
// If successful then nil is returned, otherwise an error is thrown and manual interaction is needed.
func addSecureToken(username string, userPassword string, adminInfo yaml.User) error {
	// turns out i forgot secure token access... that was rough to find out in prod
	secureTokenCmd := fmt.Sprintf("sudo sysadminctl -secureTokenOn '%s' -password '%s' -adminUser '%s' -adminPassword '%s'",
		username, userPassword, adminInfo.User_Name, adminInfo.Password)

	_, err := exec.Command("bash", "-c", secureTokenCmd).Output()
	if err != nil {
		logger.Log(fmt.Sprintf("Error enabling token for user, manual interaction needed: %v", err), 3)
		return fmt.Errorf("failed to enable secure token for user %s: %v", username, err)
	} else {
		logger.Log(fmt.Sprintf("Secure token added for %s", username), 6)
	}
	return nil
}
