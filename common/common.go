package common

import (
	"fmt"
	"os"
	"path/filepath"
)

// .spacemesh
// ├── bin
// │   └── [ Linux | macOS | Windows ]
// │       ├── go-spacemesh
// │       ├── config.json
// │       └── node.zip
// ├── logs
// │   └── go-spacemesh.log
// ├── networks.json
// ├── config.yaml
// └── state.json

func DotDirectory() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home + "/.spacemesh")
}
func DotBinDirectory() string {
	return filepath.Join(DotDirectory(), "bin")
}
func ConfigFileName() string {
	return "config"
}
func ConfigFileType() string {
	return "yaml"
}
func DotConfigFile() string {
	return filepath.Join(DotDirectory(),
		fmt.Sprintf("%s.%s", ConfigFileName(), ConfigFileType()))
}
func DotStateFile() string {
	return filepath.Join(DotDirectory(), "state.json")
}
func DotNetworksFile() string {
	return filepath.Join(DotDirectory(), "networks.json")
}
func DotLogDirectory() string {
	return filepath.Join(DotDirectory(), "logs")
}
func DotLogFile() string {
	return filepath.Join(DotDirectory(), "go-spacemesh.log")
}
