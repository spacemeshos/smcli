package wallet

import (
	"crypto/ed25519"
	"github.com/tyler-smith/go-bip39"

	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
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
	seed := []byte("spacemesh is the best blockchain")
	accts, err := accountsFromSeed(seed, 1)
	require.NoError(t, err)
	require.Len(t, accts, 1)
	keypair := accts[0]

	expPubKey :=
		"150509c0c961434e0923fa3837e24ccec114d7dca3601b8b10ecc17a7251760d"
	expPrivKey :=
		"eeca7f1df93f2e79b506efb839510e910d1e6066d443bab36917aa47421e921c150509c0c961434e0923fa3837e24ccec114d7dca3601b8b10ecc17a7251760d"

	actualPubKey := hex.EncodeToString(keypair.Public)
	actualPrivKey := hex.EncodeToString(keypair.Private)
	require.Equal(t, expPubKey, actualPubKey)
	require.Equal(t, expPrivKey, actualPrivKey)

	msg := []byte("hello world")
	// Sanity check that the keypair works with the ed25519 library
	sig := ed25519.Sign(ed25519.PrivateKey(keypair.Private), msg)
	require.True(t, ed25519.Verify(ed25519.PublicKey(keypair.Public), msg, sig))

	// create another account from the same seed
	accts2, err := accountsFromSeed(seed, 1)
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
	expPubKey :=
		"0f4e4d7ba460a187d11b14e8fed05c5ebc7c0bc06260bf457cc6c54aa525c4a4"
	expPrivKey :=
		"7f84bb162a01010832407fda8ab53749b988e8f5c530f169db17aa0b1a0d23260f4e4d7ba460a187d11b14e8fed05c5ebc7c0bc06260bf457cc6c54aa525c4a4"

	actualPubKey := hex.EncodeToString(w.Secrets.Accounts[0].Public)
	actualPrivKey := hex.EncodeToString(w.Secrets.Accounts[0].Private)
	require.Equal(t, expPubKey, actualPubKey)
	require.Equal(t, expPrivKey, actualPrivKey)

	msg := []byte("hello world")

	// Sanity check that the keypair works with the standard ed25519 library
	sig := ed25519.Sign(ed25519.PrivateKey(w.Secrets.Accounts[0].Private), msg)
	require.True(t, ed25519.Verify(ed25519.PublicKey(w.Secrets.Accounts[0].Public), msg, sig))
}

// TODO: re-add path support, then re-enable.
//func TestKeysInWalletMaintainExpectedPath(t *testing.T) {
//	n := 100
//	w, err := NewMultiWalletRandomMnemonic(n)
//	require.NoError(t, err)
//
//	for i := 0; i < n; i++ {
//		expectedPath := fmt.Sprintf("%s'/%d'/0'/0'", Bip44Prefix, i)
//		path := w.Secrets.Accounts[i].Path
//		require.Equal(t, expectedPath, path)
//	}
//}
