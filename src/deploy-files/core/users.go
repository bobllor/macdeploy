package core

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/bobllor/macdeploy/src/deploy-files/logger"
	"github.com/bobllor/macdeploy/src/deploy-files/scripts"
	"github.com/bobllor/macdeploy/src/deploy-files/utils"
	"github.com/bobllor/macdeploy/src/deploy-files/yaml"
)

type UserMaker struct {
	adminInfo yaml.UserInfo
	log       *logger.Logger
	script    *scripts.BashScripts
}

// NewUser creates a new UserMaker to handle user creation.
func NewUser(adminInfo yaml.UserInfo, scripts *scripts.BashScripts, logger *logger.Logger) *UserMaker {
	user := UserMaker{
		adminInfo: adminInfo,
		log:       logger,
		script:    scripts,
	}

	return &user
}

// CreateAccount creates the local user account on the device.
// Empty usernames and passwords will have a prompt, with the password being set for
// UserInfo if empty.
//
// It will return the internal username of the macOS account upon success.
// If there are any errors then an error is returned.
func (u *UserMaker) CreateAccount(user *yaml.UserInfo, isAdmin bool) (string, error) {
	// username will be used for both entries needed.
	username := user.Username

	if username == "" {
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("\nHit enter to skip the user creation.")
		fmt.Print("Enter the client's name: ")

		input, _ := reader.ReadString('\n')

		fmt.Println("") // for formatting purposes.

		input = input[:len(input)-1]

		if input == "" {
			u.log.Info("User creation skipped")
			return "", errors.New("user creation skipped")
		}

		username = input
	}

	u.log.Debugf("User: %s", username)

	if user.Password == "" {
		u.log.Warnf("No user password was given for %s", username)

		err := user.SetPassword(true)
		if err != nil {
			return "", err
		}

		u.log.Info("Updated user password, previously was empty")
	}

	// follows apple's naming convention
	accountName := utils.FormatUsername(username)

	u.log.Infof("Creating user %s with account name %s", username, accountName)

	admin := "false"
	if isAdmin && !user.IgnoreAdmin {
		u.log.Infof("Admin enabled for user %s", username)
		admin = strconv.FormatBool(isAdmin)
	}

	userExists, err := u.userExists(accountName)
	if err != nil {
		return "", fmt.Errorf("error occurred reading user directory %s: %v", username, err)
	}
	if userExists {
		return "", fmt.Errorf("user %s already exists in the system", username)
	}

	// CreateUserScript takes 3 arguments.
	out, err := exec.Command("sudo", "bash", "-c",
		u.script.CreateUser, username, accountName, user.Password, admin).CombinedOutput()
	if err != nil {
		u.log.Debug(fmt.Sprintf("create user script error: %s", string(out)))
		return "", fmt.Errorf("failed to create user %s: %v", username, err)
	}

	createdLog := fmt.Sprintf("User %s created", username)
	u.log.Info(createdLog)

	return accountName, nil
}

// DeleteAccount removes the given user from the device.
func (u *UserMaker) DeleteAccount(username string) error {
	cmd := fmt.Sprintf(
		"sudo sysadminctl -deleteUser '%s' -adminUser '%s' -adminPassword '%s'",
		username,
		u.adminInfo.Username,
		u.adminInfo.Password)
	_, err := exec.Command("sudo", "bash", "-c", cmd).Output()
	if err != nil {
		return err
	}

	u.log.Info(fmt.Sprintf("Removed user %s", username))

	return nil
}

