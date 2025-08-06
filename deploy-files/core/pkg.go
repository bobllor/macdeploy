package core

import (
	"errors"
	"fmt"
	"macos-deployment/deploy-files/logger"
	"os/exec"
	"strings"
)

// InstallRosetta installs the Rosetta software required for installing packages.
// If Rosetta is already installed, then this will be skipped.
//
// Package installations require Rosetta, this is required to be called.
// If the installation of Rosetta fails then an error will be returned.
func InstallRosetta() error {
	cmd := "pkgutil --pkgs | grep -i rosetta"

	// if rosetta is not installed the exec fails, so errors MUST be ignored.
	roseOut, _ := exec.Command("bash", "-c", cmd).Output()

	if string(roseOut) == "" {
		_, installErr := exec.Command("sudo", "softwareupdate", "--install-rosetta",
			"--agree-to-license").Output()
		if installErr != nil {
			return errors.New("rosetta failed to install")
		}

		logger.Log("Rosetta installed", 6)
	} else {
		// FIXME: add logging here
		logger.Log("Rosetta is already installed", 6)
	}

	return nil
}

// MakePKG creates a map with keys being the exact pkg file name and values being an array of strings
// used to find if a pkg is installed in a given searchDirectory.
// This is used to install the packages by accessing the pkg in their directory.
// The keys in the map are all lowercase.
//
// If installTeamViewer is True then TeamViewer is added into the map as a key if it exists.
// By default TeamViewer is not installed.
func MakePKG(packages map[string][]string, installTeamViewer bool) map[string][]string {
	newPackagesMap := make(map[string][]string)

	for pkg, pkgArr := range packages {
		pkgLowered := strings.ToLower(pkg)
		if !installTeamViewer && strings.Contains(pkgLowered, "teamviewer") {
			logger.Log("Removing TeamViewer from installation package", 6)
			logger.Log(fmt.Sprintf("TeamViewer flag: %v", installTeamViewer), 7)
			continue
		}

		newPackagesMap[pkgLowered] = pkgArr
	}

	return newPackagesMap
}

// InstallPKG runs a Bash script with arguments to install the given packages.
//
// foundPKGs is an array of strings that consist of all packages found in the packages directory.
func InstallPKG(pkg string, foundPKGs []string) {
	for _, file := range foundPKGs {
		fileLowered := strings.ToLower(file)

		if strings.Contains(fileLowered, pkg) {
			logger.Log(fmt.Sprintf("Installing package %s", pkg), 6)

			cmd := fmt.Sprintf("installer -pkg %s -target /", file)
			_, pkgErr := exec.Command("sudo", "bash", "-c", cmd).Output()
			if pkgErr != nil {
				// FIXME: add logging
				logger.Log(fmt.Sprintf("Failed to install %s.pkg\n", pkg), 3)
			}

			logger.Log(fmt.Sprintf("[DEBUG] Package: %s | Package Path: %s | Command: %s", pkg, file, cmd), 7)
			logger.Log(fmt.Sprintf("Successfully installed %s.pkg", pkg), 6)
			break
		}
	}
}

// IsInstalled searches for a given package in a search path from a given array of paths.
// Ensure all keys in searchPaths are lowercase, which can be done by using the function GetFileMap.
func IsInstalled(pkgNames []string, searchPaths []map[string]bool) bool {
	// this is on a mac, there are two folders that will be checked:
	// 	1. /Applications/ (general applications)
	//  2. /Library/Application\ Support/ (service files)
	// however if in any event these changes, you can configure it in config.yaml
	for _, pkg := range pkgNames {
		// unfortunately double loop is required here due to the array condition.
		// on the bright side it does exit out early if it finds a match.
		for _, pathMap := range searchPaths {
			pkgLowered := strings.ToLower(pkg)

			if _, found := pathMap[pkgLowered]; found {
				// FIXME: add logging
				logger.Log(fmt.Sprintf("Found existing installation for package %s", pkg), 6)
				return true
			}
		}
	}

	return false
}
