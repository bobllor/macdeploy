package core

import (
	"errors"
	"fmt"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"os/exec"
	"strings"
)

// EnableFireWall enables the firewall of the Mac.
func EnableFireWall() {
	firewallIsOn, err := firewallIsEnabled()
	if err != nil {
		firewallErrMsg := strings.TrimSpace(fmt.Sprintf("Failed to execute Firewall script | %s", err.Error()))
		logger.Log(firewallErrMsg, 3)
		return
	}

	logger.Log(fmt.Sprintf("Firewall status: %t", firewallIsOn), 7)

	if !firewallIsOn {
		out, err := exec.Command("sudo", "bash", "-c", scripts.EnableFirewallScript).CombinedOutput()
		if err != nil {
			errMsg := fmt.Sprintf("Failed to enable Firewall | %s", string(out))
			logger.Log(errMsg, 3)

			return
		}

		logger.Log("Firewall enabled", 6)
	} else {
		logger.Log("Firewall is already enabled", 6)
	}
}

// firewallIsEnabled gets the status of the firewall.
func firewallIsEnabled() (bool, error) {
	cmd := "sudo /usr/libexec/ApplicationFirewall/socketfilterfw --getglobalstate"
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(fmt.Sprintf("Failed to check Firewall status | %s", string(out)))
		logger.Log(errMsg, 3)

		return false, errors.New(string(out))
	}

	statusText := strings.ToLower(string(out))

	return strings.Contains(statusText, "enabled"), nil
}
