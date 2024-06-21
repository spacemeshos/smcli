package wallet

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	walletTemplate "github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"github.com/tyler-smith/go-bip39"

	"github.com/spacemeshos/smcli/common"
)

var errWhitespace = errors.New("whitespace violation in mnemonic phrase")

// Wallet is the basic data structure.
type Wallet struct {
	// keystore string
	// password string
	// unlocked bool
	Meta    walletMetadata `json:"meta"`
	Secrets walletSecrets  `json:"crypto"`
}

// EncryptedWalletFile is the encrypted representation of the wallet on the filesystem.
type EncryptedWalletFile struct {
	Meta    walletMetadata         `json:"meta"`
	Secrets walletSecretsEncrypted `json:"crypto"`
}

type walletMetadata struct {
	DisplayName string `json:"displayName"`
	Created     string `json:"created"`
	GenesisID   string `json:"genesisID"`
	// NetID       int    `json:"netId"`

	// is this needed?
	// Type WalletType
	// RemoteAPI string
}

type hexEncodedCiphertext []byte

func (c *hexEncodedCiphertext) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(*c))
}

func (c *hexEncodedCiphertext) UnmarshalJSON(data []byte) (err error) {
	var hexString string
	if err = json.Unmarshal(data, &hexString); err != nil {
		return
	}
	*c, err = hex.DecodeString(hexString)
	return
}

type walletSecretsEncrypted struct {
	Cipher       string               `json:"cipher"`
	CipherText   hexEncodedCiphertext `json:"cipherText"`
	CipherParams struct {
		IV hexEncodedCiphertext `json:"iv"`
	} `json:"cipherParams"`
	KDF       string `json:"kdf"`
	KDFParams struct {
		DKLen      int                  `json:"dklen"`
		Hash       string               `json:"hash"`
		Salt       hexEncodedCiphertext `json:"salt"`
		Iterations int                  `json:"iterations"`
	} `json:"kdfparams"`
}

type walletSecrets struct {
	Mnemonic      string `json:"mnemonic"`
	MasterKeypair *EDKeyPair
	Accounts      []*EDKeyPair `json:"accounts"`
}

func NewMultiWalletRandomMnemonic(n int) (*Wallet, error) {
	// generate a new, random mnemonic
	e, err := bip39.NewEntropy(ed25519.SeedSize * 8)
	if err != nil {
		return nil, err
	}
	m, err := bip39.NewMnemonic(e)
	if err != nil {
		return nil, err
	}

	return NewMultiWalletFromMnemonic(m, n)
}

func NewMultiWalletFromMnemonic(m string, n int) (*Wallet, error) {
	if n < 0 || n > common.MaxAccountsPerWallet {
		return nil, errors.New("invalid number of accounts")
	}

	// bip39 lib doesn't properly validate whitespace so we have to do that manually.
	if expected := strings.Join(strings.Fields(m), " "); m != expected {
		return nil, errWhitespace
	}

	// this checks the number of words and the checksum.
	if !bip39.IsMnemonicValid(m) {
		return nil, errors.New("invalid mnemonic")
	}

	// TODO: add option for user to provide passphrase
	// https://github.com/spacemeshos/smcli/issues/18

	seed := bip39.NewSeed(m, "")
	masterKeyPair, err := NewMasterKeyPair(seed)
	if err != nil {
		return nil, err
	}
	accounts, err := accountsFromMaster(masterKeyPair, seed, n)
	if err != nil {
		return nil, err
	}
	return walletFromMnemonicAndAccounts(m, masterKeyPair, accounts)
}

func NewMultiWalletFromLedger(n int) (*Wallet, error) {
	if n < 0 || n > common.MaxAccountsPerWallet {
		return nil, errors.New("invalid number of accounts")
	}
	masterKeyPair, err := NewMasterKeyPairFromLedger()
	if err != nil {
		fmt.Println("Error: ", err)
		fmt.Println("Are you sure the ledger is connected, unlocked, and the Spacemesh app is open?")
		return nil, err
	}
	// seed is not used in case of ledger
	accounts, err := accountsFromMaster(masterKeyPair, []byte{}, n)
	if err != nil {
		return nil, err
	}
	return walletFromMnemonicAndAccounts("(none)", masterKeyPair, accounts)
}

func walletFromMnemonicAndAccounts(m string, masterKp *EDKeyPair, kp []*EDKeyPair) (*Wallet, error) {
	w := &Wallet{
		Meta: walletMetadata{
			DisplayName: "Main Wallet",
			Created:     common.NowTimeString(),
			// TODO: set correctly
			GenesisID: "",
		},
		Secrets: walletSecrets{
			Mnemonic:      m,
			MasterKeypair: masterKp,
			Accounts:      kp,
		},
	}
	return w, nil
}

// accountsFromMaster generates one or more accounts from a master keypair and seed. Accounts use sequential HD paths.
// The master keypair does not contain the seed that was used to generate it, so it needs to be passed in explicitly.
func accountsFromMaster(masterKeypair *EDKeyPair, masterSeed []byte, n int) (accounts []*EDKeyPair, err error) {
	accounts = make([]*EDKeyPair, 0, n)
	for i := 0; i < n; i++ {
		acct, err := masterKeypair.NewChildKeyPair(masterSeed, i)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, acct)
	}
	return
}

func (w *Wallet) Mnemonic() string {
	return w.Secrets.Mnemonic
}

func PubkeyToAddress(pubkey []byte, hrp string) string {
	types.SetNetworkHRP(hrp)
	key := [ed25519.PublicKeySize]byte{}
	copy(key[:], pubkey)
	walletArgs := &walletTemplate.SpawnArguments{PublicKey: key}
	walletAddress := core.ComputePrincipal(walletTemplate.TemplateAddress, walletArgs)
	return walletAddress.String()
}
