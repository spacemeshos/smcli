package wallet

import (
	"errors"

	"github.com/spacemeshos/ed25519"
)

// Function names inspired by https://github.com/tyler-smith/go-bip32/blob/master/bip32.go
// This is a POC implementation that relies on the key derivation from the
// spacemeshos/ed25519 package.
// We assume all keys are hardened, and that the path is of the form m/0'/i'...
// In the future, we should use the key derivation function from the
// BIP32-ED25519 paper https://github.com/LedgerHQ/orakolo/blob/master/papers/Ed25519_BIP%20Final.pdf
// so that we can be compatible with other wallets and so that spacemesh can be stored
// in a wallet containing other coins using the same derivation scheme.

var ErrInvalidSeed = errors.New("invalid seed length")
var ErrNotHardened = errors.New("child index must be hardened")

type KeyPair struct {
	Private ed25519.PrivateKey
	Public  ed25519.PublicKey
	Path    [][]uint32 // we assume the prefix is m/ and all keys are hardened
	Salt    []byte
}

func NewMasterKey(seed []byte) (KeyPair, error) {
	if len(seed) != ed25519.SeedSize {
		return KeyPair{}, ErrInvalidSeed
	}
	privKey := ed25519.NewKeyFromSeed(seed)

	return KeyPair{
		Private: privKey,
		Public:  privKey.Public().(ed25519.PublicKey),
		Path:    [][]uint32{},
	}, nil
}

func (kp *KeyPair) NewChildKey(childIdx uint32) (KeyPair, error) {
	if childIdx < 0x80000000 {
		return KeyPair{}, ErrNotHardened
	}
	privKey := ed25519.NewDerivedKeyFromSeed(
		kp.Private.Seed(),
		uint64(childIdx),
		kp.Salt,
	)
	return KeyPair{
		Private: privKey,
		Public:  privKey.Public().(ed25519.PublicKey),
		Path:    append(kp.Path, []uint32{childIdx}),
	}, nil
}
