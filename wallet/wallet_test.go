package wallet_test

import (
	"crypto/ed25519"

	ed25519_sm "github.com/spacemeshos/ed25519"

	"encoding/hex"
	"fmt"
	"testing"

	"github.com/spacemeshos/smcli/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"
)

func TestWalletFromNewMnemonic(t *testing.T) {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.WalletFromMnemonic(mnemonic)

	assert.NotNil(t, w)
	assert.Equal(t, mnemonic, w.Mnemonic())
}

func TestWalletFromGivenMnemonic(t *testing.T) {
	mnemonic := "film theme cheese broken kingdom destroy inch ready wear inspire shove pudding"
	w := wallet.WalletFromMnemonic(mnemonic)
	keyPath, err := wallet.StringToHDPath("m/44'/540'/0'/0'/0'")
	assert.NoError(t, err)
	keyPair := w.ComputeKeyPair(keyPath)

	expPubKey :=
		"c177dee3e454b2505f9a511155aad4afe8fea227db189b3aaaa9b092fe45567b"
	expPrivKey :=
		"628d5cb651f8c0d4139dfd5fe97079c66c20452e3e2a8b7b4b6c5fc56c6c3e3ec177dee3e454b2505f9a511155aad4afe8fea227db189b3aaaa9b092fe45567b"

	actualPubKey := hex.EncodeToString(keyPair.Public)
	actualPrivKey := hex.EncodeToString(keyPair.Private)
	assert.Equal(t, expPubKey, actualPubKey)
	assert.Equal(t, expPrivKey, actualPrivKey)

	msg := []byte("hello world")
	// Sanity check that the keypair works with the standard ed25519 library
	sig := ed25519.Sign(ed25519.PrivateKey(keyPair.Private), msg)
	assert.True(t, ed25519.Verify(ed25519.PublicKey(keyPair.Public), msg, sig))

	// Sanity check that the keypair works with the spacemesh ed25519 impl
	sig = ed25519_sm.Sign(keyPair.Private, msg)
	assert.True(t, ed25519_sm.Verify(keyPair.Public, msg, sig))

	// Sanity check that the keypair works with the extended spacemesh ed25519 impl
	sig = ed25519_sm.Sign2(keyPair.Private, msg)
	assert.True(t, ed25519_sm.Verify2(keyPair.Public, msg, sig))
	extractedPubKey, err := ed25519_sm.ExtractPublicKey(msg, sig)
	assert.NoError(t, err)
	assert.Equal(t, keyPair.Public, extractedPubKey)
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

func benchMarkNewAccount(n int, b *testing.B) {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.WalletFromMnemonic(mnemonic)
	for i := 0; i < b.N; i++ {
		for j := 0; j < n; j++ {
			acctNum := j
			w.Account(fmt.Sprintf("test account %d", acctNum))
		}
	}
}

func BenchmarkNewAccount1000(b *testing.B) {
	benchMarkNewAccount(1000, b)
}
func BenchmarkNewAccount2000(b *testing.B) {
	benchMarkNewAccount(2000, b)
}
func BenchmarkNewAccount3000(b *testing.B) {
	benchMarkNewAccount(3000, b)
}
func BenchmarkNewAccount4000(b *testing.B) {
	benchMarkNewAccount(4000, b)
}
func BenchmarkNewAccount5000(b *testing.B) {
	benchMarkNewAccount(5000, b)
}
func BenchmarkNewAccount6000(b *testing.B) {
	benchMarkNewAccount(6000, b)
}
func BenchmarkNewAccount7000(b *testing.B) {
	benchMarkNewAccount(7000, b)
}
func BenchmarkNewAccount8000(b *testing.B) {
	benchMarkNewAccount(8000, b)
}
func BenchmarkNewAccount9000(b *testing.B) {
	benchMarkNewAccount(9000, b)
}

func BenchmarkNewAccount10000(b *testing.B) {
	benchMarkNewAccount(10000, b)
}

func BenchmarkNewAccount20000(b *testing.B) {
	benchMarkNewAccount(20000, b)
}

func BenchmarkNewAccount30000(b *testing.B) {
	benchMarkNewAccount(30000, b)
}
func BenchmarkNewAccount40000(b *testing.B) {
	benchMarkNewAccount(40000, b)
}
func BenchmarkNewAccount50000(b *testing.B) {
	benchMarkNewAccount(50000, b)
}
func BenchmarkNewAccount60000(b *testing.B) {
	benchMarkNewAccount(60000, b)
}
func BenchmarkNewAccount70000(b *testing.B) {
	benchMarkNewAccount(70000, b)
}
func BenchmarkNewAccount80000(b *testing.B) {
	benchMarkNewAccount(80000, b)
}

// 3103596912 ns/op
// 2777190847 ns/op
