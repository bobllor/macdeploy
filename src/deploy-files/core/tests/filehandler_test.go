package core

import (
	"bytes"
	"fmt"
	"log"
	"maps"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bobllor/macdeploy/src/deploy-files/core"
	"github.com/bobllor/macdeploy/src/deploy-files/logger"
	"github.com/bobllor/macdeploy/src/tests"
)

var packagesToAdd []string = []string{
	"some thing.pkg",
	"AnoTherCaseSenSITIVE.pkg",
}

var packagesToInstall = map[string][]string{
	"TeamViewer.pkg":   {"teamviewer"},
	"Test packAGe.PkG": {"test.package"},
}

var searchDirectoryFiles = []string{
	"teamviewer", "test.package",
	"example 1", "example 2",
}

var baseLenPkgInstall int = len(packagesToAdd)

func TestArrayLowerCase(t *testing.T) {
	handler := core.NewFileHandler(tests.TestLogger)

	handler.AddPackages(packagesToAdd)

	loweredPackages := make(map[string]struct{}, 0)

	for pkg := range packagesToInstall {
		pkgLow := strings.ToLower(pkg)

		loweredPackages[pkgLow] = struct{}{}
	}
	for _, pkg := range packagesToAdd {
		pkgLow := strings.ToLower(pkg)

		loweredPackages[pkgLow] = struct{}{}
	}

	// should already be lowered here during the constructor and add packages.
	for _, pkg := range handler.GetPackages() {
		if _, ok := loweredPackages[pkg]; !ok {
			t.Errorf("value %s does not exist", pkg)
		}
	}
}

func TestAddPackages(t *testing.T) {
	handler := core.NewFileHandler(tests.TestLogger)

	handler.AddPackages(packagesToAdd)

	packages := handler.GetPackages()
	newLen := len(packages)

	if newLen != baseLenPkgInstall {
		t.Errorf(
			"starting length: %d does not match ending length of add packages: %d",
			baseLenPkgInstall,
			newLen,
		)
	}
}

func TestRemovePackagesExactName(t *testing.T) {
	log := logger.NewLogger(log.New(bytes.NewBuffer([]byte{}), "", log.Ldate), logger.Ldebug)
	handler := core.NewFileHandler(log)

	packagesCopy := make([]string, len(packagesToAdd))
	copy(packagesCopy, packagesToAdd)
	packagesCopy = append(packagesCopy, []string{"zoom.pkg", "antivirus.pkg", "remote software.pkg", "vpn software"}...)

	handler.AddPackages(packagesCopy)

	expectedLength := len(handler.GetPackages())

	r := rand.New(rand.NewSource(time.Now().Unix()))

	randomSelection := strings.ToLower(packagesCopy[r.Intn(len(packagesCopy))])

	handler.RemovePackages([]string{randomSelection})

	newLen := len(handler.GetPackages())

	// removing 1 file, so it should be 1 less
	if newLen != expectedLength-1 {
		t.Errorf("failed to remove package, got %d instead of %d", newLen, expectedLength)
	}
}

func TestRemovePackagesWithSubstring(t *testing.T) {
	handler := core.NewFileHandler(tests.TestLogger)

	packagesCopy := make([]string, len(packagesToAdd))
	copy(packagesCopy, packagesToAdd)
	packagesCopy = append(packagesCopy, []string{"zoom.pkg", "antivirus.pkg", "remote software.pkg", "vpn software"}...)

	handler.AddPackages(packagesCopy)

	packagesSubstr := []string{}

	for _, entry := range packagesCopy {
		packagesSubstr = append(packagesSubstr, strings.ReplaceAll(entry, ".pkg", ""))
	}

	handler.RemovePackages(packagesSubstr)

	tests.Checkf(t,
		len(handler.GetAllPackages()) != 0,
		"failed to remove packages, expected 0 got %d",
		len(handler.GetAllPackages()),
	)
}

func TestRemovePackagesNoMatch(t *testing.T) {
	handler := core.NewFileHandler(tests.TestLogger)

	handler.AddPackages(packagesToAdd)

	handler.RemovePackages([]string{"nonexistent.pkg", "someother.pkg", "a pkg here"})

	tests.Checkf(t,
		len(handler.GetAllPackages()) != len(packagesToAdd),
		"expected packages length to be %d, got %d",
		len(packagesToAdd),
		len(handler.GetAllPackages()),
	)
}

func TestInstalledPackagesNormal(t *testing.T) {
	handler := core.NewFileHandler(tests.TestLogger)

	alreadyInstalledCount := 0

	handler.AddMapPackages(packagesToInstall)
	for _, installedNames := range packagesToInstall {
		if handler.IsInstalled(installedNames, searchDirectoryFiles) {
			alreadyInstalledCount += 1
		}
	}

	if alreadyInstalledCount != baseLenPkgInstall {
		t.Error("packages failed to install")
	}
}

