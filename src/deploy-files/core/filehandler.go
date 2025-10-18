package core

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"macos-deployment/deploy-files/logger"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type FileHandler struct {
	packagesToInstall map[string][]string
	log               *logger.Log
	scriptsPathCache  map[string]string // Cache for script paths, k:v <file name>:<file path>. The key is lowercase.
}

// NewFileHandler creates a new Installer for package and app installations for macOS.
func NewFileHandler(packagesToInstall map[string][]string, logger *logger.Log) *FileHandler {
	packagesLowered := make(map[string][]string)

	for pkg, appNames := range packagesToInstall {
		lowPkg := strings.ToLower(pkg)

		packagesLowered[lowPkg] = appNames
	}

	handler := FileHandler{
		packagesToInstall: packagesLowered,
		log:               logger,
		scriptsPathCache:  make(map[string]string),
	}

	return &handler
}

// InstallRosetta installs the Rosetta software required for installing packages.
// If Rosetta is already installed or the CPU is non-arm64, then nil will be returned.
//
// If the installation of Rosetta fails then an error will be returned.
func (f *FileHandler) InstallRosetta() error {
	if runtime.GOARCH == "amd64" {
		f.log.Info.Log("Skipping Rosetta installation, architecture: %s", runtime.GOARCH)
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

		f.log.Info.Log("Rosetta successfully installed")
	} else {
		f.log.Warn.Log("Found existing Rosetta installation")
	}

	return nil
}

// AddPKG adds new packages to the list of packages to install by adding packages
// included in the --include flag. All package names are lowered.
func (f *FileHandler) AddPackages(packagesToAdd []string) {
	for _, includedPkg := range packagesToAdd {
		// used to extract the package and its installed files from the flag argument.
		includeArgArr := strings.Split(includedPkg, "/")

		f.toLowerArray(&includeArgArr)

		pkg := includeArgArr[0]

		// the installed files from the pkg, f.e. the installed package name.
		pkgInstalledArr := make([]string, 0)
		if len(includeArgArr) > 1 {
			pkgInstalledArr = includeArgArr[1:]
		}

		f.packagesToInstall[pkg] = pkgInstalledArr
		f.log.Info.Log("Added %s to the installation list", pkg)
	}

	f.log.Debug.Log("Packages: %v", f.packagesToInstall)
}

// RemovePackages removes packages from the list of packages to install by removing the packages
// given in the --exclude flag.
func (f *FileHandler) RemovePackages(packagesToRemove []string) {
	for _, excludedPkg := range packagesToRemove {
		excludedPkgLow := strings.ToLower(excludedPkg)
		_, ok := f.packagesToInstall[excludedPkgLow]
		if ok {
			f.log.Info.Log("Removed %s from installation list", excludedPkg)
			delete(f.packagesToInstall, excludedPkg)
		}
	}
}

// ReadDir reads the directory and recursively matches the files to the search pattern.
// The output array contains the full path to the files.
//
// If it fails to read the directory then it returns an error.
func (f *FileHandler) ReadDir(directoryPath string, searchPattern string) ([]string, error) {
	f.log.Debug.Log("Directory: %s | Search pattern: %s", directoryPath, searchPattern)

	files := make([]string, 0)
	walk := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		fullPath := fmt.Sprintf("%s/%s", directoryPath, path)

		paths := strings.Split(fullPath, "/")
		basename := paths[len(paths)-1]

		if strings.Contains(basename, searchPattern) {
			files = append(files, fullPath)
		}
		return nil
	}
	err := fs.WalkDir(os.DirFS(directoryPath), ".", walk)
	if err != nil {
		return nil, err
	}

	f.log.Debug.Log("%s Files: %v", searchPattern, files)

	return files, nil
}

