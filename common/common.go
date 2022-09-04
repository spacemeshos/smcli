package common

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	DefaultNodeVersion  = "v0.2.16-beta.0"
	DefaultDiscoveryUrl = "https://discover.spacemesh.io/networks.json"
)

func NodeDownloadUrl() string {
	baseDownloadPath := "storage.googleapis.com/go-spacemesh-release-builds/" + DefaultNodeVersion
	downloadPath := filepath.Join(baseDownloadPath, SystemType()+".zip")

	return "https://" + downloadPath
}

const (
	Windows = "Windows"
	MacOS   = "macOS"
	Linux   = "Linux"
)

func SystemType() string {
	os := runtime.GOOS
	switch os {
	case "windows":
		return Windows
	case "darwin":
		return MacOS
	case "linux":
		return Linux
	default:
		panic(fmt.Sprintf("unsupported os: %s", os))
	}
}

// .spacemesh
// ├── bin
// │   └── [ Linux | macOS | Windows ]
// │       ├── go-spacemesh
// │       └── config.json
// ├── logs
// │   └── go-spacemesh.log
// ├── networks.json
// ├── config.yaml
// └── state.json

func DotDirectory() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home + "/.spacemesh")
}
func BinDirectory() string {
	return filepath.Join(DotDirectory(), "bin")
}

// The unzip of the node.zip creates an extra directory with the system type.
func BinDirectoryWithSysType() string {
	return filepath.Join(BinDirectory(), SystemType())
}

func NodeBin() string {
	return filepath.Join(BinDirectoryWithSysType(), "go-spacemesh")
}
func NodeConfigFile() string {
	return filepath.Join(BinDirectoryWithSysType(), "config.toml")
}
func NodeDataDirectory() string {
	return filepath.Join(DotDirectory(), "data")
}
func ConfigFileName() string {
	return "config"
}
func ConfigFileType() string {
	return "yaml"
}
func ConfigFile() string {
	return filepath.Join(DotDirectory(),
		fmt.Sprintf("%s.%s", ConfigFileName(), ConfigFileType()))
}
func StateFile() string {
	return filepath.Join(DotDirectory(), "state.json")
}
func NetworksFile() string {
	return filepath.Join(DotDirectory(), "networks.json")
}
func LogDirectory() string {
	return filepath.Join(DotDirectory(), "logs")
}
func LogFile() string {
	return filepath.Join(LogDirectory(), "go-spacemesh.log")
}

// InitDotDir creates the .spacemesh directory
// and subdirectories if they doesn't exist.
func InitDotDir() {
	cobra.CheckErr(os.MkdirAll(BinDirectory(), 0770))
	cobra.CheckErr(os.MkdirAll(LogDirectory(), 0770))
}

// OpenNodeLogFile creates the go-spacemesh.log file if it doesn't exist and
// returns the file pointer. If it exists, it returns the file pointer.
func OpenNodeLogFile() (*os.File, error) {
	return os.OpenFile(LogFile(), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
}
