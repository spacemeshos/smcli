package common

import (
	"os"
	"path"
	"path/filepath"

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
	downloadPath := filepath.Join(baseDownloadPath, SystemType()+".zip")

	return "https://" + downloadPath
}
