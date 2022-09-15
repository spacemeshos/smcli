package wallet

import (
	"bytes"
	"encoding/gob"

	"github.com/spacemeshos/ed25519"
	gosmtypes "github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

// Account is a single account in a wallet.
type Account struct {
	Name    string
	PrivKey ed25519.PrivateKey // the pub & private key
	PubKey  ed25519.PublicKey  // only the pub key part
	Salt    []byte             // salt used to derive the private key
}

func (a *Account) Address() gosmtypes.Address {
	return gosmtypes.BytesToAddress(a.PubKey[:])
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

func GenerateBasicAccount(name string) *Account {
	entropy, _ := bip39.NewEntropy(256)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	seed := bip39.NewSeed(mnemonic, "")
	key, err := bip32.NewMasterKey(seed)
	if err != nil {
		panic(err)
	}
	key.NewChildKey(0)
}
