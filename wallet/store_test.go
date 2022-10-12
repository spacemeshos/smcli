package wallet_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/spacemeshos/smcli/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"
)

func TestStoreAndRetrieveWalletToFromFile(t *testing.T) {
	wKey := wallet.NewKey(
		wallet.WithArgon2idPassword("password"),
	)
	wStore := wallet.NewStore(wKey)

	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.WalletFromMnemonic(mnemonic)

	file, err := ioutil.TempFile("./", "test_wallet.*.json")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	fmt.Println(file.Name())
	err = wStore.Export(file, w)
	assert.NoError(t, err)

	wStore2 := wallet.NewStore(wKey)
	w2, err := wStore2.Open(file)
	assert.NoError(t, err)
	assert.Equal(t, w.Mnemonic(), w2.Mnemonic())
}