func TestInstallPackagesAddNewPackages(t *testing.T) {
	projectDirectory := t.TempDir()

	log := logger.NewLogger(log.New(bytes.NewBuffer([]byte{}), "", log.Ldate), logger.Ldebug)
	handler := core.NewFileHandler(log)

	handler.AddMapPackages(packagesToInstall)
	handler.AddPackages(packagesToAdd)

	installedCount := 0
	expectedLen := len(packagesToAdd)

	// creating the pkg files in the temp folder
	for pkg, installedNames := range handler.GetAllPackages() {
		isInstalled := handler.IsInstalled(installedNames, searchDirectoryFiles)
		if !isInstalled {
			installedCount += 1

			err := os.WriteFile(strings.ReplaceAll(projectDirectory+"/"+pkg, ".pkg", ""), []byte{}, 0o644)
			if err != nil {
				continue
			}
		}
	}

	if installedCount != expectedLen {
		t.Errorf("packages failed to install, %d != %d", installedCount, baseLenPkgInstall)
	}

	files, err := os.ReadDir(projectDirectory)
	// just an extra check, to be honest the statement above is good enough.
	if err != nil {
		tests.Fatal(t, err, fmt.Sprintf("Failed to read directory %s: %v", projectDirectory, err))
	}

	if installedCount != len(files) {
		t.Errorf("packages failed to write, got packages: %v", files)
	}
}

func TestInstallPackagesNoPackages(t *testing.T) {
	handler := core.NewFileHandler(tests.TestLogger)

	count := handler.InstallPackages([]string{}, []string{})

	tests.Checkf(t, count != 0, "installed count expected to be 0, got %d", count)
}

func TestReadDmg(t *testing.T) {
	projectDirectory := t.TempDir()

	log := logger.NewLogger(log.New(bytes.NewBuffer([]byte{}), "", log.Ldate), logger.Ldebug)

	testDmgs := []string{
		"test.dmg", "another one.dmg",
	}
	baseLenDmg := len(testDmgs)

	// creating dmg files in the temp folder
	for _, dmgFile := range testDmgs {
		err := os.WriteFile(projectDirectory+"/"+dmgFile, []byte{}, 0o744)
		if err != nil {
			t.Error(err)
		}
	}

	dmg := core.NewFileHandler(log)

	dmg.AddMapPackages(packagesToInstall)
	dmgFiles, err := dmg.ReadDir(projectDirectory, ".dmg")
	if err != nil {
		t.Errorf("failed to read directory: %v", err)
	}

	newLen := len(dmgFiles)

	if newLen != baseLenDmg {
		t.Fatalf("got %d, did not match the baseline %d", newLen, baseLenDmg)
	}
}

func TestCopyApp(t *testing.T) {
	projectDirectory := t.TempDir()

	log := logger.NewLogger(log.New(bytes.NewBuffer([]byte{}), "", log.Ldate), logger.Ldebug)

	handler := core.NewFileHandler(log)

	handler.AddMapPackages(packagesToInstall)
	appBundle := "a program bundle.app"
	appDirectory := "Applications"

	appBundleContentDir := "contents"
	directories := []string{projectDirectory + "/" + appBundle + "/" + appBundleContentDir}

	for _, dir := range directories {
		err := os.MkdirAll(dir, 0o777)
		tests.Fatal(t, err, fmt.Sprintf("Failed to create folder %s: %v", dir, err))
	}

	err := os.WriteFile(projectDirectory+"/"+appBundle+"/"+"file.ini", []byte("text goes here"), 0o744)
	tests.Fatal(t, err, fmt.Sprintf("Failed to write file: %v", err))

	files, err := handler.ReadDir(projectDirectory, ".app")
	tests.Fatal(t, err, fmt.Sprintf("Failed to read folder %s: %v", projectDirectory, err))

	handler.CopyFiles(files, projectDirectory+"/"+appDirectory)

	files, err = handler.ReadDir(projectDirectory+"/"+appDirectory, ".app")
	tests.Fatal(t, err, fmt.Sprintf("Failed to read folder %s: %v", projectDirectory+"/"+appDirectory, err))

	if len(files) != 1 {
		t.Fatal("Failed to read directory")
	}

	files, err = handler.ReadDir(projectDirectory+"/"+appDirectory, ".ini")
	tests.Fatal(t, err, fmt.Sprintf("Failed to read folder %s: %v", projectDirectory+"/"+appDirectory, err))

	for _, file := range files {
		out, err := os.ReadFile(file)
		tests.Fatal(t, err, fmt.Sprintf("Failed to read file %s: %v", file, err))

		if string(out) == "" {
			t.Fatal("failed to copy and write files")
		}
	}
}

