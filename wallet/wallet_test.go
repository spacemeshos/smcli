package wallet_test

import (
	"crypto/ed25519"

	"encoding/hex"
	"fmt"
	"testing"

	ed25519sm "github.com/spacemeshos/ed25519"
	"github.com/spacemeshos/smcli/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"
)

func TestWalletFromNewMnemonic(t *testing.T) {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.NewWalletFromMnemonic(mnemonic)

	assert.NotNil(t, w)
	assert.Equal(t, mnemonic, w.Mnemonic())
}

func TestWalletFromGivenMnemonic(t *testing.T) {
	mnemonic := "film theme cheese broken kingdom destroy inch ready wear inspire shove pudding"
	w := wallet.NewWalletFromMnemonic(mnemonic)
	keyPath, err := wallet.StringToHDPath("m/44'/540'/0'/0'/0'")
	assert.NoError(t, err)
	keyPair, err := w.ComputeKeyPair(keyPath)
	assert.NoError(t, err)
	expPubKey :=
		"7b2b71492ef629c5fbd18a469ed984fa7ec749bdd877cab74344c260e60c605a"
	expPrivKey :=
		"40245a27ce660ea9a135a52a53fa388eb015b8033d253515317c975f01a93a0b7b2b71492ef629c5fbd18a469ed984fa7ec749bdd877cab74344c260e60c605a"

	actualPubKey := hex.EncodeToString(keyPair.Public)
	actualPrivKey := hex.EncodeToString(keyPair.Private)
	assert.Equal(t, expPubKey, actualPubKey)
	assert.Equal(t, expPrivKey, actualPrivKey)

	msg := []byte("hello world")
	// Sanity check that the keypair works with the standard ed25519 library
	sig := ed25519.Sign(keyPair.Private, msg)
	assert.True(t, ed25519.Verify(keyPair.Public, msg, sig))

	// Sanity check that the keypair works with the spacemesh ed25519 impl
	sig = ed25519sm.Sign(keyPair.Private, msg)
	assert.True(t, ed25519sm.Verify(keyPair.Public, msg, sig))

	// Sanity check that the keypair works with the extended spacemesh ed25519 impl
	// TODO: drop this?
	sig = ed25519sm.Sign2(keyPair.Private, msg)
	assert.True(t, ed25519sm.Verify2(keyPair.Public, msg, sig))
	extractedPubKey, err := ed25519sm.ExtractPublicKey(msg, sig)
	assert.NoError(t, err)
	assert.Equal(t, keyPair.Public, extractedPubKey)
}

func TestKeysInWalletMaintainExpectedPath(t *testing.T) {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.NewWalletFromMnemonic(mnemonic)

	for i := 0; i < 100; i++ {
		path, _ := wallet.StringToHDPath(fmt.Sprintf("m/44'/540'/%d'/%d'/%d'", i, i, i))
		keyPair, err := w.ComputeKeyPair(path)
		assert.NoError(t, err)
		assert.Equal(t, keyPair.Path, path)
	}
}

func TestKeysInWalletMaintainSalt(t *testing.T) {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.NewWalletFromMnemonic(mnemonic)
	fmt.Println(string(w.Salt()))

	path, _ := wallet.StringToHDPath("m/44'/540'")
	keyPair, err := w.ComputeKeyPair(path)
	assert.NoError(t, err)
	assert.Equal(t, keyPair.Salt, w.Salt())
	// ... and try with a different length path
	path, _ = wallet.StringToHDPath("m/44'/540'/0'")
	keyPair, err = w.ComputeKeyPair(path)
	assert.NoError(t, err)
	assert.Equal(t, keyPair.Salt, w.Salt())
}

func TestComputeKeyPairFailsForUnhardenedPathSegment(t *testing.T) {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.NewWalletFromMnemonic(mnemonic)
	path, _ := wallet.StringToHDPath("m/44'/540'/0'/0'/0")
	_, err := w.ComputeKeyPair(path)
	assert.Error(t, err)
}

func TestListHardwareWallets(t *testing.T) {

}

func benchmarkComputeKeyPair(n int, b *testing.B) {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.NewWalletFromMnemonic(mnemonic)
	for i := 0; i < b.N; i++ { // benchmark-controlled loop
		for j := 0; j < n; j++ { // specified number of iterations
			path, _ := wallet.StringToHDPath(fmt.Sprintf("m/44'/540'/0'/0'/%d'", j))
			w.ComputeKeyPair(path)
		}
	}
}
func Benchmark10000(b *testing.B) {
	benchmarkComputeKeyPair(10000, b)
}
func Benchmark20000(b *testing.B) {
	benchmarkComputeKeyPair(20000, b)
}
func Benchmark30000(b *testing.B) {
	benchmarkComputeKeyPair(30000, b)
}
func Benchmark40000(b *testing.B) {
	benchmarkComputeKeyPair(40000, b)
}
func Benchmark50000(b *testing.B) {
	benchmarkComputeKeyPair(50000, b)
}
func Benchmark100000(b *testing.B) {
	benchmarkComputeKeyPair(100000, b)
}
