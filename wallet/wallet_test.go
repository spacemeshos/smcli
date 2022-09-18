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

	// Default account already exists on wallet creation at m/44'/540'/0'
	accName := "default"
	acc := w.Account(accName)
	assert.NotNil(t, acc)
	assert.Equal(t, accName, acc.Name)
	expectedPath, err := wallet.StringToHDPath("m/44'/540'/0'")
	assert.NoError(t, err)
	assert.Equal(t, expectedPath, acc.Path())

	// Creating a new account increments the account by 1
	accName = "test account"
	acc = w.Account(accName)
	assert.NotNil(t, acc)
	assert.Equal(t, accName, acc.Name)
	expectedPath, err = wallet.StringToHDPath("m/44'/540'/1'")
	assert.NoError(t, err)
	assert.Equal(t, expectedPath, acc.Path())

	// Test account is still the same
	accName = "test account"
	acc = w.Account(accName)
	assert.NotNil(t, acc)
	assert.Equal(t, accName, acc.Name)
	expectedPath, err = wallet.StringToHDPath("m/44'/540'/1'")
	assert.NoError(t, err)
	assert.Equal(t, expectedPath, acc.Path())

	// Default account is still the same
	accName = "default"
	acc = w.Account(accName)
	assert.NotNil(t, acc)
	assert.Equal(t, accName, acc.Name)
	expectedPath, err = wallet.StringToHDPath("m/44'/540'/0'")
	assert.NoError(t, err)
	assert.Equal(t, expectedPath, acc.Path())
}
