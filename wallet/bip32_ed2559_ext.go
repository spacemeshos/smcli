package wallet

import (
	"errors"

	"github.com/spacemeshos/ed25519"
)

// Function names inspired by https://github.com/tyler-smith/go-bip32/blob/master/bip32.go
// This is a POC implementation that relies on the key derivation from the
// spacemeshos/ed25519 package.
// We assume all keys are hardened. In the future, we should use the child key
// derivation function from the BIP32-ED25519 paper
// https://github.com/LedgerHQ/orakolo/blob/master/papers/Ed25519_BIP%20Final.pdf
// so that we can be compatible with other wallets and so that spacemesh can be stored
// in a wallet containing other coins using the same derivation scheme.

var ErrInvalidSeed = errors.New("invalid seed length")
var ErrNotHardened = errors.New("child index must be hardened")

type BIP32EDKeyPair struct {
	Private ed25519.PrivateKey
	Public  ed25519.PublicKey
	Path    [][]uint32 // we assume the prefix is m/ and all keys are hardened
	Salt    []byte
}

func NewMasterBIP32EDKeyPair(seed []byte) (BIP32EDKeyPair, error) {
	if len(seed) != ed25519.SeedSize {
		return BIP32EDKeyPair{}, ErrInvalidSeed
	}
	privKey := ed25519.NewKeyFromSeed(seed)

	return BIP32EDKeyPair{
		Private: privKey,
		Public:  privKey.Public().(ed25519.PublicKey),
		Path:    [][]uint32{},
		Salt:    []byte("Spacemesh NaCl"),
	}, nil
}

func (kp *BIP32EDKeyPair) NewChildKeyPair(childIdx uint32) (BIP32EDKeyPair, error) {
	if childIdx < 0x80000000 {
		return BIP32EDKeyPair{}, ErrNotHardened
	}
	privKey := ed25519.NewDerivedKeyFromSeed(
		kp.Private.Seed(),
		uint64(childIdx),
		kp.Salt,
	)
	return BIP32EDKeyPair{
		Private: privKey,
		Public:  privKey.Public().(ed25519.PublicKey),
		Path:    append(kp.Path, []uint32{childIdx}),
	}, nil
}