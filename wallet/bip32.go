package wallet

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"

	smbip32 "github.com/spacemeshos/smkeys/bip32"
	ledger "github.com/spacemeshos/smkeys/remote-wallet"

	"github.com/spacemeshos/smcli/common"
)

// Function names inspired by https://github.com/tyler-smith/go-bip32/blob/master/bip32.go
// We assume all keys are hardened.

type PublicKey ed25519.PublicKey

type keyType int

const (
	typeSoftware keyType = iota
	typeLedger
)

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
	KeyType     keyType    `json:"keyType"`
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
	if kp.KeyType == typeLedger {
		return pubkeyFromLedger(path, false)
	} else if kp.KeyType == typeSoftware {
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
	} else {
		return nil, fmt.Errorf("unknown key type")
	}
}

func NewMasterKeyPairFromLedger() (*EDKeyPair, error) {
	return pubkeyFromLedger(DefaultPath(), true)
}

func pubkeyFromLedger(path HDPath, master bool) (*EDKeyPair, error) {
	// TODO: support multiple ledger devices (https://github.com/spacemeshos/smcli/issues/46)
	// don't bother confirming the master key; we only want the user to have to confirm a single key,
	// the one they really care about, which is the first child key.
	key, err := ledger.ReadPubkeyFromLedger("", HDPathToString(path), !master)
	if err != nil {
		return nil, fmt.Errorf("error reading pubkey from ledger. Are you sure it's connected, unlocked, and the Spacemesh app is open? err: %w", err)
	}

	name := "Ledger Master Key"
	if !master {
		name = "Ledger Child Key"
	}

	return &EDKeyPair{
		DisplayName: name,
		Created:     common.NowTimeString(),
		// note: we do not set a Private key here (it lives on the device)
		Public:  key[:],
		Path:    path,
		KeyType: typeLedger,
	}, nil
}
