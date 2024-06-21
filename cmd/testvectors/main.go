package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/sirupsen/logrus"
	"github.com/spacemeshos/economics/constants"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/config"
	"github.com/spacemeshos/go-spacemesh/config/presets"
	genvm "github.com/spacemeshos/go-spacemesh/genvm"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/genvm/sdk"

	// sdkMultisig "github.com/spacemeshos/go-spacemesh/genvm/sdk/multisig"
	// sdkVesting "github.com/spacemeshos/go-spacemesh/genvm/sdk/vesting"
	sdkWallet "github.com/spacemeshos/go-spacemesh/genvm/sdk/wallet"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"github.com/spacemeshos/go-spacemesh/signing"
	"github.com/spacemeshos/go-spacemesh/sql"
)

type typeAccount string

const (
	Wallet   typeAccount = "wallet"
	Multisig             = "multisig"
	Vault                = "vault"
	Vesting              = "vesting"
)

type typeTx string

const (
	Spawn     typeTx = "spawn"
	SelfSpawn        = "self_spawn"
	Spend            = "spend"
)

type Output struct {
	destination string
	amount      uint64
	gasMax      uint64
	gasPrice    uint64
	maxSpend    uint64
	nonce       uint64
	principal   string
	typeAccount typeAccount
	typeTx      typeTx
}

type TestVector struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Blob  string `json:"blob"`
	Output
}

func init() {
	// Set log level based on an environment variable
	if os.Getenv("DEBUG_MODE") == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		// logrus.SetLevel(logrus.InfoLevel)
		logrus.SetLevel(logrus.DebugLevel)
	}
}

// generate a random address for testing
func generateAddress() types.Address {
	pub, _ := getKey()
	return types.GenerateAddress(pub.Bytes())
}

func txToTestVector(tx []byte, vm *genvm.VM, index int, amount, nonce uint64, accountType typeAccount, txType typeTx, destination, hrp string) TestVector {
	validator := vm.Validation(types.NewRawTx(tx))
	header, err := validator.Parse()

	// apply the parsed self spawn tx
	// this will allow the spend tx to be validated
	if txType == SelfSpawn {
		logrus.Debugf("Applying self-spawn tx idx %d for account type %s", index, accountType)
		coreTx := types.Transaction{
			TxHeader: header,
			RawTx:    types.NewRawTx(tx),
		}
		skipped, results, err := vm.Apply(genvm.ApplyContext{Layer: types.FirstEffectiveGenesis()}, []types.Transaction{coreTx}, []types.CoinbaseReward{})
		if len(skipped) != 0 {
			logrus.Fatalf("Error applying self spawn tx idx %d", index)
		} else if len(results) != 1 || results[0].Status != types.TransactionSuccess {
			logrus.Fatalf("Error applying self spawn tx idx %d: %v", index, results[0].Status)
		} else if err != nil {
			logrus.Fatalf("Error applying self spawn tx idx %d: %v", index, err)
		}
		logrus.Debugf("got result: %v", results[0].TransactionResult)
	}

	if err != nil {
		logrus.Fatalf("Error parsing transaction idx %d: %v", index, err)
	}
	if !validator.Verify() {
		logrus.Fatalf("Error validating transaction idx %d", index)
	}
	return TestVector{
		Index: index,
		Name:  fmt.Sprintf("%s_%s_%s", hrp, "wallet", txType),
		Blob:  types.BytesToHash(tx).String(),
		Output: Output{
			// note: not all fields used in all tx types.
			// will be decoded in output.
			destination: destination,
			amount:      amount,
			gasMax:      header.MaxGas,
			gasPrice:    header.GasPrice,
			maxSpend:    header.MaxSpend,
			nonce:       nonce,
			principal:   header.Principal.String(),
			typeAccount: Wallet,
			typeTx:      Spawn,
		},
	}
}

func generateTestVectors() []TestVector {
	// we only need one keypair
	pub, priv := getKey()
	pubCore := types.BytesToHash(pub.Bytes())

	// read network configs - needed for genesisID
	var configMainnet, configTestnet config.GenesisConfig
	configMainnet = config.MainnetConfig().Genesis
	if testnet, err := presets.Get("testnet"); err != nil {
		logrus.Fatalf("Error getting testnet config: %v", err)
	} else {
		configTestnet = testnet.Genesis
	}
	// not sure how to get hrp programmatically from config so we just hardcode it
	networks := map[string]config.GenesisConfig{
		"sm":    configMainnet,
		"stest": configTestnet,
	}

	testVectors := make([]TestVector, 0)
	index := 0
	nonce := uint64(0)
	amount := uint64(constants.OneSmesh)
	// just use a single, random destination address
	// note: destination is not used in all tx types
	destination := generateAddress()
	for hrp, netconf := range networks {
		// hrp is used in address generation
		types.SetNetworkHRP(hrp)

		// initialization
		genesisID := netconf.GenesisID()
		opts := []sdk.Opt{
			sdk.WithGenesisID(genesisID),
			sdk.WithGasPrice(1),
		}
		principal := core.ComputePrincipal(wallet.TemplateAddress, &wallet.SpawnArguments{PublicKey: pubCore})

		// we need a VM object for validation and gas cost computation
		vm := genvm.New(sql.InMemory(), genvm.WithConfig(genvm.Config{GasLimit: math.MaxUint64, GenesisID: genesisID}))

		// our random account needs a balance so it can be spawned
		// this is not strictly necessary for the test vectors but it allows us to perform validation
		vm.ApplyGenesis([]types.Account{{
			Address: principal,
			Balance: constants.OneSmesh,
		}})

		// simple wallet
		for txtype, tx := range map[typeTx][]byte{
			Spawn:     sdkWallet.Spawn(priv, wallet.TemplateAddress, &wallet.SpawnArguments{PublicKey: pubCore}, nonce, opts...),
			SelfSpawn: sdkWallet.SelfSpawn(priv, nonce, opts...),
			Spend:     sdkWallet.Spend(priv, destination, amount, nonce, opts...),
		} {
			logrus.Debugf("Generating test vector for %s %s %s", hrp, "wallet", txtype)
			testVector := txToTestVector(tx, vm, index, amount, nonce, Wallet, txtype, destination.String(), hrp)
			testVectors = append(testVectors, testVector)
			index++
		}
	}
	return testVectors
}

func main() {
	testVectors := generateTestVectors()

	jsonData, err := json.MarshalIndent(testVectors, "", "  ")
	if err != nil {
		logrus.Fatalf("Error marshalling test vectors: %v", err)
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
	pub = signing.PublicKey{PublicKey: edPub}
	priv = signing.PrivateKey(edPriv)
	return
}
