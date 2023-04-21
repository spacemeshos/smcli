package wallet

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoreAndRetrieveWalletToFromFile(t *testing.T) {
	saltSlice, _ := hex.DecodeString("0102030405060708090a0b0c0d0e0f10")
	password, _ := hex.DecodeString("70617373776f7264")
	var salt, salt2 [Pbkdf2SaltBytesLen]byte
	copy(salt[:], saltSlice)

	wKey := NewKey(
		WithSalt(salt),
		WithPbkdf2Password(password),
	)

	w, err := NewMultiWalletRandomMnemonic(1)
	require.NoError(t, err)

	file, err := os.CreateTemp("./", "test_wallet.*.json")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file.Name())

	fmt.Println(file.Name())
	err = wKey.Export(file, w)
	require.NoError(t, err)

	file.Seek(0, io.SeekStart)
	w2, err := wKey.Open(file)
	require.NoError(t, err)
	require.Equal(t, w.Secrets.Accounts, w2.Secrets.Accounts)
	require.Equal(t, w.Secrets.Mnemonic, w2.Secrets.Mnemonic)

	// trying to open with a different wallet key, same pw and nonce, should work
	wKey = NewKey(
		WithSalt(salt),
		WithPbkdf2Password(password),
	)
	file.Seek(0, io.SeekStart)
	w2, err = wKey.Open(file)
	require.NoError(t, err)
	require.Equal(t, w.Secrets.Accounts, w2.Secrets.Accounts)
	require.Equal(t, w.Secrets.Mnemonic, w2.Secrets.Mnemonic)

	// trying to open the same file with a different key or nonce should fail
	password2 := password[:]
	password2[0]++

	// right salt, wrong password
	wKey = NewKey(
		WithSalt(salt),
		WithPbkdf2Password(password2),
	)
	file.Seek(0, io.SeekStart)
	_, err = wKey.Open(file)
	require.Error(t, err)

	// right password, wrong salt
	copy(salt2[:], saltSlice)
	salt2[0]++
	wKey = NewKey(
		WithSalt(salt2),
		WithPbkdf2Password(password),
	)
	file.Seek(0, io.SeekStart)
	_, err = wKey.Open(file)
	require.Error(t, err)

	// both wrong
	wKey = NewKey(
		WithSalt(salt2),
		WithPbkdf2Password(password2),
	)
	file.Seek(0, io.SeekStart)
	_, err = wKey.Open(file)
	require.Error(t, err)
}
