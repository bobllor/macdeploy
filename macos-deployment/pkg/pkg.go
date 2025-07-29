package pkg

import (
	"strings"
)

// MakePKG creates a map (hashset-like) that represent the names of the packages from the YAML.
func MakePKG(packages []string, installTeamViewer bool) map[string]bool {
	packagesMap := make(map[string]bool)

	for _, pkg := range packages {
		pkgLowered := strings.ToLower(pkg)
		if installTeamViewer && strings.Contains(pkgLowered, "teamviewer") {
			continue
		}

		packagesMap[pkgLowered] = true
	}

	return packagesMap
}

// InstallPKG runs a Bash script with arguments to install the given packages.
func InstallPKG() {
	// TODO: get the full paths of the packages in the pkg_dir (default installed in the home directory)
	// TODO: pass path arguments into a bash script to install via bash. copy output to a log.
}

func IsInstalled(pkg string) bool {
	// this is on a mac, so there are two folders that will be checked:
	// 	1. /Applications/ (general applications)
	//  2. /Library/Application\ Support/ (service files)
	// majority of applications will be installed in 1, but few do appear only in 2.
	appPath := "/Applications/"
	systemAppPath := "/Library/Application \\ Support/"

	return false
}
