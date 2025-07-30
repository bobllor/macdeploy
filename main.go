package main

import (
	"flag"
	"fmt"
	"macos-deployment/deploy_files/pkg"
	"macos-deployment/deploy_files/utils"
	"macos-deployment/deploy_files/yaml"
)

var config utils.Config = yaml.ReadYAML(utils.ConfigPath)

var installTeamViewer = flag.Bool("t", false, "Installs TeamViewer on the device.")
var adminStatus = flag.Bool("a", false, "Used to give Admin privileges to the user.")

func main() {
	flag.Parse()

	var packagesToInstall []string
	for key := range config.Packages {
		packagesToInstall = append(packagesToInstall, key)
	}

	pkg.InstallRosetta()

	packagesMap := pkg.MakePKG(packagesToInstall, *installTeamViewer)
	lol := pkg.IsInstalled("netdrive", []string{"/home/teboc/"})
	fmt.Println(lol)

	fmt.Println(packagesMap)

	pkg.InstallPKG("toki")

	//helpAccount := config.Accounts["help"]

	//fmt.Println(utils.BuildString(varNames["packages"], packages_regex))
	//fmt.Println(utils.BuildString(varNames["helpAccount"], helpAccount["full_name"]))
}
