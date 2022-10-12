package wallet_test

import (
	"testing"

	"github.com/spacemeshos/smcli/wallet"
	"github.com/tyler-smith/go-bip39"
)

func TestStoreAndRetrieveWalletToFromFile(t *testing.T) {
	wKey := wallet.NewKey(
		wallet.WithPbkdf2Password("password"),
	)
	wStore := wallet.NewStore(wKey)

	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.WalletFromMnemonic(mnemonic)

	wStore.Export("test_wallet.json", w)
}
