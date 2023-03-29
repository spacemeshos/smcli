package wallet_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/spacemeshos/smcli/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"
)

func TestStoreAndRetrieveWalletToFromFile(t *testing.T) {
	wKey := wallet.NewKey(
		wallet.WithPbkdf2Password("password"),
	)

	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.NewWalletFromMnemonic(mnemonic)

	file, err := os.CreateTemp("./", "test_wallet.*.json")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	fmt.Println(file.Name())
	err = wKey.Export(file, w)
	assert.NoError(t, err)

	w2, err := wKey.Open(file)
	assert.NoError(t, err)
	assert.Equal(t, w.Mnemonic(), w2.Mnemonic())
}
