package core

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bobllor/macdeploy/src/deploy-files/logger"
)

type FileHandler struct {
	packagesToInstall map[string][]string
	log               *logger.Logger
	scriptsPathCache  map[string]string // Cache for script paths, k:v <file name>:<file path>. The key is lowercase.
}

// NewFileHandler creates a new FileHandler to handle package installations.
func NewFileHandler(logger *logger.Logger) *FileHandler {
	handler := FileHandler{
		packagesToInstall: make(map[string][]string),
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
		f.log.Info(fmt.Sprintf("Skipping Rosetta installation, architecture: %s", runtime.GOARCH))
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

		f.log.Info("Rosetta successfully installed")
	} else {
		f.log.Warn("Found existing Rosetta installation")
	}

	return nil
}

// AddPackages adds new packages to the file handler. All package names
// will be lowered.
//
// packagesToAdd is a slice of strings containing the .pkg file name to install.
// The contents can be a string of the .pkg file name, the full .pkg file name,
// or a string with ',' delimiters that represents the structure: "<file>,<install files>,<install files>".
// The delimiters are used to indicate its installation files used to prevent reinstalls.
func (f *FileHandler) AddPackages(packagesToAdd []string) {
	for _, includedPkg := range packagesToAdd {
		// used to extract the package and its installed files from the flag argument.
		includeArgArr := strings.Split(includedPkg, ",")

		f.toLowerArray(&includeArgArr)

		pkg := includeArgArr[0]

		// the installed files from the pkg, f.e. the installed package name.
		pkgInstalledArr := make([]string, 0)
		// checks if there are any empty strings
		if len(includeArgArr) > 1 {
			for _, v := range includeArgArr[1:] {
				if strings.TrimSpace(v) != "" {
					pkgInstalledArr = append(pkgInstalledArr, v)
				}
			}
		}

		f.packagesToInstall[pkg] = pkgInstalledArr
		f.log.Info(fmt.Sprintf("Added %s to the installation list", pkg))
	}
}

// AddMapPackages adds new packages to the file handler using a map. All package names
// will be lowered.
//
// packagesToAdd is a map of a string with a slice of strings.
// The key represents the package to install, while its slice values are
// the installation files of the package.
func (f *FileHandler) AddMapPackages(packagesToAdd map[string][]string) {
	for key, val := range packagesToAdd {
		key = strings.ToLower(key)

		f.packagesToInstall[key] = val
	}
}

// RemovePackages removes packages from the list of packages to install by removing the packages
// in the package maps from a slice of packages.
//
// The names must match in order to remove the packages.
func (f *FileHandler) RemovePackages(packagesToRemove []string) {
	for _, excludedPkg := range packagesToRemove {
		excludedPkgLow := strings.ToLower(excludedPkg)
		_, ok := f.packagesToInstall[excludedPkgLow]
		if ok {
			f.log.Info(fmt.Sprintf("Removed %s from installation list", excludedPkg))
			delete(f.packagesToInstall, excludedPkg)
		}
	}
}

// ReadDir reads the directory and recursively matches the files to a search pattern.
// The output array contains the full path to the files from the relative path of the
// program.
//
// If it fails to read the directory then it returns an error.
func (f *FileHandler) ReadDir(directoryPath string, searchPattern string) ([]string, error) {
	f.log.Debug(fmt.Sprintf("Directory: %s | Search pattern: %s", directoryPath, searchPattern))

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
		err = fmt.Errorf("Directory path '%s' is not found", directoryPath)
		return nil, err
	}

	f.log.Debug(fmt.Sprintf("%s Files: %v", searchPattern, files))
	return files, nil
}

