package core

import (
	"errors"
	"fmt"
	"macos-deployment/deploy-files/logger"
	"os/exec"
	"runtime"
	"strings"
)

type Packager struct {
	packagesToInstall map[string][]string
	searchingFiles    []string
	log               *logger.Log
}

// NewPackager creates a new Packager for package modifications for macOS.
func NewPackager(packagesToInstall map[string][]string, searchingFiles []string, logger *logger.Log) *Packager {
	packagesLowered := make(map[string][]string)

	for pkg, appNames := range packagesToInstall {
		lowPkg := strings.ToLower(pkg)

		packagesLowered[lowPkg] = appNames
	}

	packager := Packager{
		packagesToInstall: packagesLowered,
		searchingFiles:    searchingFiles,
		log:               logger,
	}

	return &packager
}

// InstallRosetta installs the Rosetta software required for installing packages.
// If Rosetta is already installed or the CPU is non-arm64, then nil will be returned.
//
// If the installation of Rosetta fails then an error will be returned.
func (p *Packager) InstallRosetta() error {
	if runtime.GOARCH == "amd64" {
		p.log.Info.Log("Skipping Rosetta installation, architecture: %s", runtime.GOARCH)
		return nil
	}

	cmd := "pkgutil --pkgs | grep -i rosetta"

	// due to using grep, if rosetta is not installed it is an error- so errors MUST be ignored.
	// instead the output will be used to handle the error.
	out, _ := exec.Command("bash", "-c", cmd).Output()
	if string(out) == "" {
		_, installErr := exec.Command("sudo", "softwareupdate", "--install-rosetta",
			"--agree-to-license").Output()
		if installErr != nil {
			return errors.New("rosetta failed to install")
		}

		p.log.Info.Log("Rosetta successfully installed")
	} else {
		p.log.Warn.Log("Found existing Rosetta installation")
	}

	return nil
}

// AddPKG adds new packages to the list of packages to install by adding packages
// included in the --include flag. All package names are lowered.
func (p *Packager) AddPackages(packagesToAdd []string) {
	for _, includedPkg := range packagesToAdd {
		// used to extract the package and its installed files from the flag argument.
		includeArgArr := strings.Split(includedPkg, "/")

		p.toLowerArray(&includeArgArr)

		pkg := includeArgArr[0]

		// the installed files from the pkg, i.e. the installed package name.
		pkgInstalledArr := make([]string, 0)
		if len(includeArgArr) > 1 {
			pkgInstalledArr = includeArgArr[1:]
		}

		p.packagesToInstall[pkg] = pkgInstalledArr
		p.log.Info.Log("Added %s to the installation list", pkg)
	}

	p.log.Debug.Log("Packages: %v", p.packagesToInstall)
}

// RemovePackages removes packages from the list of packages to install by removing the packages
// given in the --exclude flag.
func (p *Packager) RemovePackages(packagesToRemove []string) {
	for _, excludedPkg := range packagesToRemove {
		excludedPkgLow := strings.ToLower(excludedPkg)
		_, ok := p.packagesToInstall[excludedPkgLow]
		if ok {
			p.log.Info.Log("Excluded package %s from installation", excludedPkg)
			delete(p.packagesToInstall, excludedPkg)
		}
	}
}

// ReadPackagesDirectory gets the package files path from the package directory.
// It returns an array of strings, based on the relative path from the current directory
// when the binary was executed.
// It requires the embedded find_pkgs.sh script.
//
// If the directory does not exist or if there is an issue reading the directory, then an error
// is returned.
func (p *Packager) ReadPackagesDirectory(pkgDirectory string, getPackageScript string) ([]string, error) {
	p.log.Debug.Log(fmt.Sprintf("Package folder: %s", pkgDirectory))

	out, err := exec.Command("bash", "-c", getPackageScript, pkgDirectory, "*.pkg").Output()
	if err != nil {
		p.log.Error.Log(fmt.Sprintf("Failed to search directory: %v", err))
		return nil, err
	}

	pkgArray := strings.Split(string(out), "\n")

	p.log.Debug.Log("Packages found: %v", pkgArray)

	return pkgArray, nil
}

