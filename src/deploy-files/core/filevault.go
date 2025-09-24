package core

import (
	"fmt"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"macos-deployment/deploy-files/yaml"
	"os/exec"
	"strings"
)

type FileVault struct {
	admin  *yaml.UserInfo
	script *scripts.BashScripts
	log    *logger.Log
}

func NewFileVault(admin *yaml.UserInfo, script *scripts.BashScripts, log *logger.Log) *FileVault {
	fv := FileVault{
		admin:  admin,
		script: script,
		log:    log,
	}

	return &fv
}

// Enable enables FileVault and returns the key generated from the command.
// If it fails then an empty string is returned.
func (f *FileVault) Enable(adminUser string, adminPassword string) string {
	key := ""

	f.log.Info.Log("Starting FileVault process")

	out, err := exec.Command("sudo", "bash", "-c", f.script.EnableFileVault,
		adminUser, adminPassword).CombinedOutput()

	outText := string(out)
	logMsg := strings.TrimSpace(fmt.Sprintf("Output: %s", outText))
	f.log.Debug.Log(logMsg, 7)

	if err != nil {
		f.log.Warn.Log("Failed to enable FileVault")
		return ""
	}

	// output is <name> = '<key>'
	outArr := strings.Split(outText, "'")
	// TIL: in Go an empty string is added to the array if the delimiter is at the end!
	// also println is not the same as fmt.Println...
	key = outArr[1]

	f.log.Info.Log("Enabled FileVault")

	return key
}

// Status retrieves the status of FileVault and returns true/false on its status.
//
// If the command failed to run then return an error.
func (f *FileVault) Status() (bool, error) {
	cmd := fmt.Sprintf("sudo -S fdesetup isactive <<< '%s'", f.admin.Password)
	// turns out if isactive == false the exit status is 1. ignoring the error here!
	out, _ := exec.Command("bash", "-c", cmd).Output()

	// instead of a boolean it must be in a string due to the subprocess.
	fileVaultStatus := strings.TrimSpace(strings.ToLower(string(out)))

	// either some fail happened or this is ran on a non-mac OS
	if fileVaultStatus == "" {
		return false, fmt.Errorf("filevault checking failed: %s", fileVaultStatus)
	}

	f.log.Info.Log("FileVault status: %s", fileVaultStatus)

	if strings.Contains(fileVaultStatus, "true") {
		return true, nil
	}

	return false, nil
}

// AddSecureToken adds the user to the SecureToken list for FileVault.
//
// If successful then nil is returned, otherwise an error is returned. Do not leave the
// user on the device, otherwise issues will occur due to FileVault.
func (f *FileVault) AddSecureToken(username string, userPassword string) error {
	// turns out i forgot secure token access... that was rough to find out in prod
	secureTokenCmd := fmt.Sprintf("sudo sysadminctl -secureTokenOn '%s' -password '%s' -adminUser '%s' -adminPassword '%s'",
		username, userPassword, f.admin.Username, f.admin.Password)

	_, err := exec.Command("bash", "-c", secureTokenCmd).Output()
	if err != nil {
		f.log.Error.Log(fmt.Sprintf("Error enabling token for user, manual interaction needed: %v", err))
		return fmt.Errorf("failed to enable secure token for user %s: %v", username, err)
	} else {
		f.log.Info.Log("Secure token added for %s", username)
	}

	return nil
}