// InstallPackages installs the keys of the map of packages to install.
// It will return a number of packages that were successfully installed. If a package failed to install,
// then this will be skipped and logged.
//
// packagesPath is a slice of paths of the .pkg file.
//
// installDirectoryFiles is a slice of file paths that represent the installed .pkg file. The elements are
// used to check if the file is already installed before attempting an install.
func (f *FileHandler) InstallPackages(packagesPath []string, installDirectoryFiles []string) int {
	installedFiles := 0
	if len(f.packagesToInstall) == 0 {
		f.log.Warn("No packages to install")
		fmt.Println("No packages to install")

		return installedFiles
	}

	for pkg, installedNames := range f.packagesToInstall {
		isInstalled := f.IsInstalled(installedNames, installDirectoryFiles)

		if isInstalled {
			f.log.Info(fmt.Sprintf("Found existing installation for package %s", pkg))
			f.log.Debug(fmt.Sprintf("Package: %s | Given package name: %s", pkg, installedNames))
			fmt.Printf("%s is already installed\n", pkg)

			installedFiles += 1
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

			// this cannot be hard coded with the .pkg file, this allows for
			// dynamic handling of long names (due to an edge case).
			if strings.Contains(relativePkgLow, pkgLowered) {
				f.log.Info(fmt.Sprintf("Installing package %s", pkg))
				fmt.Printf("Starting installation for %s\n", pkg)

				cmd := fmt.Sprintf(`installer -pkg "%s" -target /`, file)
				f.log.Debug(fmt.Sprintf("Package: %s | Package path: %s | Command: %s", pkg, file, cmd))

				out, err := exec.Command("sudo", "bash", "-c", cmd).Output()
				if err != nil {
					outStr := strings.TrimSpace(string(out))
					f.log.Warn(fmt.Sprintf("Failed installation of %s: %s %v", pkg, outStr, err))
					fmt.Printf("Failed to install %s", pkg)
					failedInstall = true
					break
				}

				outMsg := "Successfully installed"
				if !strings.HasSuffix(pkgLowered, ".pkg") {
					outMsg = fmt.Sprintf("%s %s.pkg", outMsg, pkg)
				} else {
					outMsg = fmt.Sprintf("%s %s", outMsg, pkg)
				}
				f.log.Info(outMsg)

				successfulInstall = true
				installedFiles += 1
				fmt.Printf("Installed %s\n", pkg)
				break
			}
		}

		if !successfulInstall && !failedInstall {
			fullFailMsg := fmt.Sprintf("Unable to find package %s to install", pkg)
			f.log.Warn(fullFailMsg)
			fmt.Println(fullFailMsg)
		}
	}

	return installedFiles
}

// GetPackages returns the packages that are being installed.
// This does not include the installed files.
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
		f.log.Info(fmt.Sprintf("Copying files in path %s", volumePath))

		// no sudo unless you want root to own it (not tested)
		_, err := exec.Command("bash", "-c", newCmd).Output()
		if err != nil {
			f.log.Warn(fmt.Sprintf("Failed to copy contents of %s: %v", volumePath, err))
			continue
		}

		f.log.Info(fmt.Sprintf("Successfully copied %s to %s", volumePath, pkgDirectory))
	}

	// error is ignored here as this is just debugging.
	distDir, err := os.ReadDir(pkgDirectory)
	if err != nil {
		f.log.Warn(fmt.Sprintf("Failed to read %s: %v", pkgDirectory, err))
		return
	}

	f.log.Debug(fmt.Sprintf("Distribution directory after adding DMG contents: %v", distDir))
}

// AttachDmgs takes an array of paths and attaches it to the disk via hdiutil.
//
// Upon successful completion, an array of paths to the mount inside the Volumes
// directory is returned.
func (f *FileHandler) AttachDmgs(dmgPaths []string) []string {
	cmd := "hdiutil attach '%s'"

	volumePaths := make([]string, 0)

	for _, dmgPath := range dmgPaths {
		f.log.Debug(fmt.Sprintf("DMG path: %s", dmgPath))

		if strings.Contains(dmgPath, ".dmg") {
			newCmd := fmt.Sprintf(cmd, dmgPath)

			f.log.Info(fmt.Sprintf("Mounting %s", dmgPath))
			f.log.Debug(fmt.Sprintf("Command: %s", newCmd))

			out, err := exec.Command("bash", "-c", newCmd).Output()
			if err != nil {
				f.log.Warn(fmt.Sprintf("Failed to mount %s: %v", dmgPath, err))
				continue
			}

			f.log.Debug(fmt.Sprintf("Command output: %s", strings.TrimSpace(string(out))))

			outArr := strings.Split(string(out), "\t")
			volumePath := strings.TrimSpace(outArr[len(outArr)-1])

			volumePaths = append(volumePaths, volumePath)
		}
	}

	f.log.Debug(fmt.Sprintf("Mounted DMG volume paths: %v", volumePaths))

	return volumePaths
}

