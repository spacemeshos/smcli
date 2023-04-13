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
	masterKeyPair, err := generateTestMasterKeyPair()
	require.NoError(t, err)
	require.NotEmpty(t, masterKeyPair)

	msg := []byte("master test")
	sig := ed25519.Sign(ed25519.PrivateKey(masterKeyPair.Private), msg)
	valid := ed25519.Verify(ed25519.PublicKey(masterKeyPair.Public), msg, sig)
	require.True(t, valid)
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
