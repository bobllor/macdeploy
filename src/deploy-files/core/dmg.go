package core

import (
	"fmt"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"os/exec"
	"strings"
)

type Dmg struct {
	log    *logger.Log
	script *scripts.BashScripts
}

func NewDmg(log *logger.Log, scripts *scripts.BashScripts) *Dmg {
	dmg := Dmg{
		log:    log,
		script: scripts,
	}

	return &dmg
}

func (d *Dmg) ReadDmgDirectory(dir string) ([]string, error) {
	out, err := exec.Command("bash", "-c", d.script.FindFiles, dir, "*.dmg").Output()
	if err != nil {
		return nil, err
	}

	// will return arr[:-1] because an empty string is present at the end of the array
	dmgArray := strings.Split(string(out), "\n")

	d.log.Debug.Log("DMGs found: %v", dmgArray)

	return dmgArray[:len(dmgArray)-1], nil
}

// MountDmgs takes an array of paths and mounts it to the disk via hdiutil.
//
// Upon successful completion, an array of paths to the mounted directory is returned.
func (d *Dmg) MountDmgs(dmgPaths []string) {
	cmd := "hdiutil attach '%s'"

	mountDirectories := make([]string, 0)

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
			fmt.Println(strings.Split(string(out), " "))
		}
	}

	d.log.Debug.Log("Mounted directories: %v", mountDirectories)
}

func (d *Dmg) UnmountDmgs(mountDir string, dmgPaths []string) {
	cmd := "hdiutil detach '%s'"

	for _, dmgPath := range dmgPaths {
		if strings.Contains(dmgPath, ".dmg") {
			dmgSplit := strings.Split(dmgPath, "/")
			dmgName := strings.ReplaceAll(dmgSplit[len(dmgSplit)-1], ".dmg", "")

			dmgMount := fmt.Sprintf("%s/%s", mountDir, dmgName)

			d.log.Debug.Log("Mount: %s | DMG name: %s | DMG path: %s", dmgMount, dmgName, dmgPath)

			newCmd := fmt.Sprintf(cmd, dmgMount)
			out, err := exec.Command("bash", "-c", newCmd).Output()
			if err != nil {
				d.log.Error.Log("Failed to unmount %s: %v", dmgMount, err)
			}

			d.log.Debug.Log("Command output: %s", strings.TrimSpace(string(out)))
		}
	}
}
