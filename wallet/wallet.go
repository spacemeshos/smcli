package wallet

import (
	"fmt"

	"github.com/spacemeshos/ed25519"
	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
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
func (w *Wallet) Mnemonic() string {
	return w.mnemonic
}
func (w *Wallet) ToBytes() []byte {
	return []byte(w.mnemonic)
}

// KeyPair returns the key pair for the given HDPath.
// It will compute it every time from the mater key.
func (w *Wallet) ComputeKeyPair(path HDPath) (*BIP32EDKeyPair, error) {
	if !IsPathCompletelyHardened(path) {
		return nil, fmt.Errorf("unhardened keys aren't supported: path must be completely hardened" +
			" until the new Child Key Derivation function is implemented")
	}
	if path.Purpose() != BIP44Purpose() {
		panic("invalid path: purpose is not BIP44 (44)")
	}
	purpose, err := w.masterKeyPair.NewChildKeyPair(BIP44Purpose())
	cobra.CheckErr(err)
	if path.CoinType() != BIP44SpacemeshCoinType() {
		panic("invalid path: coin type is not spacemesh (540)")
	}
	cointype, err := purpose.NewChildKeyPair(BIP44SpacemeshCoinType())

	acct, err := cointype.NewChildKeyPair(path.Account())
	cobra.CheckErr(err)
	chain, err := acct.NewChildKeyPair(path.Chain())
	cobra.CheckErr(err)
	index, err := chain.NewChildKeyPair(path.Index())
	cobra.CheckErr(err)
	return &index, nil
}