// InstallPackages installs the keys of the map of packages to install.
// The argument takes an array of relative paths read from the project package directory.
//
// If a package fails to install, then it will be logged and skipped.
func (f *FileHandler) InstallPackages(packagesPath []string, searchDirectoryFiles []string) {
	for pkg, installedNames := range f.packagesToInstall {
		isInstalled := f.IsInstalled(installedNames, pkg, searchDirectoryFiles)

		if isInstalled {
			f.log.Warn.Log(fmt.Sprintf("Package %s is already installed", pkg))
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
				f.log.Info.Log("Installing package %s", pkg)

				cmd := fmt.Sprintf(`installer -pkg "%s" -target /`, file)
				f.log.Debug.Log("Package: %s | Package path: %s | Command: %s", pkg, file, cmd)

				out, err := exec.Command("sudo", "bash", "-c", cmd).Output()
				if err != nil {
					outStr := strings.TrimSpace(string(out))
					f.log.Warn.Log(fmt.Sprintf("Failed to install %s: %s %v", pkg, outStr, err))
					failedInstall = true
					break
				}

				outMsg := "Successfully installed"
				if !strings.Contains(pkgLowered, ".pkg") {
					outMsg = fmt.Sprintf("%s %s.pkg", outMsg, pkg)
				} else {
					outMsg = fmt.Sprintf("%s %s", outMsg, pkg)
				}
				f.log.Info.Log(outMsg)

				successfulInstall = true
				break
			}
		}

		if !successfulInstall && !failedInstall {
			f.log.Warn.Log("Unable to find package %s", pkg)
		}
	}
}

// GetPackages returns the packages that are being installed.
func (f *FileHandler) GetPackages() []string {
	packages := make([]string, 0, len(f.packagesToInstall))

	for pkg := range f.packagesToInstall {
		packages = append(packages, pkg)
	}

	return packages
}

// GetAllPackages gets the packages to be installed and its installed file names array.
func (f *FileHandler) GetAllPackages() map[string][]string {
	return f.packagesToInstall
}

// AddDmgPackages copies the contents of the given mounted DMG file into a folder
// of the same name located inside the dist directory. The folder is created it it does not exist.
//
// The folder will be the same name as the mounted DMG as displayed in the Volumes directory.
func (f *FileHandler) AddDmgPackages(volumePaths []string, pkgDirectory string) {
	cmd := "cp -r '%s' '%s'"

	for _, volumePath := range volumePaths {
		newCmd := fmt.Sprintf(cmd, volumePath, pkgDirectory)
		f.log.Info.Log("Copying files in path %s", volumePath)

		// no sudo unless you want root to own it (not tested)
		_, err := exec.Command("bash", "-c", newCmd).Output()
		if err != nil {
			f.log.Error.Log("Failed to copy contents of %s: %v", volumePath, err)
			continue
		}

		f.log.Info.Log("Successfully copied %s to %s", volumePath, pkgDirectory)
	}

	// error is ignored here as this is just debugging.
	distDir, err := os.ReadDir(pkgDirectory)
	if err != nil {
		f.log.Warn.Log("Failed to read %s: %v", pkgDirectory, err)
		return
	}

	f.log.Debug.Log("Distribution directory after adding DMG contents: %v", distDir)
}

// AttachDmgs takes an array of paths and attaches it to the disk via hdiutil.
//
// Upon successful completion, an array of paths to the mount inside the Volumes
// Volumes directory is returned.
func (f *FileHandler) AttachDmgs(dmgPaths []string) []string {
	cmd := "hdiutil attach '%s'"

	volumePaths := make([]string, 0)

	for _, dmgPath := range dmgPaths {
		f.log.Debug.Log("DMG path: %s", dmgPath)

		if strings.Contains(dmgPath, ".dmg") {
			newCmd := fmt.Sprintf(cmd, dmgPath)

			f.log.Info.Log("Mounting %s", dmgPath)
			f.log.Debug.Log("Command: %s", newCmd)

			out, err := exec.Command("bash", "-c", newCmd).Output()
			if err != nil {
				f.log.Error.Log("failed to mount %s: %v", dmgPath, err)
				continue
			}

			f.log.Debug.Log("Command output: %s", strings.TrimSpace(string(out)))

			outArr := strings.Split(string(out), "\t")
			volumePath := strings.TrimSpace(outArr[len(outArr)-1])

			volumePaths = append(volumePaths, volumePath)
		}
	}

	f.log.Debug.Log("Mounted DMG volume paths: %v", volumePaths)

	return volumePaths
}