// GrantAdmin grants the given user admin privileges.
// The user must exist and is not an admin.
func (u *UserMaker) GrantAdmin(username string) error {
	err := u.adminInfo.InitializeSudo()
	if err != nil {
		return fmt.Errorf("failed to initialize sudo (%v)", err)
	}

	stat, err := u.isAdmin(username)
	if err != nil {
		return fmt.Errorf("failed to check admin status (%v)", err)
	}
	if stat {
		return fmt.Errorf("user %s is already admin", username)
	}

	cmd := `
	#!/usr/bin/env bash

	username=$0
	dseditgroup -o edit -a "$username" -t user admin

	user_group=$(groups "$username")
	if [[ "$user_group" =~ "admin" ]]; then
		echo true
	else
		echo false
	fi
	`

	b, err := exec.Command("sudo", "bash", "-c", cmd, username).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to grant admin (%v)", err)
	}

	out := strings.TrimSpace(string(b))
	u.log.Debugf("Grant admin output: %s", out)

	if out == "false" {
		return fmt.Errorf("failed to grant admin to user %s", username)
	}

	return nil
}

// RevokeAdmin revokes the admin privileges from the user.
// The user must exist and is an admin.
func (u *UserMaker) RevokeAdmin(username string) error {
	err := u.adminInfo.InitializeSudo()
	if err != nil {
		return fmt.Errorf("failed to initialize sudo (%v)", err)
	}

	stat, err := u.isAdmin(username)
	if err != nil {
		return fmt.Errorf("failed to check admin status (%v)", err)
	}
	if !stat {
		return fmt.Errorf("user %s is already not admin", username)
	}

	cmd := `
	#!/usr/bin/env bash

	username=$0
	dseditgroup -o edit -d "$username" -t user admin

	user_group=$(groups "$username")
	if [[ "$user_group" =~ "admin" ]]; then
		echo true
	else
		echo false
	fi
	`

	b, err := exec.Command("sudo", "bash", "-c", cmd, username).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to revoke admin (%v)", err)
	}

	out := strings.TrimSpace(string(b))
	u.log.Debugf("Revoke admin output: %s", out)

	if out == "true" {
		return fmt.Errorf("failed to revoke admin to user %s", username)
	}

	return nil
}

// AddPasswordPolicy adds a password policy for the user. This can only be
// ran after adding the Secure Token to the user.
//
// If the password policy fails to run then an error returns.
func (u *UserMaker) AddPasswordPolicy(username string) error {
	pwPolicyCmd := fmt.Sprintf("sudo pwpolicy -u '%s' -setpolicy 'newPasswordRequired=1'", username)

	err := exec.Command("bash", "-c", pwPolicyCmd).Run()
	if err != nil {
		return fmt.Errorf("failed to create user policy for %s: %v", username, err)
	}

	u.log.Info(fmt.Sprintf("Added new password policy for %s", username))

	return nil
}

// checkAdmin checks if the current user is admin. If it fails to run,
// it will return an error.
//
// The existence of the user is checked in the call and will return an error
// if they do not exist.
func (u *UserMaker) isAdmin(username string) (bool, error) {
	cmd := `
	#!/usr/bin/env bash
	username=$0

	user_group=$(groups "$username")
	if [[ "$user_group" =~ "admin" ]]; then
		echo true
	else
		echo false
	fi
	`
	exist, err := u.userExists(username)
	if err != nil {
		return false, err
	}
	if !exist {
		return false, fmt.Errorf("user %s does not exist", username)
	}

	b, err := exec.Command("sudo", "bash", "-c", cmd, strings.ToLower(username)).CombinedOutput()
	if err != nil {
		return false, err
	}

	out := strings.TrimSpace(string(b))
	if out == "true" {
		return true, nil
	}

	u.log.Debugf("Admin command output: %s", out)

	return false, nil
}

// userExists checks the Users directory for the given username.
// This reads from the /Users directory on a MacBook.
func (u *UserMaker) userExists(username string) (bool, error) {
	usersPath := "/Users"

	dirs, err := os.ReadDir(usersPath)
	if err != nil {
		u.log.Warn(fmt.Sprintf("Error reading directory: %v", err))
		return false, err
	}

	u.log.Debug(fmt.Sprintf("User directory content: %v", dirs))

	for _, dir := range dirs {
		dirName := strings.ToLower(dir.Name())
		lowerUsername := strings.ToLower(username)

		if strings.Contains(dirName, lowerUsername) {
			return true, nil
		}
	}

	return false, nil
}
