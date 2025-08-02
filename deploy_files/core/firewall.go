package core

import (
	"errors"
	"fmt"
	"macos-deployment/deploy_files/logger"
	"macos-deployment/deploy_files/utils"
	"os/exec"
	"strings"
)

// EnableFireWall enables the firewall of the Mac.
func EnableFireWall() {
	scriptName := "enable_firewall.sh"
	scriptPath := fmt.Sprintf("%s/%s/%s", utils.MainDir, utils.ScriptDir, scriptName)

	firewallIsOn, err := firewallIsEnabled()
	if err != nil {
		firewallErrMsg := fmt.Sprintf("Error executing Firewall check script %s", err.Error())
		logger.Log(firewallErrMsg, 3)
		return
	}

	logger.Log(fmt.Sprintf("Path: %s | Firewall status: %t", scriptPath, firewallIsOn), 7)

	if !firewallIsOn {
		out, err := exec.Command("bash", scriptPath).CombinedOutput()
		if err != nil {
			errMsg := fmt.Sprintf("Failed to enable Firewall | Error: %s", string(out))
			logger.Log(errMsg, 3)

			return
		}

		logger.Log("Firewall successfully enabled", 6)
	} else {
		logger.Log("Firewall is already enabled", 6)
	}
}

// firewallIsEnabled gets the status of the firewall.
func firewallIsEnabled() (bool, error) {
	cmd := "sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate"
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to check Firewall status | Error: %s", string(out))
		logger.Log(errMsg, 3)

		return false, errors.New(string(out))
	}

	statusText := strings.ToLower(string(out))

	return strings.Contains(statusText, "disabled"), nil
}
