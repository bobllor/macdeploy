package main

import (
	"flag"
	"fmt"
	"macos-deployment/deploy-files/core"
	"macos-deployment/deploy-files/logger"
	"macos-deployment/deploy-files/scripts"
	"macos-deployment/deploy-files/utils"
	"macos-deployment/deploy-files/yaml"
	"os/exec"
	"strings"
)

// NOTE: when zipping the files for the HTTP server only include the .sh files in deploy-files

var configPath string = "./config.yaml"
var config utils.Config = yaml.ReadYAML(configPath)

var installTeamViewer = flag.Bool("t", false, "Installs TeamViewer on the device.")
var adminStatus = flag.Bool("a", false, "Used to give Admin privileges to the user.")

func main() {
	flag.Parse()
	initLog()

	var accounts map[string]utils.User = config.Accounts
	accountCreation(accounts)

	var searchDirFilesArr []map[string]bool
	for _, searchDir := range config.Search_Directories {
		searchMap, searchErr := utils.GetFileMap(searchDir)
		if searchErr != nil {
			msg := fmt.Sprintf("Path %s does not exist, skipping path", searchDir)
			logger.Log(msg, 4)
			continue
		}

		searchDirFilesArr = append(searchDirFilesArr, searchMap)
	}

	if len(searchDirFilesArr) > 0 {
		packagesMap := core.MakePKG(config.Packages, *installTeamViewer)
		pkgInstallation(packagesMap, searchDirFilesArr)
	}

	if config.File_Vault {
		startFileVault()
	}

	if config.Firewall {
		startFirewall()
	}
}

// initLog initializes the serial tag (if exists) and the logger
func initLog() {
	tag, tagErr := utils.GetSerialTag()
	if tagErr != nil {
		tag = "UNKNOWN"
	}

	utils.SerialTag = tag

	logger.NewLog(utils.SerialTag)
}

func accountCreation(accounts map[string]utils.User) {
	for key := range accounts {
		currAccount := accounts[key]

		core.CreateAccount(currAccount, *adminStatus)
	}
}

func pkgInstallation(packagesMap map[string][]string, searchDirFilesArr []map[string]bool) {
	roseErr := core.InstallRosetta()
	if roseErr != nil {
		logger.Log("Failed to install Rosetta", 3)
		return
	}

	scriptOut, scriptErr := exec.Command("bash", scripts.FindPackagesScript, utils.PKGPath).Output()

	debug := fmt.Sprintf("PKG folder: %s", utils.PKGPath)
	logger.Log(debug, 7)

	if scriptErr != nil {
		scriptErrMsg := "Failed to locate package folder"
		logger.Log(scriptErrMsg, 3)
		return
	}

	foundPKGs := strings.Split(string(scriptOut), "\n")

	for pkge, pkgeArr := range packagesMap {
		isInstalled := core.IsInstalled(pkgeArr, searchDirFilesArr)
		if !isInstalled {
			core.InstallPKG(pkge, foundPKGs)
		}
	}
}

func startFileVault() {
	core.EnableFileVault()
}

func startFirewall() {
	core.EnableFireWall()
}
