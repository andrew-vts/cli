package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ansel1/merry"
	"github.com/dghubble/sling"
	"github.com/dickeyxxx/golock"
)

var installLockPath = filepath.Join(DataHome, "v5.lock")

// Install installs the CLI
func Install(channel string) {
	if exists, err := FileExists(binPath()); exists || err != nil {
		WarnIfError(err)
		return
	}
	os := runtime.GOOS
	if os == "win32" {
		os = "windows"
	}
	arch := runtime.GOARCH
	if arch == "amd64" {
		arch = "x64"
	}
	if arch == "386" {
		arch = "x86"
	}
	manifest := GetUpdateManifest(channel, os, arch)
	DownloadCLI(channel, filepath.Join(DataHome, "client"), os, arch, manifest)
	showCursor()
}

// DownloadCLI downloads a CLI update to a given path
func DownloadCLI(channel, path, runtimeOS, runtimeARCH string, manifest *Manifest) {
	locked, err := golock.IsLocked(installLockPath)
	LogIfError(err)
	if locked {
		Warn("Update in progress")
	}
	LogIfError(golock.Lock(installLockPath))
	unlock := func() {
		golock.Unlock(installLockPath)
	}
	defer unlock()
	downloadingMessage := fmt.Sprintf("sfdx-cli: Updating to %s...", manifest.Version)
	if Channel != "stable" {
		downloadingMessage = fmt.Sprintf("%s (%s)", downloadingMessage, Channel)
	}
	url := "https://developer.salesforce.com/media/salesforce-cli/v6/sfdx-cli/channels/latest/sfdx-cli-v" + manifest.Version + "-" + runtimeOS + "-" + runtimeARCH + ".tar.xz"
	reader, getSha, err := downloadXZ(url, downloadingMessage)
	must(err)
	tmp := filepath.Join(DataHome, "tmp")
	mkdirp(tmp)
	defer os.RemoveAll(tmp)
	must(extractTar(reader, tmp))
	sha := getSha()
	if sha != manifest.Sha256XZ {
		must(merry.Errorf("SHA mismatch: expected %s to be %s", sha, manifest.Sha256XZ))
	}
	exists, _ := FileExists(path)
	if exists {
		must(os.RemoveAll(path))
	}
	must(try(5, func() error {
		return os.Rename(filepath.Join(tmp, "sfdx-cli-v"+manifest.Version+"-"+runtimeOS+"-"+runtimeARCH), path)
	}))
	Debugf("updated to %s\n", manifest.Version)
}

func try(max int, fn func() error) error {
	var err error
	for i := 0; i < max; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		Errln(err)
		seconds := 2 << uint(i)
		Errf("sfdx-cli: trying again in %d seconds...\n", seconds)
		time.Sleep(time.Second * time.Duration(seconds))
	}
	return err
}

// Manifest is the manifest.json for releases
type Manifest struct {
	Version  string `json:"version"`
	Sha256XZ string `json:"sha256xz"`
}

var updateManifestRetrying = false

// GetUpdateManifest loads the manifest.json for a channel
func GetUpdateManifest(channel, os, arch string) *Manifest {
	var m Manifest
	url := "https://developer.salesforce.com/media/salesforce-cli/v6/sfdx-cli/channels/latest/" + os + "-" + arch
	rsp, err := sling.New().Client(apiHTTPClient).Get(url).ReceiveSuccess(&m)
	if err != nil && !updateManifestRetrying {
		updateManifestRetrying = true
		return GetUpdateManifest(channel, os, arch)
	}
	must(err)
	must(getHTTPError(rsp))
	return &m
}

func binPath() string {
	bin := filepath.Join(DataHome, "client", "bin", "sfdx")
	if runtime.GOOS == WINDOWS {
		bin = bin + ".cmd"
	}
	return bin
}