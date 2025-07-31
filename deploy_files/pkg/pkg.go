package pkg

import (
	"fmt"
	"os/exec"
	"strings"
)

// InstallRosetta installs the Rosetta software required for installing packages.
// If Rosetta is already installed, then this will be skipped.
// This function is required to be called before the package installation is called.
func InstallRosetta() {
	cmd := "pkgutil --pkgs | grep -i rosetta"
	roseOut, _ := exec.Command("bash", "-c", cmd).Output()

	if string(roseOut) != "" {
		installOut, installErr := exec.Command("sudo", "softwareupdate", "--install-rosetta",
			"--agree-to-license").Output()
		if installErr != nil {
			// FIXME: add logging here, this is a critical fail and exits the script.
			println(string(installOut))
			panic(installErr)
		}

		println("[INFO] Rosetta has been installed")
	} else {
		// FIXME: add logging here
		println("[INFO] Rosetta is already installed")
	}

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
			println("[INFO] Removing TeamViewer from installation package")
			continue
		}

		newPackagesMap[pkgLowered] = pkgArr
	}

	return newPackagesMap
}

// InstallPKG runs a Bash script with arguments to install the given packages.
func InstallPKG(pkg string, foundPKGs []string) {
	// TODO: get the full paths of the packages in the pkg_dir (default installed in the home directory)
	// TODO: pass path arguments into a bash script to install via bash. copy output to a log.
	// TODO: *.pkg is the condition to find packages, however we need to find the full path later.
	for _, file := range foundPKGs {
		fileLowered := strings.ToLower(file)

		if strings.Contains(fileLowered, pkg) {
			// abs path is probably not needed, it's working from home directory
			println(file, "found")
			cmd := fmt.Sprintf("installer -pkg %s -target /", file)

			pkgOut, pkgErr := exec.Command("sudo", "bash", "-c", cmd).Output()
			if pkgErr != nil {
				// FIXME: add logging
				fmt.Printf("[WARNING] Failed to install %s.pkg\n", pkg)
				println("[DEBUG] Package: %s | Package Path: %s | Command: %s", pkg, file, cmd)
				println(string(pkgOut))
				println(pkgErr)
				break
			}
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
				fmt.Printf("[INFO] Package %s is already installed\n", pkg)
				return true
			}
		}
	}

	return false
}
