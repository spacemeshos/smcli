package wallet

import (
	"github.com/spacemeshos/address"
)

// Account is a single account in a wallet.
type Account struct {
	Name    string
	KeyPair BIP32EDKeyPair
	Chain   []*Chain
}

// NextAddress
func (a *Account) LastAddress() address.Address {
	return address.GenerateAddress(a.KeyPair.Public)
}
func (a *Account) ToBytes() []byte {
	panic("not implemented")
}

func AccountFromBytes(b []byte) *Account {
	panic("not implemented")
}

type Chain struct {
	KeyPair BIP32EDKeyPair
	// These are the indices of the child keys created from this chain.
	Index []*BIP32EDKeyPair
}
