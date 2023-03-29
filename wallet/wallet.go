package wallet

import (
	"fmt"
	"github.com/spacemeshos/smcli/common"
	"github.com/tyler-smith/go-bip39"

	ed25519sm "github.com/spacemeshos/ed25519-recovery"
	"github.com/spf13/cobra"
)

// Wallet is the basic data structure.
type Wallet struct {
	//keystore string
	//password string
	//unlocked bool
	Meta    walletMetadata `json:"meta"`
	Secrets walletSecrets  `json:"crypto"`
	//Encrypted walletSecretsEncrypted `json:"crypto"`

	// this is not persisted
	masterKeypair *BIP32EDKeyPair
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
	Mnemonic string            `json:"mnemonic"`
	Accounts []*BIP32EDKeyPair `json:"accounts"`

	//accountNumber int

	// supported in smapp, leave out for now for simplicity
	//Contacts      []contact `json:"contacts"`
}

//type contact struct {
//	Nickname string `json:"nickname"`
//	Address  string `json:"address"`
//}

// NewWallet creates a brand new wallet with a random mnemonic.
func NewWallet() *Wallet {
	e, _ := bip39.NewEntropy(256)
	m, _ := bip39.NewMnemonic(e)
	return NewWalletFromMnemonic(m)
}

// NewWalletFromMnemonic creates a new wallet from the given mnemonic.
// The mnemonic must be a valid bip39 mnemonic.
func NewWalletFromMnemonic(mnemonic string) *Wallet {
	if !bip39.IsMnemonicValid(mnemonic) {
		panic("invalid mnemonic")
	}
	// TODO: add option for user to provide passphrase
	// https://github.com/spacemeshos/smcli/issues/18
	seed := bip39.NewSeed(mnemonic, "")

	// Arbitrarily taking the first 32 bytes as the seed for the private key
	// because spacemeshos/ed25519 gets angry if it gets all 64 bytes.
	// Not sure if this is the correct approach.
	masterKeyPair, err := NewMasterBIP32EDKeyPair(seed[ed25519sm.SeedSize:])
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
			Mnemonic: mnemonic,
			//Accounts: []*BIP32EDKeyPair{
			//	keyPair,
			//},
		},
		masterKeypair: masterKeyPair,
	}

	// Add the first key pair.
	// We only use the master key pair as a seed to generate child addresses.
	// We don't store the master key pair as an address.
	// Go ahead and derive the first child address.
	// To do this, we need to construct the appropriate path first.
	path := append(NewPath(), BIP44Account(0))
	keyPair, err := w.ComputeKeyPair(path)
	cobra.CheckErr(err)
	w.Secrets.Accounts = []*BIP32EDKeyPair{keyPair}

	return w
}
func (w *Wallet) Salt() []byte {
	return []byte(w.Meta.Meta.Salt)
}

func (w *Wallet) Mnemonic() string {
	return w.Secrets.Mnemonic
}

// ComputeKeyPair returns the key pair for the given HDPath.
// It will compute it every time from the master key.
// If the path is empty, it will return the master key pair.
func (w *Wallet) ComputeKeyPair(path HDPath) (*BIP32EDKeyPair, error) {
	if !IsPathCompletelyHardened(path) {
		return nil, fmt.Errorf("unhardened keys aren't supported")
	}

	keypair := w.masterKeypair

	for i, childKeyIndex := range path {
		if i == HDPurposeSegment && childKeyIndex != BIP44Purpose() {
			return nil, fmt.Errorf("invalid purpose: expected %d, got %d", BIP44Purpose(), childKeyIndex)
		}
		if i == HDCoinTypeSegment && childKeyIndex != BIP44SpacemeshCoinType() {
			return nil, fmt.Errorf("invalid coin type: expected %d, got %d", BIP44SpacemeshCoinType(), childKeyIndex)
		}
		kp, err := keypair.NewChildKeyPair(childKeyIndex)
		if err != nil {
			return nil, err
		}
		keypair = kp
	}
	return keypair, nil
}
