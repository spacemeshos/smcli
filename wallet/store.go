package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha512"
	"encoding/json"
	"os"

	"github.com/btcsuite/btcutil/base58"
	"github.com/spf13/cobra"
	"github.com/xdg-go/pbkdf2"
)

const EncKeyLen = 32

// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#pbkdf2
// TODO: should this be increased to 210,000 per the above link?
const Pbkdf2Iterations = 120000
const Pbkdf2Dklen = 256
const Pdkdf2SaltBytesLen = 16

var Pbkdf2HashFunc = sha512.New

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

// https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html#71-encryption-types-to-use
func (k *WalletKey) encrypt(plaintext []byte) (ciphertext []byte, nonce []byte) {
	block, err := aes.NewCipher(k.key)
	cobra.CheckErr(err)

	// Using default options for AES-GCM as recommended by the godoc.
	// For reference, NonceSize is 12 bytes, and TagSize is 16 bytes:
	// https://cs.opensource.google/go/go/+/refs/tags/go1.19.2:src/crypto/cipher/gcm.go;l=153-158
	aesgcm, err := cipher.NewGCM(block)
	nonce = make([]byte, aesgcm.NonceSize())
	hash := hmac.New(sha512.New, k.key)
	hash.Write(plaintext)
	nonce = hash.Sum(nil)[:aesgcm.NonceSize()]

	cobra.CheckErr(err)
	ciphertext = aesgcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce
}
func (k *WalletKey) decrypt(ciphertext []byte, nonce []byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(k.key)
	cobra.CheckErr(err)
	aesgcm, err := cipher.NewGCM(block)
	cobra.CheckErr(err)

	if plaintext, err = aesgcm.Open(nil, nonce, ciphertext, nil); err != nil {
		return nil, err
	}
	return plaintext, nil
}

//	type WalletOpener interface {
//		Open(path string) (*Wallet, error)
//	}
//
//	type WalletExporter interface {
//		Export(path string) error
//	}
//
//	type ExportableWallet struct {
//		// EncryptedWallet is the encrypted wallet data in base58 encoding.
//		EncryptedWallet string `json:"encrypted_wallet"`
//		// Salt is the salt used to derived the wallet's encryption key in base58 encoding.
//		Salt string `json:"salt"`
//		// Nonce is the nonce used to encrypt the wallet in base58 encoding.
//		Nonce []byte `json:"nonce"`
//	}
//
//	type WalletStore struct {
//		wk WalletKey
//	}
//
//	func NewStore(wk *WalletKey) *WalletStore {
//		return &WalletStore{
//			wk: *wk,
//		}
//	}
func Open(file *os.File) (w *Wallet, err error) {
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
	decMnemonic, err := s.wk.decrypt(encWallet, ew.Nonce)
	if err != nil {
		return nil, err
	}
	w = NewWalletFromMnemonic(string(decMnemonic))
	return w, nil
}

func (w *Wallet) Export(ws WalletKey, file *os.File) error {
	// encrypt the secrets
	plaintext, err := json.Marshal(w.Secrets)
	if err != nil {
		return err
	}
	ciphertext, nonce := ws.encrypt(plaintext)
	ew := &EncryptedWalletFile{
		Meta: w.Meta,
		Secrets: walletSecretsEncrypted{
			Cipher:     "AES-GCM",
			CipherText: base58.Encode(ciphertext),
			CipherParams: struct {
				IV string `json:"iv"`
				// use hex encoding? base64?
			}{IV: string(nonce)},
			KDF: "PBKDF2",
			KDFParams: struct {
				DKLen      int    `json:"dklen"`
				Hash       string `json:"hash"`
				Salt       string `json:"salt"`
				Iterations int    `json:"iterations"`
			}{
				DKLen:      Pbkdf2Dklen,
				Hash:       "SHA-256",
				Salt:       w.Meta.Meta.Salt,
				Iterations: Pbkdf2Iterations,
			},
		},
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
