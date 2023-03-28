package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"io"
	"os"

	"github.com/btcsuite/btcutil/base58"
	"github.com/spf13/cobra"
	"github.com/xdg-go/pbkdf2"
)

const EncKeyLen = 32

// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#pbkdf2
// TODO: should this be increased to 210,000 per the above link?
const Pbkdf2Iterations = 120000
const Pdkdf2SaltBytesLen = 16

var Pbkdf2HashFunc = sha512.New

// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#argon2id
//var Argon2idParams = &argon2id.Params{
//	Memory:      64 * 1024,
//	Iterations:  3,
//	Parallelism: 4,
//	SaltLength:  16,
//	KeyLength:   EncKeyLen,
//}

type WalletKeyOpt func(*WalletKey)
type WalletKey struct {
	key  []byte
	salt []byte
}

func NewKey(opts ...WalletKeyOpt) *WalletKey {
	w := &WalletKey{}
	for _, opt := range opts {
		opt(w)
	}
	if w.key == nil {
		panic("Some form of key generation method must be provided. Try WithXXXPassword.")
	}

	return w
}

// Complies with FIPS-140
func WithPbkdf2Password(password string) WalletKeyOpt {
	return func(k *WalletKey) {
		k.salt = make([]byte, Pdkdf2SaltBytesLen)
		_, err := rand.Read(k.salt)
		cobra.CheckErr(err)
		k.key = pbkdf2.Key([]byte(password), k.salt,
			Pbkdf2Iterations,
			EncKeyLen,
			Pbkdf2HashFunc,
		)
	}
}

// Is better, but not FIPS-140 compliant.
//func WithArgon2idPassword(password string) WalletKeyOpt {
//	return func(k *WalletKey) {
//		hash, err := argon2id.CreateHash(password, Argon2idParams)
//		cobra.CheckErr(err)
//		_, salt, key, err := argon2id.DecodeHash(hash)
//		cobra.CheckErr(err)
//		k.salt = salt
//		k.key = key
//	}
//}

// https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html#71-encryption-types-to-use
func (k *WalletKey) encrypt(plaintext []byte) (ciphertext []byte, nonce []byte) {
	block, err := aes.NewCipher(k.key)
	cobra.CheckErr(err)
	// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
	// TODO: Remind user to not use the same password to encrypt 4 billion times.
	nonce = make([]byte, 12)
	_, err = io.ReadFull(rand.Reader, nonce)
	cobra.CheckErr(err)
	aesgcm, err := cipher.NewGCM(block)
	cobra.CheckErr(err)
	ciphertext = aesgcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce
}
func (k *WalletKey) decrypt(ciphertext []byte, nonce []byte) (plaintext []byte) {
	block, err := aes.NewCipher(k.key)
	cobra.CheckErr(err)
	aesgcm, err := cipher.NewGCM(block)
	cobra.CheckErr(err)
	plaintext, err = aesgcm.Open(nil, nonce, ciphertext, nil)
	cobra.CheckErr(err)

	return plaintext
}

type WalletOpener interface {
	Open(path string) (*Wallet, error)
}
type WalletExporter interface {
	Export(path string) error
}

type ExportableWallet struct {
	// EncryptedWallet is the encrypted wallet data in base58 encoding.
	EncryptedWallet string `json:"encrypted_wallet"`
	// Salt is the salt used to derived the wallet's encryption key in base58 encoding.
	Salt string `json:"salt"`
	// Nonce is the nonce used to encrypt the wallet in base58 encoding.
	Nonce []byte `json:"nonce"`
}

type WalletStore struct {
	wk WalletKey
}

func NewStore(wk *WalletKey) *WalletStore {
	return &WalletStore{
		wk: *wk,
	}
}

// Just a quick storage of the encrypted mnemonic for now.
// TODO: add metadata, decide what else should actually go in the "core wallet"

func (s *WalletStore) Open(file *os.File) (w *Wallet, err error) {
	jsonWallet, err := os.ReadFile(file.Name())
	if err != nil {
		return nil, err
	}
	ew := &ExportableWallet{}
	if err = json.Unmarshal(jsonWallet, ew); err != nil {
		return nil, err
	}
	s.wk.salt = base58.Decode(ew.Salt) // Replace auto-generated salt with the one from the wallet file.
	encWallet := base58.Decode(ew.EncryptedWallet)
	decMnemonic := s.wk.decrypt(encWallet, ew.Nonce)
	w = NewWalletFromMnemonic(string(decMnemonic))
	return w, nil
}

func (s *WalletStore) Export(file *os.File, w *Wallet) error {
	encWallet, nonce := s.wk.encrypt([]byte(w.Mnemonic()))
	ew := &ExportableWallet{
		Salt:            base58.Encode(s.wk.salt),
		EncryptedWallet: base58.Encode(encWallet),
		Nonce:           nonce,
	}
	jsonWallet, err := json.Marshal(ew)
	if err != nil {
		return err
	}
	if _, err := file.Write(jsonWallet); err != nil {
		return err
	}
	return nil
}
