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

	// userCache is a in-memory cache of the users in the users directory.
	// This will be populated on the existence check. All keys are lowercased.
	userCache map[string]any
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

		fmt.Println("User password required")
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

// DeleteAccount removes the given user from the device. This requires
// the user to exist.
func (u *UserMaker) DeleteAccount(username string) error {
	cmd := fmt.Sprintf(
		"sudo sysadminctl -deleteUser '%s'",
		username,
	)

	// the output is not in stdout, it is not possible to capture.
	// error codes:
	//	- user not found (255)
	_, err := exec.Command("sudo", "bash", "-c", cmd).Output()
	if err != nil {
		newErr := errors.New("account deletion failed")
		if strings.Contains(err.Error(), "255") {
			newErr = fmt.Errorf("user %s does not exist", username)
		}

		u.log.Warnf("Failed to delete account: %v | %v", err, newErr)
		return newErr
	}

	u.log.Info(fmt.Sprintf("Removed user %s", username))

	return nil
}

// GrantAdmin grants the given user admin privileges.
// The user must exist and is not an admin.
func (u *UserMaker) GrantAdmin(username string) error {
	err := u.adminInfo.InitializeSudo()
	if err != nil {
		return fmt.Errorf("failed to initialize sudo: %v", err)
	}

	stat, err := u.isAdmin(username)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %v", err)
	}
	if stat {
		return fmt.Errorf("user %s is admin", username)
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
		return fmt.Errorf("failed to grant admin: %v", err)
	}

	out := strings.TrimSpace(string(b))
	u.log.Debugf("Grant admin output: %s", out)

	if out == "false" {
		u.log.Warnf("Granting admin failed for %s | out=%s,err=%v", username, out, err)
		return fmt.Errorf("failed to grant admin for user %s", username)
	}

	return nil
}

// RevokeAdmin revokes the admin privileges from the user.
// The user must exist and is an admin.
func (u *UserMaker) RevokeAdmin(username string) error {
	err := u.adminInfo.InitializeSudo()
	if err != nil {
		return fmt.Errorf("failed to initialize sudo: %v", err)
	}

	stat, err := u.isAdmin(username)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %v", err)
	}
	if !stat {
		return fmt.Errorf("user %s is not admin", username)
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
		return fmt.Errorf("failed to revoke admin for user %s", username)
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

// List lists the local users on the device. The slice contains the
// internal usernames, not the display names.
func (u *UserMaker) List() ([]string, error) {
	usersPath := "/Users"

	// cache is not used due to it potentially being stale
	dirs, err := os.ReadDir(usersPath)
	if err != nil {
		u.log.Warn(fmt.Sprintf("Error reading users directory: %v", err))
		return nil, err
	}

	u.log.Debugf("Found %d entries in %s", len(dirs), usersPath)
	users := []string{}

	for _, d := range dirs {
		lowerDirName := strings.ToLower(d.Name())
		if !u.isNotUser(lowerDirName) {
			u.log.Debugf("User folder %s skipped", d)
			continue
		}

		users = append(users, lowerDirName)
	}

	return users, nil
}

// userExists checks the Users directory for the given username.
// This reads from the /Users directory on a MacBook.
func (u *UserMaker) userExists(username string) (bool, error) {
	// no dscl . -list /users due to it requiring more parsing and
	// an additional subprocess exec
	usersPath := "/Users"

	_, ok := u.userCache[username]
	if ok {
		return true, nil
	}

	// reset the cache if the user does not exist
	u.userCache = make(map[string]any)

	dirs, err := os.ReadDir(usersPath)
	if err != nil {
		u.log.Warn(fmt.Sprintf("Error reading users directory: %v", err))
		return false, err
	}

	u.log.Debugf("Found %d entries in %s", len(dirs), usersPath)

	for _, dir := range dirs {
		lowerDirName := strings.ToLower(dir.Name())
		if !u.isNotUser(lowerDirName) {
			u.log.Debugf("User folder %s skipped", dir)
			continue
		}

		u.userCache[lowerDirName] = struct{}{}
	}

	lowerUsername := strings.ToLower(username)
	_, ok = u.userCache[lowerUsername]
	if ok {
		return true, nil
	}

	return false, nil
}

// isNotUser checks if the given user string is a valid user.
// Due to the /Users path containing non-users on a MacBook device,
// this checks for the default folders that exist prior to no users.
//
// It is considered not a user if it is named or contains the following:
//   - Deleted Users
//   - Shared
//   - Leading period
func (u *UserMaker) isNotUser(user string) bool {
	if len(user) == 0 {
		return false
	}

	// probably subject to change
	blacklistedWords := []string{
		"deleted users",
		"shared",
	}

	for _, word := range blacklistedWords {
		user = strings.ToLower(user)
		if strings.Contains(user, word) || user[0] == '.' {
			return false
		}
	}

	return true
}