// ExecuteScripts runs shell scripts on the device.
// This requires the script file name and an array of paths containing shell scripts.
//
// It returns a string and an error, depending on the exit status of the script.
// If the script did not get executed, an error is returned.
func (f *FileHandler) ExecuteScript(scriptName string, scriptPaths []string) (string, error) {
	f.log.Info.Log("Running %s", scriptName)
	ogName := scriptName
	scriptName = strings.TrimSpace(strings.ToLower(scriptName))

	// cache is built inside the loop
	if _, ok := f.scriptsPathCache[scriptName]; ok {
		f.log.Info.Log("Found %s in cache", ogName)

		scriptPath := f.scriptsPathCache[scriptName]
		outMsg, err := f.execute(scriptPath)
		outMsg = strings.TrimSpace(outMsg)
		if err != nil {
			return outMsg, err
		}

		return outMsg, nil
	}

	for _, scriptPath := range scriptPaths {
		scriptPathLow := strings.ToLower(scriptPath)

		if strings.Contains(scriptPathLow, scriptName) {
			outMsg, err := f.execute(scriptPath)
			outMsg = strings.TrimSpace(outMsg)
			if err != nil {
				return outMsg, err
			}

			return outMsg, nil
		} else {
			// only the path's case should be left alone for execution
			filename := strings.ToLower(filepath.Base(scriptPath))

			f.scriptsPathCache[filename] = scriptPath
			f.log.Debug.Log("Added %s to cache", filename)
		}
	}

	return "", errors.New("failed to find script")
}

// execute executes the given script path.
//
// It returns the output of the script and an error, if one occurred.
func (f *FileHandler) execute(scriptPath string) (string, error) {
	// NOTE: if the user exits non-zero on their script, this will fail.
	out, err := exec.Command("bash", "-c", scriptPath).Output()
	outMsg := strings.TrimSpace(string(out))
	if err != nil {
		return outMsg, err
	}

	return outMsg, nil
}

// GetScriptCache returns the map of the script cache.
func (f *FileHandler) GetScriptCache() map[string]string {
	return f.scriptsPathCache
}

// DetachDmgs detaches the DMG from the Volumes directory.
// The paths are obtained from AttachDmgs.
func (f *FileHandler) DetachDmgs(volumePaths []string) {
	cmd := "hdiutil detach '%s'"

	for _, volumePath := range volumePaths {
		f.log.Info.Log("Unmounting %s", volumePath)
		f.log.Debug.Log("Mount: %s", volumePath)

		newCmd := fmt.Sprintf(cmd, volumePath)
		f.log.Debug.Log("Command: %s", newCmd)

		out, err := exec.Command("bash", "-c", newCmd).Output()
		if err != nil {
			f.log.Error.Log("Manual interaction needed, failed to unmount %s: %v", volumePath, err)
			continue
		}

		f.log.Debug.Log("Command output: %s", strings.TrimSpace(string(out)))
	}
}

// CopyFiles recursively copies an array of directory paths to a target directory.
//
// Errors during the copy operation are logged and skipped, requiring manual intervention.
func (f *FileHandler) CopyFiles(paths []string, target string) {
	f.log.Debug.Log("File paths: %v", paths)
	f.log.Info.Log("Copying to %s", target)
	// lowercase not needed as it is obtained from ReadDir
	// case sensitivity doesn't matter on mac anyways (at least by default in sequoia+)
	for _, path := range paths {
		file, err := os.Stat(path)
		if err != nil {
			f.log.Error.Log("Failed to stat %s: %v", path, err)
			continue
		}

		targetFile := fmt.Sprintf("%s/%s", target, file.Name())

		f.log.Debug.Log("Target file: %s", targetFile)

		if file.IsDir() {
			// copyFS already creates the directories if missing
			err = os.CopyFS(targetFile, os.DirFS(path))
			if err != nil {
				f.log.Error.Log("Failed to copy %s: %v", path, err)
				continue
			}
		} else {
			fmt.Println(file.Size())
			reader, err := os.Open(path)
			if err != nil {
				f.log.Error.Log("Failed to read %s: %v", path, err)
				continue
			}

			newFile, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY, file.Mode())
			if err != nil {
				f.log.Error.Log("Failed to write %s: %v", targetFile, err)
				continue
			}

			_, err = io.Copy(newFile, reader)
			if err != nil {
				f.log.Error.Log("Failed to write %s: %v", targetFile, err)
				continue
			}

			reader.Close()
			newFile.Close()
		}

		f.log.Info.Log("Copied %s to %s", path, target)
	}
}

// IsInstalled searches for the names of an installed package in the search directory.
//
// If an installed package name is found in the search directory, true is returned indicating
// the package being installed is already installed.
// Otherwise, false is returned if no installed arguments are given or it doesn't exist in the search
// directories.
func (p *FileHandler) IsInstalled(installedNames []string, pkgToInstall string, searchDirectoryFiles []string) bool {
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
		for _, installedFile := range searchDirectoryFiles {
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
func (f *FileHandler) toLowerArray(arr *[]string) {
	for i, str := range *arr {
		(*arr)[i] = strings.ToLower(str)
	}
}
