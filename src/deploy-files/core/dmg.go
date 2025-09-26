package core

import (
	"fmt"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"os"
	"os/exec"
	"strings"
)

type Dmg struct {
	log    *logger.Log
	script *scripts.BashScripts
}

// NewDmg creates a new Dmg struct.
func NewDmg(log *logger.Log, scripts *scripts.BashScripts) *Dmg {
	dmg := Dmg{
		log:    log,
		script: scripts,
	}

	return &dmg
}

// ReadDmgDirectory reads the distribution directory for files with the .dmg extension,
// and returns an string array consisting the relative paths to the DMG file.
//
// It will return an error if there is an issue reading the directory.
func (d *Dmg) ReadDmgDirectory(dir string) ([]string, error) {
	out, err := exec.Command("bash", "-c", d.script.FindFiles, dir, "*.dmg").Output()
	if err != nil {
		return nil, err
	}

	dmgArray := strings.Split(string(out), "\n")
	// will return arr[:len-1] because an empty string is present at the end of the array
	dmgArray = dmgArray[:len(dmgArray)-1]

	d.log.Debug.Log("DMGs found: %v", dmgArray)

	return dmgArray, nil
}

// AddDmgPackages copies the contents of the given mounted DMG file into a folder
// located inside the dist directory. The folder is created it it does not exist.
//
// The folder will be the same name as the mounted DMG as displayed in the Volumes directory.
func (d *Dmg) AddDmgPackages(volumePaths []string, pkgDirectory string) {
	cmd := "cp -r '%s' '%s'"

	for _, volumePath := range volumePaths {
		newCmd := fmt.Sprintf(cmd, volumePath, pkgDirectory)
		d.log.Info.Log("Copying files in path %s", volumePath)

		// no sudo unless you want root to own it (not tested)
		_, err := exec.Command("bash", "-c", newCmd).Output()
		if err != nil {
			d.log.Error.Log("Failed to copy contents of %s: %v", volumePath, err)
			continue
		}

		d.log.Info.Log("Successfully copied %s to %s", volumePath, pkgDirectory)
	}

	// error is ignored here as this is just debugging.
	distDir, err := os.ReadDir(pkgDirectory)
	if err != nil {
		d.log.Warn.Log("Failed to read %s: %v", pkgDirectory, err)
		return
	}

	d.log.Debug.Log("Distribution directory after adding DMG contents: %v", distDir)
}

// AttachDmgs takes an array of paths and attaches it to the disk via hdiutil.
//
// Upon successful completion, an array of paths to the mounted directory is returned.
func (d *Dmg) AttachDmgs(dmgPaths []string) []string {
	cmd := "hdiutil attach '%s'"

	volumePaths := make([]string, 0)

	for _, dmgPath := range dmgPaths {
		d.log.Debug.Log("DMG path: %s", dmgPath)

		if strings.Contains(dmgPath, ".dmg") {
			newCmd := fmt.Sprintf(cmd, dmgPath)

			d.log.Info.Log("Mounting %s", dmgPath)
			d.log.Debug.Log("Command: %s", newCmd)

			out, err := exec.Command("bash", "-c", newCmd).Output()
			if err != nil {
				d.log.Error.Log("failed to mount %s: %v", dmgPath, err)
				continue
			}

			d.log.Debug.Log("Command output: %s", strings.TrimSpace(string(out)))

			outArr := strings.Split(string(out), "\t")
			volumePath := strings.TrimSpace(outArr[len(outArr)-1])

			volumePaths = append(volumePaths, volumePath)
		}
	}

	d.log.Debug.Log("Mounted DMG volume paths: %v", volumePaths)

	return volumePaths
}

// DetachDmgs detaches the DMG from the Volumes directory.
// The paths are obtained from AttachDmgs.
func (d *Dmg) DetachDmgs(volumePaths []string) {
	cmd := "hdiutil detach '%s'"

	for _, volumePath := range volumePaths {
		d.log.Info.Log("Unmounting %s", volumePath)
		d.log.Debug.Log("Mount: %s", volumePath)

		newCmd := fmt.Sprintf(cmd, volumePath)
		d.log.Debug.Log("Command: %s", newCmd)

		out, err := exec.Command("bash", "-c", newCmd).Output()
		if err != nil {
			d.log.Error.Log("Manual interaction needed, failed to unmount %s: %v", volumePath, err)
			continue
		}

		d.log.Debug.Log("Command output: %s", strings.TrimSpace(string(out)))
	}
}
