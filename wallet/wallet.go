package wallet

import (
	"crypto/ed25519"
	"github.com/spacemeshos/smcli/common"
	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
)

// Wallet is the basic data structure.
type Wallet struct {
	//keystore string
	//password string
	//unlocked bool
	Meta    walletMetadata `json:"meta"`
	Secrets walletSecrets  `json:"crypto"`

	// this is not persisted
	//masterKeypair *EDKeyPair
}

// EncryptedWalletFile is the encrypted representation of the wallet on the filesystem
type EncryptedWalletFile struct {
	Meta    walletMetadata         `json:"meta"`
	Secrets walletSecretsEncrypted `json:"crypto"`
}

type MetaMetadata struct {
	Salt string `json:"salt"`
}

type walletMetadata struct {
	DisplayName string       `json:"displayName"`
	Created     string       `json:"created"`
	GenesisID   string       `json:"genesisID"`
	Meta        MetaMetadata `json:"meta"`
	//NetID       int    `json:"netId"`

	// is this needed?
	//Type WalletType
	//RemoteAPI string
}

type walletSecretsEncrypted struct {
	Cipher       string `json:"cipher"`
	CipherText   string `json:"cipherText"`
	CipherParams struct {
		IV string `json:"iv"`
	} `json:"cipherParams"`
	KDF       string `json:"kdf"`
	KDFParams struct {
		DKLen      int    `json:"dklen"`
		Hash       string `json:"hash"`
		Salt       string `json:"salt"`
		Iterations int    `json:"iterations"`
	} `json:"kdfparams"`
}

type walletSecrets struct {
	//Mnemonic string       `json:"mnemonic"`
	Accounts []*EDKeyPair `json:"accounts"`

	//accountNumber int

	// supported in smapp, leave out for now for simplicity
	//Contacts      []contact `json:"contacts"`
}

//type contact struct {
//	Nickname string `json:"nickname"`
//	Address  string `json:"address"`
//}

// NewWallet creates a brand new wallet with a random seed.
func NewWallet() *Wallet {
	e, _ := bip39.NewEntropy(ed25519.SeedSize * 8)
	return NewWalletFromSeed(e)
}

// NewWalletFromSeed creates a new wallet from the given seed.
func NewWalletFromSeed(seed []byte) *Wallet {
	// Arbitrarily taking the first 32 bytes as the seed for the private key.
	// Not sure if this is the correct approach.
	masterKeyPair, err := NewMasterKeyPair(seed[:ed25519.SeedSize])
	cobra.CheckErr(err)

	displayName := "Main Wallet"
	createTime := common.NowTimeString()

	w := &Wallet{
		Meta: walletMetadata{
			DisplayName: displayName,
			Created:     createTime,
			// TODO: set correctly
			GenesisID: "",
			Meta: MetaMetadata{
				Salt: string(masterKeyPair.Salt),
			},
		},
		Secrets: walletSecrets{
			//Mnemonic: mnemonic,
			Accounts: []*EDKeyPair{
				masterKeyPair,
			},
		},
		//masterKeypair: masterKeyPair,
	}

	// Add the first key pair.
	// We only use the master key pair as a seed to generate child addresses.
	// We don't store the master key pair as an address.
	// Go ahead and derive the first child address.
	// To do this, we need to construct the appropriate path first.
	//path := append(NewPath(), BIP44Account(0))
	//keyPair, err := w.ComputeKeyPair(path)
	//cobra.CheckErr(err)
	//w.Secrets.Accounts = []*EDKeyPair{keyPair}

	return w
}
func (w *Wallet) Salt() []byte {
	return []byte(w.Meta.Meta.Salt)
}

//func (w *Wallet) Mnemonic() string {
//	return w.Secrets.Mnemonic
//}

// ComputeKeyPair returns the key pair for the given HDPath.
// It will compute it every time from the master key.
// If the path is empty, it will return the master key pair.
//func (w *Wallet) ComputeKeyPair(path HDPath) (*EDKeyPair, error) {
//	if !IsPathCompletelyHardened(path) {
//		return nil, fmt.Errorf("unhardened keys aren't supported")
//	}
//
//	keypair := w.masterKeypair
//
//	for i, childKeyIndex := range path {
//		if i == HDPurposeSegment && childKeyIndex != BIP44Purpose() {
//			return nil, fmt.Errorf("invalid purpose: expected %d, got %d", BIP44Purpose(), childKeyIndex)
//		}
//		if i == HDCoinTypeSegment && childKeyIndex != BIP44SpacemeshCoinType() {
//			return nil, fmt.Errorf("invalid coin type: expected %d, got %d", BIP44SpacemeshCoinType(), childKeyIndex)
//		}
//		kp, err := keypair.NewChildKeyPair(childKeyIndex)
//		if err != nil {
//			return nil, err
//		}
//		keypair = kp
//	}
//	return keypair, nil
//}
