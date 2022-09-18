package wallet

import (
	"github.com/spacemeshos/address"
)

// Account is a single account in a wallet.
type Account struct {
	Name    string
	keyPair BIP32EDKeyPair
	chain   []*Chain
}

// NewAddress creates a new key pair on the given chain in this account and
// returns it's address.
func (a *Account) NewAddress(chain uint32) address.Address {
	// return address.GenerateAddress(a.KeyPair.Public)
	panic("not implemented")
}

func (a *Account) KeyPair(addr address.Address) BIP32EDKeyPair {
	panic("not implemented")
}

func (a *Account) Path() HDPath {
	return a.keyPair.Path
}
func (a *Account) ToBytes() []byte {
	panic("not implemented")
}

func AccountFromBytes(b []byte) *Account {
	panic("not implemented")
}

type Chain struct {
	keyPair BIP32EDKeyPair
	// These are the indices of the child keys created from this chain.
	index []*BIP32EDKeyPair
}
