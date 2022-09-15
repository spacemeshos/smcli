package wallet

import (
	"bytes"
	"encoding/gob"

	"github.com/spacemeshos/address"
)

// Account is a single account in a wallet.
type Account struct {
	Name    string
	KeyPair BIP32EDKeyPair
}

func (a *Account) Address() address.Address {
	return address.GenerateAddress(a.KeyPair.Public)
}
func (a *Account) ToBytes() []byte {
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(a)
	return buf.Bytes()
}

func AccountFromBytes(b []byte) *Account {
	acct := &Account{}
	gob.NewDecoder(bytes.NewReader(b)).Decode(acct)
	// TODO: try raw sign/verify to check the keypair
	return acct
}
