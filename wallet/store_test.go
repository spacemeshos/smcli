package wallet_test

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/spacemeshos/smcli/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/tyler-smith/go-bip39"
)

func TestStoreAndRetrieveWalletToFromFile(t *testing.T) {
	wKey := wallet.NewKey(
		wallet.WithArgon2idPassword("password"),
	)
	wStore := wallet.NewStore(wKey)

	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.WalletFromMnemonic(mnemonic)

	file, err := ioutil.TempFile("./", "test_wallet.*.json")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	fmt.Println(file.Name())
	err = wStore.Export(file, w)
	assert.NoError(t, err)

	wStore2 := wallet.NewStore(wKey)
	w2, err := wStore2.Open(file)
	assert.NoError(t, err)
	assert.Equal(t, w.Mnemonic(), w2.Mnemonic())
}

func TestWalletPbkdf2KeyEncryptDeterminism(t *testing.T) {
	// Test vectors
	plaintext, _ := hex.DecodeString("737570657220736563726574207365637265742074686174206e6565647320746f20626520736563726574")
	saltSlice, _ := hex.DecodeString("0102030405060708090a0b0c0d0e0f10")
	password, _ := hex.DecodeString("70617373776f7264")
	var salt [wallet.Pdkdf2SaltBytesLen]byte
	copy(salt[:], saltSlice)
	expectedCiphertext, _ := hex.DecodeString("30159eecbccaca65d85622be889e2b5c4887e6e65a3ba3a887f23565b67dc29b05cb1e49698d0d7e599174573fff936b894155e5ea35f67f2ce746")
	expectedNonce, _ := hex.DecodeString("13aabfda6f09c9a005d862d0")

	// Lets do it
	wKey := wallet.NewKey(
		wallet.WithPbkdf2Password(
			password,
			salt,
		),
	)
	ciphertext, nonce := wKey.Encrypt(plaintext)
	assert.Equal(t, expectedCiphertext, ciphertext)
	assert.Equal(t, expectedNonce, nonce)

	// Decrypt for better mental health stability
	decrypted, _ := wKey.Decrypt(ciphertext, nonce)
	assert.Equal(t, plaintext, decrypted)
}
