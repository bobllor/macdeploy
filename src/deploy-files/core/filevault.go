package core

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/bobllor/macdeploy/src/deploy-files/logger"
	"github.com/bobllor/macdeploy/src/deploy-files/scripts"
	"github.com/bobllor/macdeploy/src/deploy-files/yaml"
)

type FileVault struct {
	admin  yaml.UserInfo
	script *scripts.BashScripts
	log    *logger.Logger
}

func NewFileVault(admin yaml.UserInfo, script *scripts.BashScripts, log *logger.Logger) *FileVault {
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

	f.log.Info("Starting FileVault process")

	out, err := exec.Command("sudo", "bash", "-c", f.script.EnableFileVault,
		adminUser, adminPassword).CombinedOutput()
	outText := string(out)
	if err != nil {
		f.log.Warn("Failed to enable FileVault")
		return ""
	}

	// output is <name> = '<key>'
	outArr := strings.Split(outText, "'")
	// TIL an empty string is added to the array if there is a delimiter at the end!
	key = outArr[len(outArr)-2]

	f.log.Info("FileVault enabled")

	return key
}

// Disable disables FileVault. It will return a return a bool indicating if it
// was successful or not.
func (f *FileVault) Disable(adminUser string, adminPassword string) bool {
	f.log.Info("Disabling FileVault")

	out, err := exec.Command("sudo", "bash", "-c", f.script.DisableFileVault,
		adminUser, adminPassword).CombinedOutput()
	outText := string(out)
	if err != nil {
		f.log.Warnf("Failed to disable FileVault: %v", err)
		return false
	}

	f.log.Debugf("Disable FileVault output: %s", outText)

	return true
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
		return false, fmt.Errorf("%s", fileVaultStatus)
	}

	f.log.Debugf("FileVault status: %s", fileVaultStatus)

	if strings.Contains(fileVaultStatus, "true") {
		f.log.Info("Filevault is already enabled")
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
	secureTokenCmd := fmt.Sprintf(
		"sudo sysadminctl -secureTokenOn '%s' -password '%s' -adminUser '%s' -adminPassword '%s'",
		username, userPassword, f.admin.Username, f.admin.Password)

	f.log.Debugf("Ran secure token command for user %s", username)

	// VERY IMPORTANT:
	// sysadminctl outputs to tty and it always returns 0.
	// this means it is not possible to determine if the command fails.
	// the user's password, user's username, and the admin username are not the point of failure.
	// the point of failure is the admin password, because this can either be wrong from the config
	// or the terminal input was wrong.
	err := f.admin.ResetSudo()
	if err != nil {
		f.log.Warnf("Failed to run sudo reset command: %v", err)
		return err
	}

	err = f.admin.InitializeSudo()
	if err != nil {
		f.log.Warnf("Error enabling token for user, manual interaction needed: %v", err)
		f.log.Warn("Admin password is likely incorrect")
		return err
	}

	_ = exec.Command("bash", "-c", secureTokenCmd).Run()
	f.log.Infof("Secure token added for %s", username)

	return nil
}
