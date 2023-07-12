package internal

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-secure-stdlib/password"

	"github.com/spacemeshos/smcli/wallet"
)

// LoadWallet from a file, asks for the password from stdin.
func LoadWallet(path string, debug bool) (*wallet.Wallet, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	fmt.Print("Enter wallet password: ")
	pass, err := password.Read(os.Stdin)
	fmt.Println()
	if err != nil {
		return nil, err
	}

	wk := wallet.NewKey(wallet.WithPasswordOnly([]byte(pass)))

	w, err := wk.Open(f, debug)
	if err != nil {
		return nil, err
	}

	return w, nil
}
