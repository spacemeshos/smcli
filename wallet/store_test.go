package wallet_test

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/spacemeshos/smcli/wallet"
	"github.com/stretchr/testify/require"
)

func TestStoreAndRetrieveWalletToFromFile(t *testing.T) {
	saltSlice, _ := hex.DecodeString("0102030405060708090a0b0c0d0e0f10")
	password, _ := hex.DecodeString("70617373776f7264")
	var salt, salt2 [wallet.Pbkdf2SaltBytesLen]byte
	copy(salt[:], saltSlice)

	wKey := wallet.NewKey(
		wallet.WithSalt(salt),
		wallet.WithPbkdf2Password(password),
	)

	//entropy, _ := bip39.NewEntropy(256)
	//mnemonic, _ := bip39.NewMnemonic(entropy)
	w := wallet.NewWallet()

	file, err := os.CreateTemp("./", "test_wallet.*.json")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	fmt.Println(file.Name())
	err = wKey.Export(file, w)
	require.NoError(t, err)

	w2, err := wKey.Open(file)
	require.NoError(t, err)
	require.Equal(t, w.Secrets.Accounts, w2.Secrets.Accounts)

	// trying to open with a different wallet key, same pw and nonce, should work
	wKey = wallet.NewKey(
		wallet.WithSalt(salt),
		wallet.WithPbkdf2Password(password),
	)
	w2, err = wKey.Open(file)
	require.NoError(t, err)
	require.Equal(t, w.Secrets.Accounts, w2.Secrets.Accounts)

	// trying to open the same file with a different key or nonce should fail
	password2 := password[:]
	password2[0]++

	// right salt, wrong password
	wKey = wallet.NewKey(
		wallet.WithSalt(salt),
		wallet.WithPbkdf2Password(password2),
	)
	_, err = wKey.Open(file)
	require.Error(t, err)

	// right password, wrong salt
	copy(salt2[:], saltSlice)
	salt2[0]++
	wKey = wallet.NewKey(
		wallet.WithSalt(salt2),
		wallet.WithPbkdf2Password(password),
	)
	_, err = wKey.Open(file)
	require.Error(t, err)

	// both wrong
	wKey = wallet.NewKey(
		wallet.WithSalt(salt2),
		wallet.WithPbkdf2Password(password2),
	)
	_, err = wKey.Open(file)
	require.Error(t, err)
}
