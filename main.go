package main

import (
	"flag"
	"fmt"
	"macos-deployment/deploy_files/pkg"
	"macos-deployment/deploy_files/settings"
	"macos-deployment/deploy_files/vars"
)

var config settings.Settings = settings.ReadYAML(vars.ConfigPath)

// used for bash during eval
var varNames = map[string]string{
	"packages":    "regex",
	"helpAccount": "helpAccount",
}

var installTeamViewer = flag.Bool("t", false, "Installs TeamViewer on the device.")
var adminStatus = flag.Bool("a", false, "Used to give Admin privileges to the user.")

func main() {
	flag.Parse()

	packagesMap := pkg.MakePKG(config.Packages, *installTeamViewer)
	lol := pkg.IsInstalled("netdrive", []string{"/home/teboc/"})
	fmt.Println(lol)

	fmt.Println(packagesMap)
	fmt.Println(*adminStatus)

	pkg.InstallPKG("toki")

	//helpAccount := config.Accounts["help"]
	//packages_regex := "(" + strings.Join(packages, "|") + ")"

	//fmt.Println(utils.BuildString(varNames["packages"], packages_regex))
	//fmt.Println(utils.BuildString(varNames["helpAccount"], helpAccount["full_name"]))
}
