package common

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	DefaultNodeVersion  = "v0.2.16-beta.0"
	DefaultDiscoveryUrl = "https://discover.spacemesh.io/networks.json"
)

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	return path.Join(home, ".spacemesh")
}

func NodeDownloadUrl() string {
	baseDownloadPath := "storage.googleapis.com/go-spacemesh-release-builds/" + DefaultNodeVersion
	downloadPath := ""
	os := runtime.GOOS
	switch os {
	case "windows":
		downloadPath = path.Join(baseDownloadPath, "Windows.zip")
	case "darwin":
		downloadPath = path.Join(baseDownloadPath, "macOS.zip")
	case "linux":
		downloadPath = path.Join(baseDownloadPath, "Linux.zip")
	default:
		panic(fmt.Sprintf("unsupported os: %s", os))
	}

	return "https://" + downloadPath
}
