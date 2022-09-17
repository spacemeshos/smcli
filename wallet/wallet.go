package wallet

import (
	"crypto/sha512"
	"sync"

	"github.com/tyler-smith/go-bip39"
	"github.com/xdg-go/pbkdf2"
)

// Wallet is a collection of accounts.
type Wallet struct {
	mnemonic string
	master   *BIP32EDKeyPair
	purpose  *BIP32EDKeyPair
	coinType *BIP32EDKeyPair
	account  map[uint32]*Account
	lock     sync.Mutex
}

// WalletFromMnemonic creates a new wallet from the given mnemonic.
// the mnemonic must be a valid bip39 mnemonic.
func WalletFromMnemonic(mnemonic string) *Wallet {
	if !bip39.IsMnemonicValid(mnemonic) {
		panic("invalid mnemonic")
	}
	mBytes := []byte(mnemonic)
	salt := []byte("The lottery is a tax on people who are bad at math.")
	if len(salt) < Pdkdf2SaltBytesLen {
		panic("salt too short")
	}
	seed := pbkdf2.Key([]byte(mnemonic), append([]byte(salt), mBytes...),
		Pbkdf2Itterations, Pbkdf2KeyBytesLen, sha512.New)
	masterKeyPair, _ := NewMasterBIP32EDKeyPair(seed)

	w := &Wallet{
		mnemonic: mnemonic,
		master:   masterKeyPair,
		purpose:  nil,
		coinType: nil,
		account:  make(map[uint32]*Account),
	}

	purposeKeyPair, _ := w.master.NewChildKeyPair(BIP44Purpose())
	w.purpose = &purposeKeyPair
	coinTypeKeyPair, _ := purposeKeyPair.NewChildKeyPair(BIP44SpacemeshCoinType())
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
	accntNum := uint32(len(w.account))
	accountKeyPair, _ := w.coinType.NewChildKeyPair(BIP44Account(accntNum))
	w.account[accntNum] = &Account{
		Name:    name,
		KeyPair: accountKeyPair,
		Chain:   make([]*Chain, 0),
	}
	// Initialize 0th chain and index.
	chainNum := uint32(0)
	chainKeyPair, _ := accountKeyPair.NewChildKeyPair(BIP44HardenedChain(chainNum))
	w.account[accntNum].Chain = append(w.account[accntNum].Chain, &Chain{
		KeyPair: chainKeyPair,
		Index:   make([]*BIP32EDKeyPair, 0),
	})
	indexNum := uint32(0)
	indexKeyPair, _ := chainKeyPair.NewChildKeyPair(BIP44HardenedAccountIndex(indexNum))
	w.account[accntNum].Chain[chainNum].Index = append(
		w.account[accntNum].Chain[chainNum].Index,
		&indexKeyPair,
	)
	return w.account[accntNum]
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
