package core

import (
	"errors"
	"fmt"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Installer struct {
	packagesToInstall    map[string][]string
	searchDirectoryFiles []string
	log                  *logger.Log
	script               *scripts.BashScripts
}

// NewInstaller creates a new Installer for package and app installations for macOS.
func NewInstaller(packagesToInstall map[string][]string,
	searchDirectoryFiles []string, logger *logger.Log, scripts *scripts.BashScripts) *Installer {
	packagesLowered := make(map[string][]string)

	for pkg, appNames := range packagesToInstall {
		lowPkg := strings.ToLower(pkg)

		packagesLowered[lowPkg] = appNames
	}

	packager := Installer{
		packagesToInstall:    packagesLowered,
		searchDirectoryFiles: searchDirectoryFiles,
		log:                  logger,
		script:               scripts,
	}

	return &packager
}

// InstallRosetta installs the Rosetta software required for installing packages.
// If Rosetta is already installed or the CPU is non-arm64, then nil will be returned.
//
// If the installation of Rosetta fails then an error will be returned.
func (p *Installer) InstallRosetta() error {
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
func (i *Installer) AddPackages(packagesToAdd []string) {
	for _, includedPkg := range packagesToAdd {
		// used to extract the package and its installed files from the flag argument.
		includeArgArr := strings.Split(includedPkg, "/")

		i.toLowerArray(&includeArgArr)

		pkg := includeArgArr[0]

		// the installed files from the pkg, i.e. the installed package name.
		pkgInstalledArr := make([]string, 0)
		if len(includeArgArr) > 1 {
			pkgInstalledArr = includeArgArr[1:]
		}

		i.packagesToInstall[pkg] = pkgInstalledArr
		i.log.Info.Log("Added %s to the installation list", pkg)
	}

	i.log.Debug.Log("Packages: %v", i.packagesToInstall)
}

// RemovePackages removes packages from the list of packages to install by removing the packages
// given in the --exclude flag.
func (p *Installer) RemovePackages(packagesToRemove []string) {
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
func (p *Installer) ReadPackagesDirectory(pkgDirectory string, getPackageScript string) ([]string, error) {
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
func (i *Installer) InstallPackages(packagesPath []string) {
	for pkg, installedNames := range i.packagesToInstall {
		isInstalled := i.IsInstalled(installedNames, pkg)

		if isInstalled {
			i.log.Warn.Log(fmt.Sprintf("Package %s is already installed", pkg))
			continue
		}

		// used for logging at the end
		successfulInstall := false
		failedInstall := false

		pkgLowered := strings.ToLower(pkg)
		// paths relative to the directory that ran the binary, but the contents
		// are required to be in the same directory as the binary.
		for _, file := range packagesPath {
			relativePkgLow := strings.ToLower(file)
			if strings.Contains(relativePkgLow, pkgLowered) {
				i.log.Info.Log("Installing package %s", pkg)

				cmd := fmt.Sprintf(`installer -pkg "%s" -target /`, file)
				i.log.Debug.Log("Package: %s | Package path: %s | Command: %s", pkg, file, cmd)

				out, err := exec.Command("sudo", "bash", "-c", cmd).Output()
				if err != nil {
					outStr := strings.TrimSpace(string(out))
					i.log.Warn.Log(fmt.Sprintf("Failed to install %s: %s %v", pkg, outStr, err))
					failedInstall = true
					break
				}

				outMsg := "Successfully installed"
				if !strings.Contains(pkgLowered, ".pkg") {
					outMsg = fmt.Sprintf("%s %s.pkg", outMsg, pkg)
				} else {
					outMsg = fmt.Sprintf("%s %s", outMsg, pkg)
				}
				i.log.Info.Log(outMsg)

				successfulInstall = true
				break
			}
		}

		if !successfulInstall && !failedInstall {
			i.log.Warn.Log("Unable to find package %s", pkg)
		}
	}
}

// GetPackages returns the packages that are being installed.
func (i *Installer) GetPackages() []string {
	packages := make([]string, 0, len(i.packagesToInstall))

	for pkg := range i.packagesToInstall {
		packages = append(packages, pkg)
	}

	return packages
}

// GetAllPackages gets the packages to be installed and its installed file names array.
func (i *Installer) GetAllPackages() map[string][]string {
	return i.packagesToInstall
}

// ReadDmgDirectory reads the distribution directory for files with the .dmg extension,
// and returns an string array consisting the relative paths to the DMG file.
//
// It will return an error if there is an issue reading the directory.
func (i *Installer) ReadDmgDirectory(dir string) ([]string, error) {
	out, err := exec.Command("bash", "-c", i.script.FindFiles, dir, "*.dmg").Output()
	if err != nil {
		return nil, err
	}

	dmgArray := strings.Split(string(out), "\n")
	// will return arr[:len-1] because an empty string is present at the end of the array
	dmgArray = dmgArray[:len(dmgArray)-1]

	i.log.Debug.Log("DMGs found: %v", dmgArray)

	return dmgArray, nil
}

// AddDmgPackages copies the contents of the given mounted DMG file into a folder
// of the same name located inside the dist directory. The folder is created it it does not exist.
//
// The folder will be the same name as the mounted DMG as displayed in the Volumes directory.
func (i *Installer) AddDmgPackages(volumePaths []string, pkgDirectory string) {
	cmd := "cp -r '%s' '%s'"

	for _, volumePath := range volumePaths {
		newCmd := fmt.Sprintf(cmd, volumePath, pkgDirectory)
		i.log.Info.Log("Copying files in path %s", volumePath)

		// no sudo unless you want root to own it (not tested)
		_, err := exec.Command("bash", "-c", newCmd).Output()
		if err != nil {
			i.log.Error.Log("Failed to copy contents of %s: %v", volumePath, err)
			continue
		}

		i.log.Info.Log("Successfully copied %s to %s", volumePath, pkgDirectory)
	}

	// error is ignored here as this is just debugging.
	distDir, err := os.ReadDir(pkgDirectory)
	if err != nil {
		i.log.Warn.Log("Failed to read %s: %v", pkgDirectory, err)
		return
	}

	i.log.Debug.Log("Distribution directory after adding DMG contents: %v", distDir)
}

// AttachDmgs takes an array of paths and attaches it to the disk via hdiutil.
//
// Upon successful completion, an array of paths to the mounted directory is returnei.
func (i *Installer) AttachDmgs(dmgPaths []string) []string {
	cmd := "hdiutil attach '%s'"

	volumePaths := make([]string, 0)

	for _, dmgPath := range dmgPaths {
		i.log.Debug.Log("DMG path: %s", dmgPath)

		if strings.Contains(dmgPath, ".dmg") {
			newCmd := fmt.Sprintf(cmd, dmgPath)

			i.log.Info.Log("Mounting %s", dmgPath)
			i.log.Debug.Log("Command: %s", newCmd)

			out, err := exec.Command("bash", "-c", newCmd).Output()
			if err != nil {
				i.log.Error.Log("failed to mount %s: %v", dmgPath, err)
				continue
			}

			i.log.Debug.Log("Command output: %s", strings.TrimSpace(string(out)))

			outArr := strings.Split(string(out), "\t")
			volumePath := strings.TrimSpace(outArr[len(outArr)-1])

			volumePaths = append(volumePaths, volumePath)
		}
	}

	i.log.Debug.Log("Mounted DMG volume paths: %v", volumePaths)

	return volumePaths
}

// DetachDmgs detaches the DMG from the Volumes directory.
// The paths are obtained from AttachDmgs.
func (i *Installer) DetachDmgs(volumePaths []string) {
	cmd := "hdiutil detach '%s'"

	for _, volumePath := range volumePaths {
		i.log.Info.Log("Unmounting %s", volumePath)
		i.log.Debug.Log("Mount: %s", volumePath)

		newCmd := fmt.Sprintf(cmd, volumePath)
		i.log.Debug.Log("Command: %s", newCmd)

		out, err := exec.Command("bash", "-c", newCmd).Output()
		if err != nil {
			i.log.Error.Log("Manual interaction needed, failed to unmount %s: %v", volumePath, err)
			continue
		}

		i.log.Debug.Log("Command output: %s", strings.TrimSpace(string(out)))
	}
}

// AddApp copies any files ending in the .app extension to the Applications directory.
//
// An error is returned if the disttribution directory is failed to be read from.
// Otherwise it will always return nil, any errors in the file moving is logged and skipped.
func (i *Installer) AddApp(distDirectory string) error {

	return nil
}

// isInstalled searches for the names of an installed package in the search directory.
//
// If an installed package name is found in the search directory, true is returned indicating
// the package being installed is already installed.
// Otherwise, false is returned if no installed arguments are given or it doesn't exist in the search
// directories.
func (p *Installer) IsInstalled(installedNames []string, pkgToInstall string) bool {
	// installedName is the user given installed file
	// installedFile is the installed file inside the directory files
	for _, installedName := range installedNames {
		// if the given name is blank then install regardless of check.
		if installedName == "" {
			return false
		}

		lowInstalledName := strings.ToLower(installedName)
		// unfortunately a nested loop is required here due to the array.
		// on the bright side it does exit out early if it finds a match.
		for _, installedFile := range p.searchDirectoryFiles {
			lowInstalledFile := strings.ToLower(installedFile)

			// NOTE: this is a fuzzy finder, so the exact names should be expected.
			// if a generic name is given, there is a good possibility the wrong name will be matched.
			// compares the files in the search directory, to the user defined package name for installation checks.
			if strings.Contains(lowInstalledFile, lowInstalledName) {
				p.log.Info.Log(fmt.Sprintf("Found existing installation for package %s", pkgToInstall))
				p.log.Debug.Log(fmt.Sprintf("Package: %s | Given package name: %s", pkgToInstall, installedName))
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
func (i *Installer) toLowerArray(arr *[]string) {
	for i, str := range *arr {
		(*arr)[i] = strings.ToLower(str)
	}
}