func TestCopyFile(t *testing.T) {
	projectDirectory := t.TempDir()

	log := logger.NewLogger(log.New(bytes.NewBuffer([]byte{}), "", log.Ldate), logger.Ldebug)

	handler := core.NewFileHandler(log)

	handler.AddMapPackages(packagesToInstall)
	newTestDir := projectDirectory + "/" + "test-dir"

	err := os.MkdirAll(newTestDir, 0o777)
	tests.Fatal(t, err, fmt.Sprintf("Failed to create folder %s: %v", newTestDir, err))

	fileNames := []string{
		"test.txt", "sample.txt", "ok.txt",
	}
	textContent := "some text here"

	for i, fileName := range fileNames {
		err := os.WriteFile(newTestDir+"/"+fileName, []byte(textContent+strconv.Itoa(i)), 0o744)
		tests.Fatal(t, err, fmt.Sprintf("Failed to write file %s: %v", fileName, err))
	}

	files, err := handler.ReadDir(projectDirectory, ".txt")
	tests.Fatal(t, err, fmt.Sprintf("Failed to read folder %s: %v", projectDirectory, err))

	if len(files) != len(fileNames) {
		t.Fatalf("expected %d got %d", len(fileNames), len(files))
	}

	handler.CopyFiles(files, newTestDir)

	for _, filePath := range files {
		outBytes, err := os.ReadFile(filePath)
		tests.Fatal(t, err, fmt.Sprintf("Failed to read file %s: %v", filePath, err))

		if !strings.Contains(string(outBytes), textContent) {
			t.Fatalf("Got content %s expected %s", string(outBytes), textContent)
		}
	}
}

func TestScriptCacheAddition(t *testing.T) {
	projectDirectory := t.TempDir()

	log := logger.NewLogger(log.New(bytes.NewBuffer([]byte{}), "", log.Ldate), logger.Ldebug)
	handler := core.NewFileHandler(log)

	handler.AddMapPackages(packagesToInstall)
	fakeScriptFiles := []string{
		"file1.sh", "file2.sh",
	}

	for _, file := range fakeScriptFiles {
		err := os.WriteFile(projectDirectory+"/"+file, []byte{}, 0o744)
		if err != nil {
			t.Fatal(err)
		}
	}

	scriptPaths, err := handler.ReadDir(projectDirectory, ".sh")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range fakeScriptFiles {
		_, err := handler.ExecuteScript(file, scriptPaths)
		tests.Fatal(t, err, fmt.Sprintf("Script %s failed: %v", file, err))
	}

	cache := handler.GetScriptCache()
	if len(cache) != len(fakeScriptFiles) {
		t.Fatalf("got %d from cache expected %d", len(cache), len(fakeScriptFiles))
	}

	indx := rand.Intn(len(fakeScriptFiles))

	if _, ok := cache[strings.ToLower(fakeScriptFiles[indx])]; !ok {
		t.Fatalf("could not find %s in cache", fakeScriptFiles[indx])
	}
}

func TestScriptExecution(t *testing.T) {
	projectDirectory := t.TempDir()

	log := logger.NewLogger(log.New(bytes.NewBuffer([]byte{}), "", log.Ldate), logger.Ldebug)
	handler := core.NewFileHandler(log)

	handler.AddMapPackages(packagesToInstall)
	baseScriptNames := []string{
		"file1.sh",
	}

	msg := "A test message here"
	cmd := fmt.Sprintf(`echo "%s"`, msg)

	env := "#/usr/bin/env bash"
	scriptContent := fmt.Sprintf("%s\n%s", env, cmd)

	for _, path := range baseScriptNames {
		err := os.WriteFile(projectDirectory+"/"+path, []byte(scriptContent), 0o755)
		if err != nil {
			t.Fatal(err)
		}
	}

	executingScripts := []string{
		"file1.sh",
	}

	scriptPaths, err := handler.ReadDir(projectDirectory, ".sh")
	if err != nil {
		t.Fatal(err)
	}

	for _, execScript := range executingScripts {
		out, err := handler.ExecuteScript(execScript, scriptPaths)
		if err != nil {
			t.Fatal(err)
		}

		if out != msg {
			t.Fatalf("got %s expected %s", out, msg)
		}
	}
}

func TestPackageString(t *testing.T) {
	handler := core.NewFileHandler(tests.TestLogger)

	pkg := "chrome.pkg"
	installFiles := "chrome.app,google chrome.app"

	copyPkgToInstall := maps.Clone(packagesToInstall)

	copyPkgToInstall[pkg] = strings.Split(installFiles, ",")

	handler.AddMapPackages(copyPkgToInstall)

	str := handler.PackageString()

	for key, val := range copyPkgToInstall {
		// installation packages are separated by a comma
		valStr := strings.ToLower(strings.Join(val, ","))
		key = strings.ToLower(key)

		tests.Checkf(t, strings.Contains(str, key) == false, "failed to find %s in %s", key, str)
		tests.Checkf(t, strings.Contains(str, valStr) == false, "failed to find %s in %s", valStr, str)
	}
}

func TestPackageStringNoInstallFiles(t *testing.T) {
	handler := core.NewFileHandler(tests.TestLogger)

	pkg := "chrome.pkg"

	pkgCopy := make([]string, len(packagesToAdd))
	copy(pkgCopy, packagesToAdd)
	pkgCopy = append(pkgCopy, pkg)

	handler.AddPackages(pkgCopy)

	str := handler.PackageString()

	for _, val := range pkgCopy {
		val = strings.ToLower(val)

		tests.Checkf(t, strings.Contains(str, val) == false, "string %s not found in %s", val, str)
	}
}
