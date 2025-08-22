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

// AddPKG adds new packages to an existing map.
func AddPKG(packages map[string][]string, addedPackages []string) {
	for _, includedPkg := range addedPackages {
		argArr := strings.Split(includedPkg, "/")

		mainPkg := argArr[0]
		pkgInstallNameArr := make([]string, 0)

		if len(argArr) > 1 {
			pkgInstallNameArr = argArr[1:]
		}

		packages[mainPkg] = pkgInstallNameArr
	}

	logger.Log(fmt.Sprintf("Packages: %v", packages), 7)
}

// InstallPKG runs a Bash script with arguments to install the given packages.
//
// foundPKGs is an array of strings that consist of all packages found in the packages directory.
func InstallPKG(pkg string, foundPKGs []string) {
	pkg = strings.ToLower(pkg)

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
func IsInstalled(pkgNames []string, searchPaths *[]map[string]bool) bool {
	// this is on a mac, there are two folders that will be checked:
	// 	1. /Applications/ (general applications)
	//  2. /Library/Application\ Support/ (service files)
	// however if in any event these changes, you can configure it in config.yaml
	for _, pkg := range pkgNames {
		// unfortunately double loop is required here due to the array condition.
		// on the bright side it does exit out early if it finds a match.

		for _, pathMap := range *searchPaths {
			pkgLowered := strings.ToLower(pkg)

			if _, found := pathMap[pkgLowered]; found {
				logger.Log(fmt.Sprintf("Found existing installation for package %s", pkg), 6)
				return true
			}
		}
	}

	return false
}

// RemovePKG removes packages from a map by iterating over a list of packages to remove.
func RemovePKG(pkgMap map[string][]string, packagesToRemove []string) {
	for _, excludedPkg := range packagesToRemove {
		_, ok := pkgMap[excludedPkg]
		if ok {
			logger.Log(fmt.Sprintf("Excluding package: %s", excludedPkg), 6)
			delete(pkgMap, excludedPkg)
		}
	}
}
