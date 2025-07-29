package pkg

import (
	"os"
	"os/exec"
	"strings"
)

// MakePKG creates a map (hashset-like) that represent the names of the packages from the YAML.
func MakePKG(packages []string, installTeamViewer bool) map[string]bool {
	packagesMap := make(map[string]bool)

	for _, pkg := range packages {
		pkgLowered := strings.ToLower(pkg)
		if !installTeamViewer && strings.Contains(pkgLowered, "teamviewer") {
			continue
		}

		packagesMap[pkgLowered] = true
	}

	return packagesMap
}

// InstallPKG runs a Bash script with arguments to install the given packages.
func InstallPKG(pkg string) {
	// TODO: get the full paths of the packages in the pkg_dir (default installed in the home directory)
	// TODO: pass path arguments into a bash script to install via bash. copy output to a log.
	// TODO: *.pkg is the condition to find packages, however we need to find the full path later.
	var filePath string = "./deploy_files/find_files.sh"

	out, err := exec.Command("bash", filePath, "/home/teboc/Pictures", pkg).Output()
	if err != nil {
		panic(err)
	}
	arr := strings.Split(string(out), "\n")

	for _, file := range arr {
		fileLowered := strings.ToLower(file)

		if strings.Contains(fileLowered, pkg) {
			println(file, "found")
		}
	}
}

func IsInstalled(pkg string, searchPaths []string) bool {
	// this is on a mac, so there are two folders that will be checked:
	// 	1. /Applications/ (general applications)
	//  2. /Library/Application\ Support/ (service files)
	// majority of applications will be installed in 1, but few do appear only in 2.
	for _, path := range searchPaths {
		files, err := os.ReadDir(path)
		if err != nil {
			// TODO: handle this properly
			panic(err)
		}

		for _, file := range files {
			pkgLowered := strings.ToLower(pkg)
			fileNameLowered := strings.ToLower(file.Name())

			if strings.Contains(fileNameLowered, pkgLowered) {
				return true
			}
		}
	}

	return false
}
