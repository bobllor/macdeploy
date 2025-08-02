package core

import (
	"fmt"
	"macos-deployment/deploy_files/logger"
	"macos-deployment/deploy_files/utils"
	"os/exec"
	"strings"
)

// EnableFileVault enables FileVault and returns the key generated from the command.
// If it fails then an empty string is returned.
func EnableFileVault() string {
	cmd := "fdesetup isactive"
	out, _ := exec.Command("bash", "-c", cmd).Output()
	fileVaultStatus := strings.ToLower(string(out))

	fileVaultMsg := fmt.Sprintf("FileVault status: %s", fileVaultStatus)
	logger.Log(fileVaultMsg, 6)

	if strings.Contains(fileVaultStatus, "false") {
		scriptName := "enable_filevault.sh"

		logger.Log("Enabling FileVault", 6)

		scriptPath := fmt.Sprintf("%s/%s/%s", utils.Home, "macos-deployment/deploy_files", scriptName)
		out, err := exec.Command("bash", scriptPath).CombinedOutput()

		outText := string(out)
		logMsg := fmt.Sprintf("Path: %s, Output: %s", scriptPath, outText)
		logger.Log(logMsg, 7)

		if err != nil {
			println(err)
			return ""
		}

		// output is <name> = '<key>'
		outArr := strings.Split(outText, "'")
		// TIL: in Go an empty string is added to the array if the delimiter is at the end!
		// also println is not the same as fmt.Println...
		key := outArr[1]

		logger.Log("FileVault enabled", 6)
		keyMsg := fmt.Sprintf("Generated FileVault key %s", key)
		logger.Log(keyMsg, 6)

		return key
	} else {
		logger.Log("Skipping FileVault process: Already enabled", 6)
	}

	return ""
}
