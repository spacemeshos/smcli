package main

import (
	"encoding/json"
	"fmt"
	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"log"
	"math"
	"math/rand"

	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/config"
	"github.com/spacemeshos/go-spacemesh/config/presets"
	genvm "github.com/spacemeshos/go-spacemesh/genvm"
	"github.com/spacemeshos/go-spacemesh/genvm/sdk"
	// sdkMultisig "github.com/spacemeshos/go-spacemesh/genvm/sdk/multisig"
	// sdkVesting "github.com/spacemeshos/go-spacemesh/genvm/sdk/vesting"
	sdkWallet "github.com/spacemeshos/go-spacemesh/genvm/sdk/wallet"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"github.com/spacemeshos/go-spacemesh/signing"
	"github.com/spacemeshos/go-spacemesh/sql"
)

// type RegularOutput struct {
// 	// destinationaddress string `json:"destination_address"`
// 	// amount             string `json:"amount"`
// 	// nonce              int    `json:"nonce"`
// }

// type ExpertOutput struct {
// 	// TransactionID string `json:"transaction_id"`
// 	// Status        string `json:"status"`
// }

type typeAccount int

const (
	Wallet typeAccount = iota
	Multisig
	Vault
	Vesting
)

type typeTx int

const (
	Spawn typeTx = iota
	SelfSpawn
	Spend
)

type Output struct {
	destination string
	amount      uint64
	fee         uint64
	gasMax      uint64
	gasPrice    uint64
	maxSpend    uint64
	nonce       uint64
	principal   string
	typeAccount typeAccount
	typeTx      typeTx
	// Regular []RegularOutput `json:"output"`
	// Expert  []ExpertOutput  `json:"output_expert"`
}

type TestVector struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Blob  string `json:"blob"`
	Output
}

func generateTestVectors() []TestVector {
	// we only need one keypair
	pub, priv := getKey()
	pubCore := types.BytesToHash(pub.Bytes())

	// read network configs - needed for genesisID
	var configMainnet, configTestnet config.GenesisConfig
	configMainnet = config.MainnetConfig().Genesis
	if testnet, err := presets.Get("testnet"); err != nil {
		log.Fatalf("Error getting testnet config: %v", err)
	} else {
		configTestnet = testnet.Genesis
	}
	// not sure how to get hrp programmatically from config so we just hardcode it
	networks := map[string]config.GenesisConfig{
		"sm":    configMainnet,
		"stest": configTestnet,
	}

	testVectors := make([]TestVector, 0)
	// accountTypes := []string{"wallet", "multisig", "vault", "vesting"}
	// transactionTypes := map[string][]string{
	// 	"wallet":   {"spawn", "spend"},
	// 	"multisig": {"spawn", "addSigner"},
	// 	"vault":    {"withdraw"},
	// 	"vesting":  {"claim"},
	// }

	index := 0
	nonce := uint64(0)
	for hrp, netconf := range networks {
		// initialization
		var tx []byte
		genesisID := netconf.GenesisID()
		opts := []sdk.Opt{
			sdk.WithGenesisID(genesisID),
			sdk.WithGasPrice(1),
		}

		// we need a VM object for validation and gas cost computation
		vm := genvm.New(sql.InMemory(), genvm.WithConfig(genvm.Config{GasLimit: math.MaxUint64, GenesisID: genesisID}))

		// simple wallet

		// spawn
		tx = sdkWallet.Spawn(priv, wallet.TemplateAddress, &wallet.SpawnArguments{PublicKey: pubCore}, nonce, opts...)
		validator := vm.Validation(types.NewRawTx(tx))
		header, err := validator.Parse()
		if !validator.Verify() {
			log.Fatalf("Error validating transaction")
		}
		if err != nil {
			log.Fatalf("Error parsing transaction: %v", err)
		}
		testVector := TestVector{
			Index: index,
			Name:  fmt.Sprintf("%s_%s_%s", hrp, "wallet", "spawn"),
			Blob:  types.BytesToHash(tx).String(),
			Output: Output{
				gasMax:      header.MaxGas,
				gasPrice:    header.GasPrice,
				maxSpend:    header.MaxSpend,
				nonce:       nonce,
				principal:   header.Principal.String(),
				typeAccount: Wallet,
				typeTx:      Spawn,
			},
		}
		index++
		testVectors = append(testVectors, testVector)

		// self spawn
		tx = sdkWallet.SelfSpawn(priv, 1)
		// spend
	}
	return testVectors
}

func main() {
	testVectors := generateTestVectors()

	jsonData, err := json.MarshalIndent(testVectors, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling test vectors: %v", err)
	}

	fmt.Println(string(jsonData))
}

func getKey() (pub signing.PublicKey, priv signing.PrivateKey) {
	// generate a random pubkey and discard the private key
	edPub, edPriv, err := ed25519.GenerateKey(rand.New(rand.NewSource(rand.Int63())))
	if err != nil {
		log.Fatal("failed to generate ed25519 key")
	}
	// pub = *signing.NewPublicKey(edPub)
	pub = signing.PublicKey{edPub}
	priv = signing.PrivateKey(edPriv)
	return
}
