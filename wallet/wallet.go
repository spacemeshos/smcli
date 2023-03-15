package wallet

import (
	"fmt"
	"github.com/tyler-smith/go-bip39"

	"github.com/spacemeshos/ed25519"
	"github.com/spf13/cobra"
)

// Wallet is a collection of accounts.
type Wallet struct {
	mnemonic      string
	masterKeyPair *BIP32EDKeyPair
}

// WalletFromMnemonic creates a new wallet from the given mnemonic.
// the mnemonic must be a valid bip39 mnemonic.
func WalletFromMnemonic(mnemonic string) *Wallet {
	if !bip39.IsMnemonicValid(mnemonic) {
		panic("invalid mnemonic")
	}
	// TODO: add option for user to provide passphrase
	seed := bip39.NewSeed(mnemonic, "")
	// Arbitrarily taking the first 32 bytes as the seed for the private key
	// because spacemeshos/ed25519 gets angry if it gets all 64 bytes.
	// Not sure if this is the correct approach.
	masterKeyPair, err := NewMasterBIP32EDKeyPair(seed[ed25519.SeedSize:])
	cobra.CheckErr(err)

	w := &Wallet{
		mnemonic:      mnemonic,
		masterKeyPair: masterKeyPair,
	}
	return w
}
func (w *Wallet) Salt() []byte {
	return w.masterKeyPair.Salt
}
func (w *Wallet) Mnemonic() string {
	return w.mnemonic
}
func (w *Wallet) ToBytes() []byte {
	return []byte(w.mnemonic)
}

// KeyPair returns the key pair for the given HDPath.
// It will compute it every time from the master key.
// If the path is empty, it will return the master key pair.
func (w *Wallet) ComputeKeyPair(path HDPath) (*BIP32EDKeyPair, error) {
	if !IsPathCompletelyHardened(path) {
		return nil, fmt.Errorf("unhardened keys aren't supported: path must be completely hardened" +
			" until a homomorphic Child Key Derivation function is implemented")
	}

	keypair := w.masterKeyPair

	for i, childKeyIndex := range path {
		if i == HDPurposeSegment && childKeyIndex != BIP44Purpose() {
			return nil, fmt.Errorf("invalid purpose: expected %d, got %d", BIP44Purpose(), childKeyIndex)
		}
		if i == HDCoinTypeSegment && childKeyIndex != BIP44SpacemeshCoinType() {
			return nil, fmt.Errorf("invalid coin type: expected %d, got %d", BIP44SpacemeshCoinType(), childKeyIndex)
		}
		kp, err := keypair.NewChildKeyPair(childKeyIndex)
		if err != nil {
			return nil, err
		}
		keypair = kp
	}
	return keypair, nil
}
