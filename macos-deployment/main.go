package main

import (
	"flag"
	"fmt"
	"macos-deployment/pkg"
	"macos-deployment/settings"
	"macos-deployment/utils"
)

var config settings.Settings = settings.ReadYAML()

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

	fmt.Println(packagesMap)
	fmt.Println(*adminStatus)

	helpAccount := config.Accounts["helpAccount"]
	//packages_regex := "(" + strings.Join(packages, "|") + ")"

	//fmt.Println(utils.BuildString(varNames["packages"], packages_regex))
	fmt.Println(utils.BuildString(varNames["helpAccount"], helpAccount["fullName"]))
}
