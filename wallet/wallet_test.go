package wallet_test

import (
	"testing"

	"github.com/spacemeshos/smcli/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"
)

func TestWalletFromMnemonic(t *testing.T) {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.WalletFromMnemonic(mnemonic)

	assert.NotNil(t, w)
	assert.Equal(t, mnemonic, w.Mnemonic())
}

func TestNewAccount(t *testing.T) {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.WalletFromMnemonic(mnemonic)

	accName := "test"
	acc := w.Account(accName)
	assert.NotNil(t, acc)
	assert.Equal(t, accName, acc.Name)
}
