package wallet_test

import (
	"crypto/sha512"
	"testing"

	"github.com/spacemeshos/ed25519-recovery"
	"github.com/spacemeshos/smcli/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/xdg-go/pbkdf2"
)

func generateTestMasterKeyPair() (*wallet.BIP32EDKeyPair, error) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	password := "Winning together!"
	seed := pbkdf2.Key([]byte(mnemonic), []byte("mnemonic"+password), 2048, 32, sha512.New)
	return wallet.NewMasterBIP32EDKeyPair(seed)
}

func TestNewMasterBIP32EDKeyPair(t *testing.T) {
	masterKeyPair, err := generateTestMasterKeyPair()
	assert.NoError(t, err)
	assert.NotEmpty(t, masterKeyPair)

	msg := []byte("master test")
	sig := ed25519.Sign(masterKeyPair.Private, msg)
	valid := ed25519.Verify(masterKeyPair.Public, msg, sig)
	assert.True(t, valid)

	extracted_pub, err := ed25519.ExtractPublicKey(msg, sig)
	assert.NoError(t, err)
	assert.Equal(t, masterKeyPair.Public, extracted_pub)
}

func TestNewChildKeyPair(t *testing.T) {
	masterKeyPair, _ := generateTestMasterKeyPair()

	childKeyPair, err := masterKeyPair.NewChildKeyPair(wallet.BIP44Purpose())

	assert.NoError(t, err)
	assert.NotEmpty(t, childKeyPair)

	msg := []byte("child test")
	sig := ed25519.Sign(childKeyPair.Private, msg)
	valid := ed25519.Verify(childKeyPair.Public, msg, sig)
	assert.True(t, valid)

	extracted_pub, err := ed25519.ExtractPublicKey(msg, sig)
	assert.NoError(t, err)
	assert.Equal(t, childKeyPair.Public, extracted_pub)
}
