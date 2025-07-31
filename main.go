package main

import (
	"flag"
	"fmt"
	"macos-deployment/deploy_files/pkg"
	"macos-deployment/deploy_files/utils"
	"macos-deployment/deploy_files/yaml"
	"os/exec"
	"strings"
)

var config utils.Config = yaml.ReadYAML(utils.ConfigPath)

var installTeamViewer = flag.Bool("t", false, "Installs TeamViewer on the device.")
var adminStatus = flag.Bool("a", false, "Used to give Admin privileges to the user.")

func main() {
	flag.Parse()

	var searchDirFilesArr []map[string]bool
	for _, searchDir := range config.Search_Directories {
		searchMap, searchErr := utils.GetFileMap(searchDir)
		if searchErr != nil {
			fmt.Printf("[WARNING] Path %s does not exist, skipping path", searchDir)
			continue
		}

		searchDirFilesArr = append(searchDirFilesArr, searchMap)
	}

	packagesMap := pkg.MakePKG(config.Packages, *installTeamViewer)
	pkgInstallation(packagesMap, searchDirFilesArr)
}

func pkgInstallation(packagesMap map[string][]string, searchDirFilesArr []map[string]bool) {
	pkg.InstallRosetta()

	var findPKGScript string = utils.Home + "/macos-deployment/deploy_files/find_pkgs.sh"
	scriptOut, scriptErr := exec.Command("bash", findPKGScript, utils.PKGPath).Output()
	if scriptErr != nil {
		fmt.Printf("[DEBUG] script: %s | pkg folder: %s\n", findPKGScript, utils.PKGPath)
		println(string(scriptOut))
		panic(scriptErr)
	}

	foundPKGs := strings.Split(string(scriptOut), "\n")

	for pkge, pkgeArr := range packagesMap {
		isInstalled := pkg.IsInstalled(pkgeArr, searchDirFilesArr)
		if !isInstalled {
			pkg.InstallPKG(pkge, foundPKGs)
		}
	}
}
