package users

import (
	"bufio"
	"macos-deployment/deploy_files/utils"
	"os"
	"os/exec"
	"strings"
)

// CreateAccount creates the user account on the device.
func CreateAccount(user utils.User, isAdmin string) error {
	var CreateUserScript string = utils.Home + "/macos-deployment/deploy_files/create_user.sh"

	// userName will be used for both entries needed.
	fullName := user.FullName
	userName := user.UserName

	if fullName == "" || strings.ToUpper(fullName) == "DONOTUSE" {
		reader := bufio.NewReader(os.Stdin)

		input, _ := reader.ReadString('\n')

		fullName = input

		// TODO: parse the userName here
	}

	exec.Command("sudo", "bash", CreateUserScript, userName, userName, user.Password, isAdmin)

	return nil
}
