package wallet

import (
	"crypto/ed25519"
	"fmt"
	"github.com/spacemeshos/smcli/common"
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

// NewWallet creates a new wallet containing one account generated using a random seed.
func NewWallet() (*Wallet, error) {
	return NewMultiWallet(1)
}

// NewMultiWallet creates a new wallet containing multiple accounts generated using random seeds.
func NewMultiWallet(n int) (*Wallet, error) {
	if n < 0 || n > common.MaxAccountsPerWallet {
		return nil, fmt.Errorf("invalid number of accounts")
	}
	var accounts []*EDKeyPair
	for i := 0; i < n; i++ {
		e, err := bip39.NewEntropy(ed25519.SeedSize * 8)
		if err != nil {
			return nil, err
		}
		a, err := accountFromSeed(e)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return walletFromAccounts(accounts)
}

// NewWalletFromSeed creates a new wallet containing one account generated using the given seed.
func NewWalletFromSeed(seed []byte) (*Wallet, error) {
	kp, err := accountFromSeed(seed)
	if err != nil {
		return nil, err
	}
	return walletFromAccounts([]*EDKeyPair{kp})
}

func walletFromAccounts(kp []*EDKeyPair) (*Wallet, error) {
	displayName := "Main Wallet"
	createTime := common.NowTimeString()

	w := &Wallet{
		Meta: walletMetadata{
			DisplayName: displayName,
			Created:     createTime,
			// TODO: set correctly
			GenesisID: "",
			Meta: MetaMetadata{
				Salt: common.DefaultEncryptionSalt,
			},
		},
		Secrets: walletSecrets{
			//Mnemonic: mnemonic,
			Accounts: kp,
		},
		//masterKeypair: masterKeyPair,
	}
	return w, nil
}

// accountFromSeed creates a new account from the given seed.
func accountFromSeed(seed []byte) (*EDKeyPair, error) {
	masterKeyPair, err := NewMasterKeyPair(seed[:ed25519.SeedSize])
	if err != nil {
		return nil, err
	}
	return masterKeyPair, nil
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