// ExecuteScripts runs shell scripts on the device.
// This requires the script file name and an array of paths containing shell scripts.
//
// It returns a string and an error, depending on the exit status of the script.
// If the script did not get executed, an error is returned.
func (f *FileHandler) ExecuteScript(scriptName string, scriptPaths []string) (string, error) {
	f.log.Infof("Starting script execution for %s", scriptName)

	ogName := scriptName // only used for logging
	scriptName = strings.TrimSpace(strings.ToLower(scriptName))

	for _, scriptPath := range scriptPaths {
		scriptPathLow := strings.TrimSpace(strings.ToLower(scriptPath))

		// only the path's case should be left alone for execution
		filename := strings.ToLower(filepath.Base(scriptPath))

		// cache is used to rerun scripts in case the same script is reused
		if _, ok := f.scriptsPathCache[scriptName]; ok {
			f.log.Info(fmt.Sprintf("Found %s in cache", ogName))

			scriptPath := f.scriptsPathCache[scriptName]
			outMsg, err := f.execute(scriptPath)
			if err != nil {
				return outMsg, err
			}

			return outMsg, nil
		} else {
			f.scriptsPathCache[filename] = scriptPath
			f.log.Debug(fmt.Sprintf("Added %s to cache", filename))
		}

		// substring match
		if strings.Contains(scriptPathLow, scriptName) {
			outMsg, err := f.execute(scriptPath)

			f.log.Debugf("Script %s output: %s, error: %v", ogName, outMsg, err)
			if err != nil {
				return outMsg, err
			}

			return outMsg, nil
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
		f.log.Info(fmt.Sprintf("Unmounting %s", volumePath))
		f.log.Debug(fmt.Sprintf("Mount: %s", volumePath))

		newCmd := fmt.Sprintf(cmd, volumePath)
		f.log.Debug(fmt.Sprintf("Command: %s", newCmd))

		out, err := exec.Command("bash", "-c", newCmd).Output()
		if err != nil {
			f.log.Warn(fmt.Sprintf("Manual interaction needed, failed to unmount %s: %v", volumePath, err))
			continue
		}

		f.log.Debug(fmt.Sprintf("Command output: %s", strings.TrimSpace(string(out))))
	}
}

// CopyFiles recursively copies an array of directory paths to a target directory.
//
// Errors during the copy operation are logged and skipped, requiring manual intervention.
func (f *FileHandler) CopyFiles(paths []string, target string) {
	f.log.Info(fmt.Sprintf("Copying %d paths to %s", len(paths), target))
	f.log.Debug(fmt.Sprintf("File paths: %v", paths))
	// lowercase not needed as it is obtained from ReadDir
	// case sensitivity doesn't matter on mac anyways (at least by default in sequoia+)
	for _, path := range paths {
		file, err := os.Stat(path)
		if err != nil {
			f.log.Warn(fmt.Sprintf("Failed to stat %s: %v", path, err))
			continue
		}

		targetFile := fmt.Sprintf("%s/%s", target, file.Name())

		f.log.Debug(fmt.Sprintf("Target file: %s", targetFile))

		if file.IsDir() {
			// copyFS already creates the directories if missing
			err = os.CopyFS(targetFile, os.DirFS(path))
			if err != nil {
				f.log.Warn(fmt.Sprintf("Failed to copy %s: %v", path, err))
				continue
			}
		} else {
			fmt.Println(file.Size())
			reader, err := os.Open(path)
			if err != nil {
				f.log.Warn(fmt.Sprintf("Failed to read %s: %v", path, err))
				continue
			}

			newFile, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY, file.Mode())
			if err != nil {
				f.log.Warn(fmt.Sprintf("Failed to write %s: %v", targetFile, err))
				continue
			}

			_, err = io.Copy(newFile, reader)
			if err != nil {
				f.log.Warnf("Failed to write %s: %v", targetFile, err)
				continue
			}

			reader.Close()
			newFile.Close()
		}

		f.log.Infof("Copied %s to %s", path, target)
	}
}

// IsInstalled searches for the names of an installed package in the search directory.
//
// If an installed package name is found in the search directory, true is returned indicating
// the package being installed is already installed.
// Otherwise, false is returned if no installed arguments are given or it doesn't exist in the search
// directories.
func (p *FileHandler) IsInstalled(installedNames []string, searchDirectoryFiles []string) bool {
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

// PackageString returns a string representation of the packages to install
// and any installation file names for the package.
func (p *FileHandler) PackageString() string {
	strSlice := []string{}

	for key, val := range p.packagesToInstall {
		installationFiles := "No installed files given"
		if len(val) > 0 {
			installationFiles = strings.Join(val, ",")
		}

		str := fmt.Sprintf("%s+%s", key, installationFiles)

		strSlice = append(strSlice, str)
	}

	out := "Install packages: " + strings.Join(strSlice, "|")

	return out
}
