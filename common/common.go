package common

import (
	"os"
	"path/filepath"
	"time"
)

const (
	// MaxAccountsPerWallet is the maximum number of accounts that a single wallet file may contain.
	// It's relatively arbitrary but we need some limit.
	MaxAccountsPerWallet = 128
)

func NowTimeString() string {
	return time.Now().UTC().Format("2006-01-02T15-04-05.000") + "Z"
}

// .spacemesh
// ├── bin
// │   └── [ Linux | macOS | Windows ]
// │       ├── go-spacemesh
// │       └── config.json
// ├── logs
// │   └── go-spacemesh.log
// ├── config.yaml
// └── state.json

func DotDirectory() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home + "/.spacemesh")
}
func ConfigFileName() string {
	return "config"
}
func ConfigFileType() string {
	return "yaml"
}
func StateFile() string {
	return filepath.Join(DotDirectory(), "state.json")
}
func WalletFile() string {
	return filepath.Join(DotDirectory(), "wallet_"+NowTimeString()+".json")
}
