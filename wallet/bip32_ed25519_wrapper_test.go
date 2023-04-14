package wallet_test

import (
	"crypto/ed25519"
	"crypto/sha512"
	"testing"

	"github.com/spacemeshos/smcli/wallet"
	"github.com/stretchr/testify/require"
	"github.com/xdg-go/pbkdf2"
)

func generateTestMasterKeyPair() (*wallet.EDKeyPair, error) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	password := "Winning together!"
	seed := pbkdf2.Key([]byte(mnemonic), []byte("mnemonic"+password), 2048, 32, sha512.New)
	return wallet.NewMasterKeyPair(seed)
}

func TestNewMasterBIP32EDKeyPair(t *testing.T) {
	// first iteration
	keyPair1, err := generateTestMasterKeyPair()
	require.NoError(t, err)
	require.NotEmpty(t, keyPair1)

	// second iteration
	keyPair2, err := generateTestMasterKeyPair()
	require.NoError(t, err)
	require.NotEmpty(t, keyPair2)

	msg := []byte("master test")

	// Testing the private key signature generated from the first iteration and verifying with public key from the second iteration
	sig1 := ed25519.Sign(ed25519.PrivateKey(keyPair1.Private), msg)
	valid1 := ed25519.Verify(ed25519.PublicKey(keyPair2.Public), msg, sig1)
	require.True(t, valid1)

	// Same test with swapped private and public key
	sig2 := ed25519.Sign(ed25519.PrivateKey(keyPair2.Private), msg)
	valid2 := ed25519.Verify(ed25519.PublicKey(keyPair1.Public), msg, sig2)
	require.True(t, valid2)
}

//func TestNewChildKeyPair(t *testing.T) {
//	masterKeyPair, _ := generateTestMasterKeyPair()
//
//	childKeyPair, err := masterKeyPair.NewChildKeyPair(wallet.BIP44Purpose())
//
//	require.NoError(t, err)
//	require.NotEmpty(t, childKeyPair)
//
//	msg := []byte("child test")
//	sig := ed25519.Sign(childKeyPair.Private, msg)
//	valid := ed25519.Verify(childKeyPair.Public, msg, sig)
//	require.True(t, valid)
//
//	extractedPub, err := ed25519.ExtractPublicKey(msg, sig)
//	require.Equal(t, childKeyPair.Public, extractedPub)
//}
