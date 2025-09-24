package tests

import (
	"fmt"
	"macos-deployment/deploy-files/core"
	"macos-deployment/deploy-files/logger"
	"math/rand"
	"strings"
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

func TestRemovePackages(t *testing.T) {
	packager := getPackager(t)
	expectedLength := len(append(packagesToAdd, packager.GetPackages()...)) - 1

	packager.AddPackages(packagesToAdd)

	randomSelection := strings.ToLower(packagesToAdd[rand.Intn(len(packagesToAdd))])

	packager.RemovePackages([]string{randomSelection})

	newLen := len(packager.GetPackages())

	if newLen != expectedLength {
		t.Errorf("failed to remove package, got %d instead of %d", newLen, expectedLength)
	}

	fmt.Println(packager.GetPackages())
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
