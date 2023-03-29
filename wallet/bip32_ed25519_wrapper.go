package wallet

import (
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	ed25519sm "github.com/spacemeshos/ed25519-recovery"
	"github.com/spacemeshos/smcli/common"
	"log"
)

// BIP32HardenedKeyStart: keys with index >= this must be hardened as per BIP32.
// https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki#extended-keys
const BIP32HardenedKeyStart uint32 = 0x80000000

// Function names inspired by https://github.com/tyler-smith/go-bip32/blob/master/bip32.go
// This is a POC implementation that relies on the key derivation from the
// spacemeshos/ed25519 package.
// We assume all keys are hardened.
// TODO: In the future, we should use the child key derivation function from the BIP32-ED25519 paper
// https://github.com/LedgerHQ/orakolo/blob/master/papers/Ed25519_BIP%20Final.pdf
// so that we can be compatible with other wallets and so that spacemesh can be stored
// in a wallet containing other coins using the same derivation scheme.

var ErrInvalidSeed = errors.New("invalid seed length")
var ErrNotHardened = errors.New("child index must be hardened")

const (
	HDPurposeSegment  = 0
	HDCoinTypeSegment = 1
	HDAccountSegment  = 2
	HDChainSegment    = 3
	HDIndexSegment    = 4
)

type HDPath []uint32

// NewPath constructs and returns a new, default "empty" path, ready to add
// new accounts to
// TODO: support multiple chain IDs
func NewPath() HDPath {
	hdp, err := StringToHDPath("m/44'/540'/0'/0'")
	if err != nil {
		log.Fatalln(err)
	}
	return hdp
}

func (p HDPath) MarshalJSON() ([]byte, error) {
	s := p.String()
	fmt.Println("Marshaled:", s)
	return json.Marshal(p.String())
}

func (p HDPath) String() string {
	return HDPathToString(p)
}

func (p HDPath) Purpose() uint32 {
	return p[HDPurposeSegment]
}
func (p HDPath) CoinType() uint32 {
	return p[HDCoinTypeSegment]
}
func (p HDPath) Account() uint32 {
	return p[HDAccountSegment]
}
func (p HDPath) Chain() uint32 {
	return p[HDChainSegment]
}
func (p HDPath) Index() uint32 {
	return p[HDIndexSegment]
}

type BIP32EDKeyPair struct {
	DisplayName string `json:"displayName"`
	Created     string `json:"created"`
	// we assume the prefix is m/ and all keys are hardened
	Path    HDPath             `json:"path"`
	Public  ed25519.PublicKey  `json:"publicKey"`
	Private ed25519.PrivateKey `json:"secretKey"`
	Salt    []byte
}

func NewMasterBIP32EDKeyPair(seed []byte) (*BIP32EDKeyPair, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, ErrInvalidSeed
	}
	salt := []byte(common.DefaultEncryptionSalt)
	privKey := ed25519.NewKeyFromSeed(seed)

	return &BIP32EDKeyPair{
		Private: privKey,
		Public:  privKey.Public().(ed25519.PublicKey),
		Path:    NewPath(),
		Salt:    salt,
	}, nil
}

func (kp *BIP32EDKeyPair) NewChildKeyPair(childIdx uint32) (*BIP32EDKeyPair, error) {
	if childIdx < BIP32HardenedKeyStart {
		return nil, ErrNotHardened
	}

	// We still need to use the extended library for this, since the core library doesn't
	// support this operation natively. Ideally we'd use a reliable, standard BIP32 derivation
	// library for this, but I couldn't find one.
	privKey := ed25519sm.NewDerivedKeyFromSeed(
		kp.Private.Seed(),
		uint64(childIdx),
		kp.Salt,
	)
	return &BIP32EDKeyPair{
		Private: privKey,
		Public:  privKey.Public().(ed25519.PublicKey),
		Path:    append(kp.Path, childIdx),
		Salt:    kp.Salt,
	}, nil
}
