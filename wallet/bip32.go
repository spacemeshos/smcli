package wallet

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/spacemeshos/smcli/common"
	"github.com/spacemeshos/smkeys/bip32"
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
	Path        string     `json:"path"`
	Public      PublicKey  `json:"publicKey"`
	Private     PrivateKey `json:"secretKey"`
	Salt        []byte     `json:"salt"`
}

func NewMasterKeyPair(seed []byte) (*EDKeyPair, error) {
	if len(seed) != bip32.SeedSize {
		return nil, ErrInvalidSeed
	}
	//salt := []byte(common.DefaultEncryptionSalt)
	// TODO: is salt needed here?
	key, err := bip32.FromSeed(seed)
	if err != nil {
		return nil, err
	}
	privKey := PrivateKey(key[:])

	return &EDKeyPair{
		DisplayName: "Main Wallet",
		Created:     common.NowTimeString(),
		Private:     privKey,
		Public:      PublicKey(ed25519.PrivateKey(privKey).Public().(ed25519.PublicKey)),
		//Salt:        salt,
	}, nil
}

func (kp *EDKeyPair) NewChildKeyPair(childIdx uint32) (*EDKeyPair, error) {
	// TODO: underlying fn does not support salt, do we need it?
	key, err := bip32.DeriveChild(
		ed25519.PrivateKey(kp.Private).Seed(),
		childIdx,
	)
	if err != nil {
		return nil, err
	}
	return &EDKeyPair{
		Private: key[:],
		Public:  PublicKey(ed25519.PrivateKey(key[:]).Public().(ed25519.PublicKey)),
		//Path:    append(kp.Path, childIdx),
		//Salt:    kp.Salt,
	}, nil
}
