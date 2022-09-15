package wallet

import (
	"bytes"
	"encoding/gob"

	"github.com/tyler-smith/go-bip39"
)

// Wallet is a collection of accounts.
type Wallet struct {
	mnemonic string
	master   *BIP32EDKeyPair
	purpose  *BIP32EDKeyPair
	coinType *BIP32EDKeyPair
	account  map[uint32]*Account
	chain    map[uint32]*BIP32EDKeyPair
	index    map[uint32]*BIP32EDKeyPair
}

func NewWallet() *Wallet {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	seed := bip39.NewSeed(mnemonic, "Winning together!")
	masterKeyPair, _ := NewMasterBIP32EDKeyPair(seed)

	w := &Wallet{
		mnemonic: mnemonic,
		master:   &masterKeyPair,
		purpose:  nil,
		coinType: nil,
		account:  make(map[uint32]*Account),
		chain:    make(map[uint32]*BIP32EDKeyPair),
		index:    make(map[uint32]*BIP32EDKeyPair),
	}

	purposeKeyPair, _ := w.master.NewChildKeyPair(BIP44Purpose())
	w.purpose = &purposeKeyPair
	coinTypeKeyPair, _ := purposeKeyPair.NewChildKeyPair(BIP44SpacemeshCoinType())
	w.coinType = &coinTypeKeyPair
	w.NewAccount("default")
	return w
}
func (w *Wallet) ToBytes() []byte {
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(w)
	return buf.Bytes()
}
func (w *Wallet) NewAccount(name string) *Account {
	nextAccountNumber := uint32(len(w.account))
	accountKeyPair, _ := w.coinType.NewChildKeyPair(BIP44Account(nextAccountNumber))
	w.account[nextAccountNumber] = &Account{
		Name:    name,
		KeyPair: accountKeyPair,
	}
	return w.account[nextAccountNumber]
}
