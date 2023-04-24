package wallet

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/spacemeshos/smcli/common"
	smbip32 "github.com/spacemeshos/smkeys/bip32"
)

// Function names inspired by https://github.com/tyler-smith/go-bip32/blob/master/bip32.go
// We assume all keys are hardened.

var ErrInvalidSeed = errors.New("invalid seed length")

type PublicKey ed25519.PublicKey

func (k *PublicKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(*k))
}

func (k *PublicKey) UnmarshalJSON(data []byte) (err error) {
	var hexString string
	if err = json.Unmarshal(data, &hexString); err != nil {
		return
	}
	*k, err = hex.DecodeString(hexString)
	return
}

type PrivateKey ed25519.PrivateKey

func (k *PrivateKey) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(*k))
}

func (k *PrivateKey) UnmarshalJSON(data []byte) (err error) {
	var hexString string
	if err = json.Unmarshal(data, &hexString); err != nil {
		return
	}
	*k, err = hex.DecodeString(hexString)
	return
}

type EDKeyPair struct {
	DisplayName string     `json:"displayName"`
	Created     string     `json:"created"`
	Path        HDPath     `json:"path"`
	Public      PublicKey  `json:"publicKey"`
	Private     PrivateKey `json:"secretKey"`
}

func NewMasterKeyPair(seed []byte) (*EDKeyPair, error) {
	if len(seed) != ed25519.SeedSize {
		return nil, ErrInvalidSeed
	}
	key, err := smbip32.FromSeed(seed)
	if err != nil {
		return nil, err
	}
	privKey := PrivateKey(key[:])

	return &EDKeyPair{
		DisplayName: "Main Wallet",
		Created:     common.NowTimeString(),
		Private:     privKey,
		Public:      PublicKey(ed25519.PrivateKey(privKey).Public().(ed25519.PublicKey)),
		Path:        DefaultPath(),
		//Salt:        salt,
	}, nil
}

func (kp *EDKeyPair) NewChildKeyPair(childIdx int) (*EDKeyPair, error) {
	key, err := smbip32.DeriveChild(
		ed25519.PrivateKey(kp.Private).Seed(),
		uint32(childIdx),
	)
	if err != nil {
		return nil, err
	}
	return &EDKeyPair{
		Private: key[:],
		Public:  PublicKey(ed25519.PrivateKey(key[:]).Public().(ed25519.PublicKey)),
		Path:    kp.Path.Extend(BIP32HardenedKeyStart | uint32(childIdx)),
	}, nil
}