// InstallPackages installs the keys of the map of packages to install.
// The argument takes an array of relative paths read from the project package directory.
//
// If a package fails to install, then it will be logged and skipped.
func (p *Packager) InstallPackages(packagesPath []string) {
	for pkg, installedPkgNames := range p.packagesToInstall {
		isInstalled := p.isInstalled(installedPkgNames, pkg)

		if isInstalled {
			p.log.Warn.Log(fmt.Sprintf("Package %s is already installed", pkg))
			continue
		}

		pkgLowered := strings.ToLower(pkg)
		// paths relative to the directory that ran the binary, but the contents
		// are required to be in the same directory as the binary.
		for _, file := range packagesPath {
			relativePkgLow := strings.ToLower(file)
			if strings.Contains(relativePkgLow, pkgLowered) {
				p.log.Info.Log("Installing package %s", pkg)

				cmd := fmt.Sprintf(`installer -pkg "%s" -target /`, file)
				p.log.Debug.Log("Package: %s | Package path: %s | Command: %s", pkg, file, cmd)

				out, err := exec.Command("sudo", "bash", "-c", cmd).Output()
				if err != nil {
					outStr := strings.TrimSpace(string(out))
					p.log.Warn.Log(fmt.Sprintf("Failed to install %s: %s %v", pkg, outStr, err))
					continue
				}

				outMsg := "Successfully installed"
				if !strings.Contains(pkgLowered, ".pkg") {
					outMsg = fmt.Sprintf("%s %s.pkg", outMsg, pkg)
				} else {
					outMsg = fmt.Sprintf("%s %s", outMsg, pkg)
				}
				p.log.Info.Log(outMsg)
			}
		}
	}
}

// GetPackages returns the packages that are being installed.
func (p *Packager) GetPackages() []string {
	packages := make([]string, 0, len(p.packagesToInstall))

	for pkg := range p.packagesToInstall {
		packages = append(packages, pkg)
	}

	return packages
}

// isInstalled searches for the names of an installed package in the search directory.
//
// If an installed package name is found in the search directory, true is returned indicating
// the package being installed is already installed.
// Otherwise, false is returned if no installed arguments are given or it doesn't exist in the search
// directories.
func (p *Packager) isInstalled(installedPkgNames []string, pkgToInstall string) bool {
	for _, pkgSearchName := range installedPkgNames {
		// if the given name is blank then install regardless of check.
		if pkgSearchName == "" {
			return false
		}

		lowPkgSearchName := strings.ToLower(pkgSearchName)
		// unfortunately a nested loop is required here due to the array.
		// on the bright side it does exit out early if it finds a match.
		for _, installedPkg := range p.searchingFiles {
			lowInstalledPkgName := strings.ToLower(installedPkg)

			// NOTE: this is a fuzzy finder, so the exact names should be expected.
			// if a generic name is given, there is a good possibility the wrong name will be matched.
			// compares the files in the search directory, to the user defined package name for installation checks.
			if strings.Contains(lowInstalledPkgName, lowPkgSearchName) {
				p.log.Info.Log(fmt.Sprintf("Found existing installation for package %s", pkgToInstall))
				p.log.Debug.Log(fmt.Sprintf("Package: %s | Given package name: %s", pkgToInstall, pkgSearchName))
				return true
			}
		}
	}

	return false
}

// toLowerArray lowers all strings to lowercase in an array by mutating
// the pointer to the given array.
//
// Used to remove the case sensitivity of package matching.
func (p *Packager) toLowerArray(arr *[]string) {
	for i, str := range *arr {
		(*arr)[i] = strings.ToLower(str)
	}
}
