package wallet

import (
	"crypto/ed25519"
	"fmt"
	"github.com/tyler-smith/go-bip39"

	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

const Bip44Prefix = "m/44'/540'"

func TestNewWallet(t *testing.T) {
	w, err := NewMultiWalletRandomMnemonic(1)
	require.NoError(t, err)
	require.NotNil(t, w)
}

func TestAccountFromSeed(t *testing.T) {
	seed := []byte("spacemesh is the best blockchain")
	accts, err := accountsFromSeed(seed, 1)
	require.NoError(t, err)
	require.Len(t, accts, 1)
	keypair := accts[0]

	expPubKey :=
		"680e002010e2a12d17922bade5c3c25d54a911abd41f48e16b3bbe0c83df96a9"
	expPrivKey :=
		"73706163656d65736820697320746865206265737420626c6f636b636861696e680e002010e2a12d17922bade5c3c25d54a911abd41f48e16b3bbe0c83df96a9"

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
		"205d8d4e458b163d5ba15ac712951d5659cc51379e7e0ad13acc97303aa85093"
	expPrivKey :=
		"669d091195f950e6255a2e8778eea7be4f7a66afe855957404ec1520c8a11ff1205d8d4e458b163d5ba15ac712951d5659cc51379e7e0ad13acc97303aa85093"

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
		expectedPath := fmt.Sprintf("%s'/%d'/0'/0'", Bip44Prefix, i)
		path := w.Secrets.Accounts[i].Path
		require.Equal(t, expectedPath, path)
	}
}
