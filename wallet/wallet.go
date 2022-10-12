package wallet

import (
	"sync"

	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
)

// Wallet is a collection of accounts.
type Wallet struct {
	mnemonic            string
	master              *BIP32EDKeyPair
	purpose             *BIP32EDKeyPair
	coinType            *BIP32EDKeyPair
	account             map[uint32]*Account
	numHardenedAccounts uint32
	lock                sync.Mutex
}

// WalletFromMnemonic creates a new wallet from the given mnemonic.
// the mnemonic must be a valid bip39 mnemonic.
func WalletFromMnemonic(mnemonic string) *Wallet {
	if !bip39.IsMnemonicValid(mnemonic) {
		panic("invalid mnemonic")
	}
	// mBytes := []byte(mnemonic)
	// salt := []byte("The lottery is a tax on people who are bad at math.")
	// if len(salt) < Pdkdf2SaltBytesLen {
	// panic("salt too short")
	// }
	// seed := pbkdf2.Key([]byte(mnemonic), append([]byte(salt), mBytes...),
	// 	Pbkdf2Itterations, EncKeyLen, sha512.New)
	// TODO: add option for user to provide passphrase
	seed := bip39.NewSeed(mnemonic, "")
	// Arbitrarily taking the first 32 bytes as the seed for the private key.
	// Not sure if this is the right thing to do. Or if it matters at all...
	masterKeyPair, err := NewMasterBIP32EDKeyPair(seed[32:])
	cobra.CheckErr(err)

	w := &Wallet{
		mnemonic: mnemonic,
		master:   masterKeyPair,
		purpose:  nil,
		coinType: nil,
		account:  make(map[uint32]*Account),
	}

	purposeKeyPair, err := w.master.NewChildKeyPair(BIP44Purpose())
	cobra.CheckErr(err)
	w.purpose = &purposeKeyPair
	coinTypeKeyPair, err := purposeKeyPair.NewChildKeyPair(BIP44SpacemeshCoinType())
	cobra.CheckErr(err)
	w.coinType = &coinTypeKeyPair
	w.newAccountLocked("default")
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
func (w *Wallet) newAccountLocked(name string) *Account {
	accntNum := w.numHardenedAccounts
	accountKeyPair, _ := w.coinType.NewChildKeyPair(BIP44Account(accntNum))
	w.account[accntNum] = &Account{
		Name:    name,
		keyPair: accountKeyPair,
		chains:  make(map[uint32]*chain, 0),
		lock:    &sync.Mutex{},
	}
	w.numHardenedAccounts = accntNum + 1
	// Initialize 0th chain and index.
	chainNum := 0 | BIP32HardenedKeyStart
	w.account[accntNum].NextAddress(chainNum)
	return w.account[accntNum]
}

// KeyPair returns the key pair for the given HDPath.
// It will compute it every time from the mater key.
func (w *Wallet) ComputeKeyPair(path HDPath) BIP32EDKeyPair {
	w.lock.Lock()
	defer w.lock.Unlock()
	if path.Purpose() != BIP44Purpose() {
		panic("invalid path: purpose is not BIP44 (44)")
	}
	if path.CoinType() != BIP44SpacemeshCoinType() {
		panic("invalid path: coin type is not spacemesh (540)")
	}
	ct, err := w.master.NewChildKeyPair(BIP44Account(path.CoinType()))
	cobra.CheckErr(err)
	acct, err := ct.NewChildKeyPair(BIP44Account(path.Account()))
	cobra.CheckErr(err)
	chain, err := acct.NewChildKeyPair(BIP44Account(path.Chain()))
	cobra.CheckErr(err)
	index, err := chain.NewChildKeyPair(BIP44Account(path.Index()))
	cobra.CheckErr(err)
	return index
}

// Account returns the account with the given name.
// If the account does not exist, it is created.
func (w *Wallet) Account(name string) *Account {
	w.lock.Lock()
	defer w.lock.Unlock()
	for _, acct := range w.account {
		if acct.Name == name {
			return acct
		}
	}
	return w.newAccountLocked(name)
}
