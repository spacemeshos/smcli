package wallet

import (
	"math"
	"sync"

	"github.com/spacemeshos/address"
	"github.com/spf13/cobra"
)

// Account is a single account in a wallet.
type Account struct {
	Name    string
	keyPair BIP32EDKeyPair
	chains  map[uint32]*chain
	lock    *sync.Mutex
}

// NextAddress generates and appends the next key pair onto the given chain in
// this account and returns it's address. If the chain does not exist, it is
// created.
func (a *Account) NextAddress(chainNum uint32) address.Address {
	a.lock.Lock()
	defer a.lock.Unlock()
	var kp *BIP32EDKeyPair
	if _, exists := a.chains[chainNum]; !exists {
		chainKeyPair, err := a.keyPair.NewChildKeyPair(BIP44HardenedChain(chainNum))
		cobra.CheckErr(err)
		a.chains[chainNum] = &chain{
			keyPair: chainKeyPair,
			index:   make(map[uint32]*BIP32EDKeyPair, 0),
		}
	}
	kp = a.chains[chainNum].nextHardenedKeyPair()
	return address.GenerateAddress(kp.Public)
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

type chain struct {
	keyPair BIP32EDKeyPair
	// These are the indices of the child keys created from this chain.
	index       map[uint32]*BIP32EDKeyPair
	numHardened uint32
}

func (c *chain) nextHardenedKeyPair() *BIP32EDKeyPair {
	if c.numHardened == math.MaxUint32 {
		panic("too many hardened keys in one chain")
	}
	index := c.numHardened
	keyPair, err := c.keyPair.NewChildKeyPair(BIP44HardenedAccountIndex(index))
	cobra.CheckErr(err)
	c.index[index] = &keyPair
	c.numHardened = index + 1
	return &keyPair
}
