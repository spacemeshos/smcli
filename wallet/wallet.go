package wallet

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
)

// Wallet is a collection of accounts.
type Wallet struct {
	mnemonic      string
	masterKeyPair *BIP32EDKeyPair
	lock          sync.Mutex
}

// WalletFromMnemonic creates a new wallet from the given mnemonic.
// the mnemonic must be a valid bip39 mnemonic.
func WalletFromMnemonic(mnemonic string) *Wallet {
	if !bip39.IsMnemonicValid(mnemonic) {
		panic("invalid mnemonic")
	}
	// TODO: add option for user to provide passphrase
	seed := bip39.NewSeed(mnemonic, "")
	// Arbitrarily taking the first 32 bytes as the seed for the private key.
	// Not sure if this is the right thing to do. Or if it matters at all...
	masterKeyPair, err := NewMasterBIP32EDKeyPair(seed[32:])
	cobra.CheckErr(err)

	w := &Wallet{
		mnemonic:      mnemonic,
		masterKeyPair: masterKeyPair,
	}
	return w
}
func (w *Wallet) Mnemonic() string {
	w.lock.Lock()
	defer w.lock.Unlock()
	return w.mnemonic
}
func (w *Wallet) ToBytes() []byte {
	panic("not implemented")
}

// KeyPair returns the key pair for the given HDPath.
// It will compute it every time from the mater key.
func (w *Wallet) ComputeKeyPair(path HDPath) (*BIP32EDKeyPair, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
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
