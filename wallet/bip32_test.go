package wallet

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"

	"github.com/spacemeshos/smkeys/bip32"
	"github.com/stretchr/testify/require"
)

var goodSeed = []byte("abandon abandon abandon abandon abandon abandon abandon abandon ")

func TestNonHardenedPath(t *testing.T) {
	path1Str := "m/44'/540'/0'/0'/0'"
	path1Hd, err := StringToHDPath(path1Str)
	require.NoError(t, err)
	require.True(t, IsPathCompletelyHardened(path1Hd))

	path2Str := "m/44'/540'/0'/0'/0"
	path2Hd, err := StringToHDPath(path2Str)
	require.NoError(t, err)
	require.False(t, IsPathCompletelyHardened(path2Hd))
}

// Test that path string produces expected path and vice-versa.
func TestPath(t *testing.T) {
	path1Str := "m/44'/540'/0'/0'/0'"
	path1Hd, err := StringToHDPath(path1Str)
	require.NoError(t, err)
	path := DefaultPath()
	path2Hd := path.Extend(BIP44HardenedAccountIndex(0))
	path2Str := HDPathToString(path2Hd)
	require.Equal(t, path1Str, path2Str)
	require.Equal(t, path1Hd, path2Hd)
}

// Test deriving a child keypair.
func TestChildKeyPair(t *testing.T) {
	defaultPath := DefaultPath()
	path := defaultPath.Extend(BIP44HardenedAccountIndex(0))

	// generate first keypair
	masterKeyPair, err := NewMasterKeyPair(goodSeed)
	require.NoError(t, err)
	childKeyPair1, err := masterKeyPair.NewChildKeyPair(goodSeed, 0)
	require.Equal(t, path, childKeyPair1.Path)
	require.Len(t, childKeyPair1.Private, ed25519.PrivateKeySize)
	require.Len(t, childKeyPair1.Public, ed25519.PublicKeySize)
	require.NoError(t, err)
	require.NotEmpty(t, childKeyPair1)

	// test signing
	msg := []byte("child test")
	sig := ed25519.Sign(ed25519.PrivateKey(childKeyPair1.Private), msg)
	valid := ed25519.Verify(ed25519.PublicKey(childKeyPair1.Public), msg, sig)
	require.True(t, valid)

	// generate second keypair and check lengths
	childKeyPair2, err := bip32.Derive(HDPathToString(path), goodSeed)
	require.NoError(t, err)
	require.Len(t, childKeyPair2, ed25519.PrivateKeySize)
	privkey2 := PrivateKey(childKeyPair2[:])
	require.Len(t, privkey2, ed25519.PrivateKeySize)
	edpubkey2 := ed25519.PrivateKey(privkey2).Public().(ed25519.PublicKey)
	require.Len(t, edpubkey2, ed25519.PublicKeySize)
	pubkey2 := PublicKey(edpubkey2)
	require.Len(t, pubkey2, ed25519.PublicKeySize)

	// make sure they agree
	require.Equal(t, "feae6977b42bf3441d04314d09c72c5d6f2d1cb4bf94834680785b819f8738dd", hex.EncodeToString(pubkey2))
	require.Equal(t, hex.EncodeToString(childKeyPair1.Public), hex.EncodeToString(pubkey2))
	//nolint:lll
	require.Equal(t, "05fe9affa5562ca833faf3803ce5f6f7615d3c37c4a27903492027f6853e486dfeae6977b42bf3441d04314d09c72c5d6f2d1cb4bf94834680785b819f8738dd", hex.EncodeToString(privkey2))
	require.Equal(t, hex.EncodeToString(childKeyPair1.Private), hex.EncodeToString(privkey2))
}
