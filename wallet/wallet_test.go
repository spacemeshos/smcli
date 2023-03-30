package wallet_test

import (
	"crypto/ed25519"

	"encoding/hex"
	"testing"

	"github.com/spacemeshos/smcli/wallet"
	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	w := wallet.NewWallet()
	assert.NotNil(t, w)
}

func TestNewWalletFromSeed(t *testing.T) {
	seed := []byte("spacemesh is the best blockchain")
	w := wallet.NewWalletFromSeed(seed)
	assert.NotNil(t, w)
	assert.Len(t, w.Secrets.Accounts, 1)
	keypair := w.Secrets.Accounts[0]

	expPubKey :=
		"680e002010e2a12d17922bade5c3c25d54a911abd41f48e16b3bbe0c83df96a9"
	expPrivKey :=
		"73706163656d65736820697320746865206265737420626c6f636b636861696e680e002010e2a12d17922bade5c3c25d54a911abd41f48e16b3bbe0c83df96a9"

	actualPubKey := hex.EncodeToString(keypair.Public)
	actualPrivKey := hex.EncodeToString(keypair.Private)
	assert.Equal(t, expPubKey, actualPubKey)
	assert.Equal(t, expPrivKey, actualPrivKey)

	msg := []byte("hello world")
	// Sanity check that the keypair works with the ed25519 library
	sig := ed25519.Sign(keypair.Private, msg)
	assert.True(t, ed25519.Verify(keypair.Public, msg, sig))

	// create another wallet from the same seed
	w2 := wallet.NewWalletFromSeed(seed)
	assert.NotNil(t, w2)
	assert.Equal(t, w.Secrets.Accounts, w2.Secrets.Accounts)
}

//func TestWalletFromNewMnemonic(t *testing.T) {
//	entropy, _ := bip39.NewEntropy(256)
//	mnemonic, _ := bip39.NewMnemonic(entropy)
//	w := wallet.NewWalletFromSeed(mnemonic)
//
//	assert.NotNil(t, w)
//	assert.Equal(t, mnemonic, w.Mnemonic())
//}

//func TestWalletFromGivenMnemonic(t *testing.T) {
//	mnemonic := "film theme cheese broken kingdom destroy inch ready wear inspire shove pudding"
//	w := wallet.NewWalletFromSeed(mnemonic)
//	keyPath, err := wallet.StringToHDPath("m/44'/540'/0'/0'/0'")
//	assert.NoError(t, err)
//	keyPair, err := w.ComputeKeyPair(keyPath)
//	assert.NoError(t, err)
//	expPubKey :=
//		"205d8d4e458b163d5ba15ac712951d5659cc51379e7e0ad13acc97303aa85093"
//	expPrivKey :=
//		"669d091195f950e6255a2e8778eea7be4f7a66afe855957404ec1520c8a11ff1205d8d4e458b163d5ba15ac712951d5659cc51379e7e0ad13acc97303aa85093"
//
//	actualPubKey := hex.EncodeToString(keyPair.Public)
//	actualPrivKey := hex.EncodeToString(keyPair.Private)
//	assert.Equal(t, expPubKey, actualPubKey)
//	assert.Equal(t, expPrivKey, actualPrivKey)
//
//	msg := []byte("hello world")
//	// Sanity check that the keypair works with the standard ed25519 library
//	sig := ed25519.Sign(keyPair.Private, msg)
//	assert.True(t, ed25519.Verify(keyPair.Public, msg, sig))
//}

//func TestKeysInWalletMaintainExpectedPath(t *testing.T) {
//	entropy, _ := bip39.NewEntropy(256)
//	mnemonic, _ := bip39.NewMnemonic(entropy)
//	w := wallet.NewWalletFromSeed(mnemonic)
//
//	for i := 0; i < 100; i++ {
//		path, _ := wallet.StringToHDPath(fmt.Sprintf("m/44'/540'/%d'/%d'/%d'", i, i, i))
//		keyPair, err := w.ComputeKeyPair(path)
//		assert.NoError(t, err)
//		assert.Equal(t, path, keyPair.Path)
//	}
//}

//func TestKeysInWalletMaintainSalt(t *testing.T) {
//	entropy, _ := bip39.NewEntropy(256)
//	mnemonic, _ := bip39.NewMnemonic(entropy)
//	w := wallet.NewWalletFromSeed(mnemonic)
//	fmt.Println(string(w.Salt()))
//
//	path, _ := wallet.StringToHDPath("m/44'/540'")
//	keyPair, err := w.ComputeKeyPair(path)
//	assert.NoError(t, err)
//	assert.Equal(t, keyPair.Salt, w.Salt())
//	// ... and try with a different length path
//	path, _ = wallet.StringToHDPath("m/44'/540'/0'")
//	keyPair, err = w.ComputeKeyPair(path)
//	assert.NoError(t, err)
//	assert.Equal(t, keyPair.Salt, w.Salt())
//}

//func TestComputeKeyPairFailsForUnhardenedPathSegment(t *testing.T) {
//	entropy, _ := bip39.NewEntropy(256)
//	mnemonic, _ := bip39.NewMnemonic(entropy)
//	w := wallet.NewWalletFromSeed(mnemonic)
//	path, _ := wallet.StringToHDPath("m/44'/540'/0'/0'/0")
//	_, err := w.ComputeKeyPair(path)
//	assert.Error(t, err)
//}

func benchmarkComputeKeyPair(n int, b *testing.B) {
	wallet.NewWallet()
}

//func benchmarkComputeKeyPair(n int, b *testing.B) {
//	entropy, _ := bip39.NewEntropy(256)
//	mnemonic, _ := bip39.NewMnemonic(entropy)
//	w := wallet.NewWalletFromSeed(mnemonic)
//	for i := 0; i < b.N; i++ { // benchmark-controlled loop
//		for j := 0; j < n; j++ { // specified number of iterations
//			path, _ := wallet.StringToHDPath(fmt.Sprintf("m/44'/540'/0'/0'/%d'", j))
//			w.ComputeKeyPair(path)
//		}
//	}
//}

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
