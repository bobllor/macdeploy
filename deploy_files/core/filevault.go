package core

import (
	"fmt"
	"macos-deployment/deploy_files/utils"
	"os/exec"
	"strings"
)

func EnableFileVault() {
	cmd := "fdesetup isactive"
	out, _ := exec.Command("bash", "-c", cmd).Output()
	fileVaultStatus := strings.ToLower(string(out))

	fmt.Printf("[DEBUG] FileVault status: %s", fileVaultStatus)

	if strings.Contains(fileVaultStatus, "false") {
		scriptName := "enable_filevault.sh"

		println("[INFO] Enabling FileVault")

		scriptPath := fmt.Sprintf("%s/%s/%s", utils.Home, "macos-deployment/deploy_files", scriptName)
		out, err := exec.Command("bash", scriptPath).Output()

		outText := string(out)
		fmt.Printf("[DEBUG] Path: %s, Output: %s", scriptPath, outText)

		if err != nil {
			println(err)
			return
		}

		// output is <name> = '<key>'
		outArr := strings.Split(outText, "'")
		// TIL: in Go an empty string is added to the array if the delimiter is at the end!
		// also println is not the same as fmt.Println...
		key := outArr[1]

		fmt.Println("[INFO] FileVault enabled")
		fmt.Printf("[INFO] Generated FileVault key %s", key)
	} else {
		println("[INFO] Skipping FileVault process: Already enabled")
	}
}
