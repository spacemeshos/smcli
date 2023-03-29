package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"github.com/spacemeshos/smcli/common"
	"os"

	"github.com/btcsuite/btcutil/base58"
	"github.com/spf13/cobra"
	"github.com/xdg-go/pbkdf2"
)

const EncKeyLen = 32

// https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#pbkdf2
const Pbkdf2Iterations = 210000
const Pbkdf2Dklen = 256

//const Pdkdf2SaltBytesLen = 16

var Pbkdf2HashFunc = sha512.New

type WalletKeyOpt func(*WalletKey)
type WalletKey struct {
	key []byte
	//salt []byte
}

func NewKey(opts ...WalletKeyOpt) WalletKey {
	w := &WalletKey{}
	for _, opt := range opts {
		opt(w)
	}
	if w.key == nil {
		panic("Some form of key generation method must be provided. Try WithXXXPassword.")
	}

	return *w
}

func WithPbkdf2Password(password string) WalletKeyOpt {
	return func(k *WalletKey) {
		k.key = pbkdf2.Key(
			[]byte(password),
			[]byte(common.DefaultEncryptionSalt),
			Pbkdf2Iterations,
			EncKeyLen,
			Pbkdf2HashFunc,
		)
	}
}

// https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html#71-encryption-types-to-use
func (k WalletKey) encrypt(plaintext []byte) (ciphertext []byte, nonce []byte) {
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
func (k WalletKey) decrypt(ciphertext []byte, nonce []byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(k.key)
	cobra.CheckErr(err)
	aesgcm, err := cipher.NewGCM(block)
	cobra.CheckErr(err)

	if plaintext, err = aesgcm.Open(nil, nonce, ciphertext, nil); err != nil {
		return nil, err
	}
	return plaintext, nil
}

func (k WalletKey) Open(file *os.File) (w *Wallet, err error) {
	jsonWallet, err := os.ReadFile(file.Name())
	if err != nil {
		return nil, err
	}
	ew := &EncryptedWalletFile{}
	if err = json.Unmarshal(jsonWallet, ew); err != nil {
		return nil, err
	}

	// make sure the salt matches
	//salt := ew.Meta.Meta.Salt
	//if salt != k.salt {
	//	return nil, fmt.Errorf("wallet file salt mismatch")
	//}

	nonce := base58.Decode(ew.Secrets.CipherParams.IV)
	encWallet := base58.Decode(ew.Secrets.CipherText)

	// TODO: before decrypting, check that other meta params match
	plaintext, err := k.decrypt(encWallet, nonce)
	if err != nil {
		return nil, err
	}
	secrets := &walletSecrets{}
	if err = json.Unmarshal(plaintext, secrets); err != nil {
		return nil, err
	}

	// we have everything we need, construct and return the wallet. first, we construct
	// a new wallet from the mnemonic. then we restore the metadata.
	w = NewWalletFromMnemonic(secrets.Mnemonic)
	w.Meta = ew.Meta
	return
}

func (k WalletKey) Export(file *os.File, w *Wallet) error {
	// encrypt the secrets
	plaintext, err := json.Marshal(w.Secrets)
	if err != nil {
		return err
	}
	ciphertext, nonce := k.encrypt(plaintext)
	ew := &EncryptedWalletFile{
		Meta: w.Meta,
		Secrets: walletSecretsEncrypted{
			Cipher:     "AES-GCM",
			CipherText: base58.Encode(ciphertext),
			CipherParams: struct {
				IV string `json:"iv"`
				// use hex encoding? base64?
			}{
				IV: base58.Encode(nonce),
			},
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
