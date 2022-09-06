package wallet

import (
	"crypto/rand"
	"crypto/sha512"

	"github.com/spf13/cobra"
	"github.com/xdg-go/pbkdf2"
)

// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#pbkdf2
const pbkdf2Itterations = 120000
const pbkdf2KeyBytesLen = 32
const pdkdf2SaltBytesLen = 16

var pbkdf2HashFunc = sha512.New

type WalletKeyOpt func(*WalletKey)
type WalletKey struct {
	key  []byte
	salt []byte
}

func NewWalletKey(opts ...WalletKeyOpt) *WalletKey {
	w := &WalletKey{}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

func WithPassword(password string) WalletKeyOpt {
	return func(k *WalletKey) {
		k.salt = make([]byte, pdkdf2SaltBytesLen)
		_, err := rand.Read(k.salt)
		cobra.CheckErr(err)
		k.key = pbkdf2.Key([]byte(password), k.salt,
			pbkdf2Itterations,
			pbkdf2KeyBytesLen,
			pbkdf2HashFunc,
		)
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

// Wallet is a collection of accounts.
type Wallet struct {
	accounts []Account `json:"accounts"`
}

func NewWallet() *Wallet {
	w := &Wallet{
		accounts: make([]Account, 0),
	}
	return w
}
func (w *Wallet) AddAccount(a Account) {
	w.accounts = append(w.accounts, a)
}

type Account struct {
	Id           string `json:"id"`
	FriendlyName string `json:"friendly_name"`
	PrivateKey   string `json:"private_key"`
	PublicKey    string `json:"public_key"`
}

func newAccount() *Account {
	acct := &Account{}

	return acct
}
