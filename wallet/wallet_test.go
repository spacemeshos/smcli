package wallet

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tyler-smith/go-bip39"
)

const Bip44Prefix = "m/44'/540'"

func TestRandomAndMnemonic(t *testing.T) {
	n := 3

	// generate a wallet with a random mnemonic
	w1, err := NewMultiWalletRandomMnemonic(n)
	require.NoError(t, err)
	require.Len(t, w1.Secrets.Accounts, n)

	// now use that mnemonic to generate a new wallet
	w2, err := NewMultiWalletFromMnemonic(w1.Mnemonic(), n)
	require.NoError(t, err)
	require.Len(t, w2.Secrets.Accounts, n)

	// make sure all the keys match
	for i := 0; i < n; i++ {
		require.Equal(t, w1.Secrets.Accounts[i].Private, w2.Secrets.Accounts[i].Private)
		require.Equal(t, w1.Secrets.Accounts[i].Public, w2.Secrets.Accounts[i].Public)
	}
}

func TestAccountFromSeed(t *testing.T) {
	master, err := NewMasterKeyPair(goodSeed)
	require.NoError(t, err)
	accts, err := accountsFromMaster(master, goodSeed, 1)
	require.NoError(t, err)
	require.Len(t, accts, 1)
	keypair := accts[0]

	expPubKey := "feae6977b42bf3441d04314d09c72c5d6f2d1cb4bf94834680785b819f8738dd"
	expPrivKey := "05fe9affa5562ca833faf3803ce5f6f7615d3c37c4a27903492027f6853e486dfeae6977b42bf3441d04314d09c72c5d6f2d1cb4bf94834680785b819f8738dd"

	actualPubKey := hex.EncodeToString(keypair.Public)
	actualPrivKey := hex.EncodeToString(keypair.Private)
	require.Equal(t, expPubKey, actualPubKey)
	require.Equal(t, expPrivKey, actualPrivKey)

	msg := []byte("hello world")
	// Sanity check that the keypair works with the ed25519 library
	sig := ed25519.Sign(ed25519.PrivateKey(keypair.Private), msg)
	require.True(t, ed25519.Verify(ed25519.PublicKey(keypair.Public), msg, sig))

	// create another account from the same seed
	accts2, err := accountsFromMaster(master, goodSeed, 1)
	require.NoError(t, err)
	require.Len(t, accts2, 1)
	require.Equal(t, keypair.Public, accts2[0].Public)
	require.Equal(t, keypair.Private, accts2[0].Private)
}

func TestWalletFromNewMnemonic(t *testing.T) {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w, err := NewMultiWalletFromMnemonic(mnemonic, 1)

	require.NoError(t, err)
	require.NotNil(t, w)
	require.Equal(t, mnemonic, w.Mnemonic())
}

func TestWalletFromGivenMnemonic(t *testing.T) {
	mnemonic := "film theme cheese broken kingdom destroy inch ready wear inspire shove pudding"
	w, err := NewMultiWalletFromMnemonic(mnemonic, 1)
	require.NoError(t, err)
	expPubKey := "de30fc9b812248583da6259433626fcdd2cb5ce589b00047b81e127950b9bca6"
	expPrivKey := "cd85df73aa3bc31de2f0b69bb1421df7eb0cdca7cb170a457869ab337749dae1de30fc9b812248583da6259433626fcdd2cb5ce589b00047b81e127950b9bca6"

	actualPubKey := hex.EncodeToString(w.Secrets.Accounts[0].Public)
	actualPrivKey := hex.EncodeToString(w.Secrets.Accounts[0].Private)
	require.Equal(t, expPubKey, actualPubKey)
	require.Equal(t, expPrivKey, actualPrivKey)

	msg := []byte("hello world")

	// Sanity check that the keypair works with the standard ed25519 library
	sig := ed25519.Sign(ed25519.PrivateKey(w.Secrets.Accounts[0].Private), msg)
	require.True(t, ed25519.Verify(ed25519.PublicKey(w.Secrets.Accounts[0].Public), msg, sig))
}

func TestKeysInWalletMaintainExpectedPath(t *testing.T) {
	n := 100
	w, err := NewMultiWalletRandomMnemonic(n)
	require.NoError(t, err)

	for i := 0; i < n; i++ {
		expectedPath := fmt.Sprintf("%s/0'/0'/%d'", Bip44Prefix, i)
		path := w.Secrets.Accounts[i].Path
		require.Equal(t, expectedPath, HDPathToString(path))
	}
}

func TestMnemonicWhitespace(t *testing.T) {
	mnemonics := []string{
		"film  theme cheese broken kingdom destroy inch ready wear inspire shove pudding",
		"film theme cheese broken kingdom destroy inch ready wear  inspire shove pudding",
		"film theme cheese broken kingdom destroy inch ready wear\ninspire shove pudding",
		"film theme cheese broken kingdom destroy inch ready wear inspire shove pudding\t",
		" film theme cheese broken kingdom destroy inch ready wear inspire shove pudding",
		"film theme cheese broken kingdom destroy inch ready wear inspire shove pudding ",
		"film  theme  cheese  broken  kingdom  destroy  inch  ready  wear  inspire  shove  pudding",
	}
	for _, m := range mnemonics {
		_, err := NewMultiWalletFromMnemonic(m, 1)
		require.Equal(t, errWhitespace, err, "expected whitespace error in mnemonic")
	}
}
