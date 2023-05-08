package wallet

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"

	smbip32 "github.com/spacemeshos/smkeys/bip32"

	"github.com/spacemeshos/smcli/common"
)

// Function names inspired by https://github.com/tyler-smith/go-bip32/blob/master/bip32.go
// We assume all keys are hardened.

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
	path := DefaultPath()
	key, err := smbip32.Derive(HDPathToString(path), seed)
	if err != nil {
		return nil, err
	}

	return &EDKeyPair{
		DisplayName: "Master Key",
		Created:     common.NowTimeString(),
		Private:     key[:],
		Public:      PublicKey(ed25519.PrivateKey(key[:]).Public().(ed25519.PublicKey)),
		Path:        path,
	}, nil
}

func (kp *EDKeyPair) NewChildKeyPair(seed []byte, childIdx int) (*EDKeyPair, error) {
	path := kp.Path.Extend(BIP44HardenedAccountIndex(uint32(childIdx)))
	key, err := smbip32.Derive(HDPathToString(path), seed)
	if err != nil {
		return nil, err
	}
	return &EDKeyPair{
		DisplayName: fmt.Sprintf("Child Key %d", childIdx),
		Created:     common.NowTimeString(),
		Private:     key[:],
		Public:      PublicKey(ed25519.PrivateKey(key[:]).Public().(ed25519.PublicKey)),
		Path:        path,
	}, nil
}
