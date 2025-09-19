package core

import (
	"errors"
	"fmt"
	"macos-deployment/deploy-files/logger"
	"os/exec"
	"runtime"
	"strings"
)

// InstallRosetta installs the Rosetta software required for installing packages.
// If Rosetta is already installed or the CPU is non-arm64, then this will be skipped.
//
// Package installations require Rosetta, this is required to be called.
// If the installation of Rosetta fails then an error will be returned.
func InstallRosetta() error {
	if runtime.GOARCH == "amd64" {
		logger.Log(fmt.Sprintf("Skipping Rosetta, device: %s", runtime.GOARCH), 6)
		return nil
	}

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
func InstallPKG(pkg string, foundPKGs *[]string) error {
	pkg = strings.ToLower(pkg)

	for _, file := range *foundPKGs {
		fileLowered := strings.ToLower(file)

		if strings.Contains(fileLowered, pkg) {
			logger.Log(fmt.Sprintf("Installing package %s", pkg), 6)

			cmd := fmt.Sprintf(`installer -pkg "%s" -target /`, file)
			logger.Log(fmt.Sprintf("Package: %s | Package Path: %s | Command: %s", pkg, file, cmd), 7)

			_, pkgErr := exec.Command("sudo", "bash", "-c", cmd).Output()
			if pkgErr != nil {
				return pkgErr
			}

			logger.Log(fmt.Sprintf("Successfully installed %s.pkg", pkg), 6)

			return nil
		}
	}

	logger.Log(fmt.Sprintf("Unable to install package %s.pkg", pkg), 4)
	return fmt.Errorf("unable to find package %s.pkg", pkg)
}

// IsInstalled searches for a given package in a search path from a given array of paths.
// Ensure all keys in searchPaths are lowercase, which can be done by using the function GetFileMap.
//
// pkgNames contains the file names of an installed .pkg, not the actual .pkg installer.
// These can be found in the default directories where they are installed, for example /Applications.
func IsInstalled(pkgNames []string, searchPaths []string, pkgToInstall string) bool {
	for _, pkg := range pkgNames {
		// if no installed names are given, then install regardless.
		if pkg == "" {
			return false
		}

		pkgLowered := strings.ToLower(pkg)
		// unfortunately a nested loop is required here due to the array.
		// on the bright side it does exit out early if it finds a match.
		for _, installedPkg := range searchPaths {
			loweredInstalledPkgName := strings.ToLower(installedPkg)

			// NOTE: this is a fuzzy finder, so the exact names should be expected.
			// if a generic name is given, there is a good possibility the wrong name will be matched.
			if strings.Contains(loweredInstalledPkgName, pkgLowered) {
				logger.Log(fmt.Sprintf("Found existing installation for package %s", pkgLowered), 6)
				logger.Log(fmt.Sprintf("Package: %s | Given package name: %s", pkgToInstall, pkgLowered), 7)
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
