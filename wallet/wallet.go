package wallet

import (
	"crypto"
)

type WalletKeyOpt func(*WalletKey)
type WalletKey struct {
	password string
}

func NewWalletKey(opts ...WalletKeyOpt) *WalletKey {
	return &WalletKey{}
}
func WithPassword(password string) WalletKeyOpt {
	return func(k *WalletKey) {
		k.password = password
	}
}
func (k *WalletKey) encrypt() ([]byte, error) {
	panic("implement me")
}

type WalletOpener interface {
	Open(path string) (*Wallet, error)
}
type WalletExporter interface {
	Export(path string) error
}

type WalletStore struct {
	wk WalletKey
}

func NewWalletStore(wk *WalletKey, opts ...WalletStoreOpt) *WalletStore {
	return &WalletStore{
		wk: *wk,
	}
}

type WalletStoreOpt func(*WalletStore)

func (s *WalletStore) Open(path string, w *Wallet) error {
	panic("implement me")
}

func (s *WalletStore) Export(path string, w *Wallet) error {
	panic("implement me")
}

// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#pbkdf2

// Wallet is a collection of accounts.
type Wallet struct {
	accounts map[string]*Account
}

func NewWallet() *Wallet {
	w := &Wallet{
		accounts: make(map[string]*Account),
	}
	return w
}

type Account struct {
	friendlyName string
	signer       crypto.Signer
}

func newAccount() *Account {
	acct := &Account{}

	return acct
}
