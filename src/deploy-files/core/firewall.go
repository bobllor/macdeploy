package core

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/bobllor/macdeploy/src/deploy-files/logger"
	"github.com/bobllor/macdeploy/src/deploy-files/scripts"
)

type Firewall struct {
	log    *logger.Logger
	script *scripts.BashScripts
}

func NewFirewall(log *logger.Logger, scripts *scripts.BashScripts) *Firewall {
	return &Firewall{
		log:    log,
		script: scripts,
	}
}

// Enable enables the Firewall.
func (f *Firewall) Enable() error {
	out, err := exec.Command("sudo", "bash", "-c", f.script.EnableFirewall).CombinedOutput()
	if err != nil {
		// FIXME: i dont remember why i use string(out) instead of just error. i added err in a rewrite.
		return fmt.Errorf("failed to enable Firewall: %s | %v", string(out), err)
	}

	f.log.Info("Firewall enabled")

	return nil
}

// Status gets the status of the firewall.
func (f *Firewall) Status() (bool, error) {
	cmd := "sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate"
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(fmt.Sprintf("Failed to check Firewall status: %s", string(out)))
		f.log.Warn(errMsg)

		return false, errors.New(string(out))
	}

	statusText := strings.ToLower(string(out))

	status := strings.Contains(statusText, "enabled")

	return status, nil
}
