package tests

import (
	"macos-deployment/deploy-files/core"
	"macos-deployment/deploy-files/logger"
	"testing"
)

var packagesToAdd = []string{
	"some thing.pkg",
	"AnoTherCaseSenSITIVE.pkg",
	"no pkg here",
}

var packagesToInstall = map[string][]string{
	"TeamViewer.pkg":   {"teamviewer"},
	"Test packAGe.PkG": {"lol"},
}

var baseLenPkgInstall int = len(packagesToInstall)

func TestArrayCase(t *testing.T) {
	packager := getPackager(t)

	packager.AddPackages(packagesToAdd)
}

func TestAddPackages(t *testing.T) {
	packager := getPackager(t)

	packager.AddPackages(packagesToAdd)

	packages := packager.GetPackages()
	newLen := len(packages)

	if newLen != baseLenPkgInstall+len(packagesToAdd) {
		t.Errorf(
			"starting length: %d does not match ending length of add packages: %d",
			baseLenPkgInstall, newLen,
		)
	}
}

func getLogger(t *testing.T) *logger.Log {
	logDir := t.TempDir() + "/log"
	verbose := false

	logger := logger.NewLog("lol123", logDir, verbose)

	return logger
}

func getPackager(t *testing.T) *core.Packager {
	log := getLogger(t)

	return core.NewPackager(packagesToInstall, []string{}, log)
}
