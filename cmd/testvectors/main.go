package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"

	"github.com/oasisprotocol/curve25519-voi/primitives/ed25519"
	"github.com/spacemeshos/economics/constants"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/config"
	"github.com/spacemeshos/go-spacemesh/config/presets"
	genvm "github.com/spacemeshos/go-spacemesh/genvm"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/genvm/sdk"
	"github.com/spacemeshos/go-spacemesh/log"

	sdkMultisig "github.com/spacemeshos/go-spacemesh/genvm/sdk/multisig"
	sdkVesting "github.com/spacemeshos/go-spacemesh/genvm/sdk/vesting"
	sdkWallet "github.com/spacemeshos/go-spacemesh/genvm/sdk/wallet"
	templateMultisig "github.com/spacemeshos/go-spacemesh/genvm/templates/multisig"
	templateVesting "github.com/spacemeshos/go-spacemesh/genvm/templates/vesting"
	templateWallet "github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"github.com/spacemeshos/go-spacemesh/signing"
	"github.com/spacemeshos/go-spacemesh/sql"
	"go.uber.org/zap"
)

type typeAccount string

const (
	Wallet   typeAccount = "wallet"
	Multisig typeAccount = "multisig"
	Vault    typeAccount = "vault"
	Vesting  typeAccount = "vesting"
)

type typeTx string

const (
	Spawn     typeTx = "spawn"
	SelfSpawn typeTx = "self_spawn"
	Spend     typeTx = "spend"
	// applied to state but not output in tests
	Ignore typeTx = "ignore"
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
	level := zap.DebugLevel
	// level := zap.InfoLevel
	if os.Getenv("DEBUG") != "" {
		level = zap.DebugLevel
	}
	log.SetLogger(log.NewWithLevel("testvectors", zap.NewAtomicLevelAt(level)))
}

// generate a random address for testing
func generateAddress() types.Address {
	pub, _ := getKey()
	return types.GenerateAddress(pub)
}

func applyTx(tx []byte, vm *genvm.VM) {
	validator := vm.Validation(types.NewRawTx(tx))
	header, err := validator.Parse()
	if err != nil {
		log.Fatal("Error parsing transaction to apply: %v", err)
	}
	coreTx := types.Transaction{
		TxHeader: header,
		RawTx:    types.NewRawTx(tx),
	}
	skipped, results, err := vm.Apply(genvm.ApplyContext{Layer: types.FirstEffectiveGenesis()}, []types.Transaction{coreTx}, []types.CoinbaseReward{})
	if len(skipped) != 0 {
		log.Fatal("Error applying transaction")
	} else if len(results) != 1 || results[0].Status != types.TransactionSuccess {
		log.Fatal("Error applying transaction: %v", results[0].Status)
	} else if err != nil {
		log.Fatal("Error applying transaction: %v", err)
	}
	log.Debug("got result: %v", results[0].TransactionResult)
}

// m, n only used for multisig; ignored for single sig wallet
func txToTestVector(
	tx []byte,
	vm *genvm.VM,
	index int,
	amount, nonce uint64,
	accountType typeAccount,
	txType typeTx,
	destination, hrp string,
	m, n uint8,
	validity bool,
) TestVector {
	validator := vm.Validation(types.NewRawTx(tx))
	header, err := validator.Parse()
	if err != nil {
		log.Fatal("Error parsing transaction idx %d: %v", index, err)
	}

	// we should be able to validate all txs EXCEPT partially aggregated multisig txs,
	// which are not valid as standalone txs
	if !validator.Verify() {
		if validity {
			log.Fatal("Error validating supposedly valid transaction idx %d", index)
		}
		log.Debug("Expected error parsing partially aggregated transaction idx %d: %v, ignoring", index, err)
	}
	return TestVector{
		Index: index,
		Name:  fmt.Sprintf("%s_%s_%s", hrp, accountType, txType),
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

type TxPair struct {
	txtype typeTx
	tx     []byte

	// whether this tx is valid as a standalone tx
	// partially-aggregated multisig txs are not standalone valid so we don't attempt to validate them!
	valid bool
}

// maximum "n" value for multisig
const MaxKeys = 2

func handleMultisig(
	vm *genvm.VM,
	opts []sdk.Opt,
	destination types.Address,
	index int,
	hrp string,
	pubkeysSigning []signing.PublicKey,
	pubkeysCore []core.PublicKey,
	pubkeysEd []ed25519.PublicKey,
	privkeys []ed25519.PrivateKey,
	m, n uint8,
) []TestVector {
	spawnArgsMultisig := &templateMultisig.SpawnArguments{
		Required:   m,
		PublicKeys: pubkeysCore[:n],
	}

	// principal address depends on the set of pubkeys
	principal := core.ComputePrincipal(templateMultisig.TemplateAddress, spawnArgsMultisig)
	log.Debug("MULTISIG: %d of %d, principal: %s", m, n, principal.String())

	// fund the principal account (to allow verification later)
	vm.ApplyGenesis([]types.Account{{
		Address: principal,
		Balance: constants.OneSmesh,
	}})

	// multisig operations require m signers per operation
	spawnAgg := sdkMultisig.Spawn(0, privkeys[0], principal, templateMultisig.TemplateAddress, spawnArgsMultisig, Nonce, opts...)
	selfSpawnAgg := sdkMultisig.SelfSpawn(0, privkeys[0], templateMultisig.TemplateAddress, m, pubkeysEd[:n], Nonce, opts...)
	spendAgg := sdkMultisig.Spend(0, privkeys[0], principal, destination, Amount, Nonce, opts...)

	// add an individual test vector for each signing operation
	// one list per tx type so we can assemble the final list in order
	// start with the first operation
	// three m-length lists plus one additional, final, aggregated self-spawn tx
	txList := []TxPair{}
	txListSpawn := make([]TxPair, m)
	txListSelfSpawn := make([]TxPair, m)
	txListSpend := make([]TxPair, m)

	// multisig txs are valid as standalone only if idx==m-1, i.e., it's the final part
	txListSpawn[0] = TxPair{txtype: Spawn, tx: spawnAgg.Raw(), valid: m == 1}
	txListSelfSpawn[0] = TxPair{txtype: SelfSpawn, tx: selfSpawnAgg.Raw(), valid: m == 1}
	txListSpend[0] = TxPair{txtype: Spend, tx: spendAgg.Raw(), valid: m == 1}

	// now add a test vector for each additional required signature
	// note: this assumes signer n has the signed n-1 tx
	for signerIdx := uint8(1); signerIdx < m; signerIdx++ {
		spawnAgg.Add(*sdkMultisig.Spawn(signerIdx, privkeys[signerIdx], principal, templateMultisig.TemplateAddress, spawnArgsMultisig, Nonce, opts...).Part(signerIdx))
		selfSpawnAgg.Add(*sdkMultisig.SelfSpawn(signerIdx, privkeys[signerIdx], templateMultisig.TemplateAddress, m, pubkeysEd[:n], Nonce, opts...).Part(signerIdx))
		spendAgg.Add(*sdkMultisig.Spend(signerIdx, privkeys[signerIdx], principal, destination, Amount, Nonce, opts...).Part(signerIdx))

		// only the final, fully aggregated tx is valid
		txListSpawn[signerIdx] = TxPair{txtype: Spawn, tx: spawnAgg.Raw(), valid: signerIdx == m-1}
		txListSelfSpawn[signerIdx] = TxPair{txtype: SelfSpawn, tx: selfSpawnAgg.Raw(), valid: signerIdx == m-1}
		txListSpend[signerIdx] = TxPair{txtype: Spend, tx: spendAgg.Raw(), valid: signerIdx == m-1}
	}

	// assemble the final list of txs in order: spawn, self-spawn, final aggregated self-spawn to apply, spend
	txList = append(txList, txListSpawn...)
	txList = append(txList, txListSelfSpawn...)
	txList = append(txList, TxPair{txtype: Ignore, tx: selfSpawnAgg.Raw()})
	txList = append(txList, txListSpend...)

	testVectors := []TestVector{}
	for _, txPair := range txList {
		if txPair.txtype == Ignore {
			log.Debug("Applying tx ignored for test vectors for %s %s", hrp, "multisig")
			applyTx(txPair.tx, vm)
			continue
		}
		log.Debug("[%d] Generating test vector for %s %s %s %d of %d", index, hrp, "multisig", txPair.txtype, m, n)
		testVector := txToTestVector(txPair.tx, vm, index, Amount, Nonce, Multisig, txPair.txtype, destination.String(), hrp, m, n, txPair.valid)
		testVectors = append(testVectors, testVector)
		index++
	}
	return testVectors
}

const (
	Nonce  = uint64(0)
	Amount = uint64(constants.OneSmesh)
)

func generateTestVectors(
	pubkeysSigning []signing.PublicKey,
	pubkeysCore []core.PublicKey,
	pubkeysEd []ed25519.PublicKey,
	privkeys []ed25519.PrivateKey,
) []TestVector {
	// read network configs - needed for genesisID
	var configMainnet, configTestnet config.GenesisConfig
	configMainnet = config.MainnetConfig().Genesis
	if testnet, err := presets.Get("testnet"); err != nil {
		log.Fatal("Error getting testnet config: %v", err)
	} else {
		configTestnet = testnet.Genesis
	}
	// not sure how to get hrp programmatically from config so we just hardcode it
	networks := map[string]config.GenesisConfig{
		"sm":    configMainnet,
		"stest": configTestnet,
	}

	testVectors := []TestVector{}
	index := 0
	// just use a single, random destination address
	// note: destination is not used in all tx types
	destination := generateAddress()
	for hrp, netconf := range networks {
		log.Debug("NETWORK: %s", hrp)
		// hrp is used in address generation
		types.SetNetworkHRP(hrp)

		// initialization
		genesisID := netconf.GenesisID()
		opts := []sdk.Opt{
			sdk.WithGenesisID(genesisID),
			sdk.WithGasPrice(1),
		}

		// we need a VM object for validation and gas cost computation
		vm := genvm.New(
			sql.InMemory(),
			genvm.WithConfig(genvm.Config{GasLimit: math.MaxUint64, GenesisID: genesisID}),
			genvm.WithLogger(log.NewWithLevel("genvm", zap.NewAtomicLevelAt(zap.DebugLevel))),
		)

		// SIMPLE WALLET (SINGLE SIG)
		log.Debug("TEMPLATE: WALLET")
		spawnArgsWallet := &templateWallet.SpawnArguments{
			PublicKey: pubkeysCore[0],
		}
		principal := core.ComputePrincipal(templateWallet.TemplateAddress, spawnArgsWallet)

		// our random account needs a balance so it can be spawned
		// this is not strictly necessary for the test vectors but it allows us to perform validation
		vm.ApplyGenesis([]types.Account{{
			Address: principal,
			Balance: constants.OneSmesh,
		}})

		// need a list, not a map, since order matters here
		// (self-spawn must come before spend)
		// simple wallet txs are always valid as standalone
		txList := []TxPair{
			{txtype: Spawn, tx: sdkWallet.Spawn(privkeys[0], templateWallet.TemplateAddress, spawnArgsWallet, Nonce, opts...), valid: true},
			{txtype: SelfSpawn, tx: sdkWallet.SelfSpawn(privkeys[0], Nonce, opts...), valid: true},
			// apply the parsed self spawn tx
			// this will allow the spend tx to be validated
			{txtype: Ignore, tx: sdkWallet.SelfSpawn(privkeys[0], Nonce, opts...)},
			{txtype: Spend, tx: sdkWallet.Spend(privkeys[0], destination, Amount, Nonce, opts...), valid: true},
		}
		for _, txPair := range txList {
			if txPair.txtype == Ignore {
				log.Debug("Applying tx ignored for test vectors for %s %s", hrp, "wallet")
				applyTx(txPair.tx, vm)
				continue
			}
			log.Debug("[%d] Generating test vector for %s %s %s", index, hrp, "wallet", txPair.txtype)
			testVector := txToTestVector(txPair.tx, vm, index, Amount, Nonce, Wallet, txPair.txtype, destination.String(), hrp, 1, 1, true)
			testVectors = append(testVectors, testVector)
			index++
		}

		// MULTISIG
		// 1-of-1, 1-of-2, 2-of-2
		log.Debug("TEMPLATE: MULTISIG")
		for _, n := range []uint8{1, MaxKeys} {
			for m := uint8(1); m <= n; m++ {
				multisigVectors := handleMultisig(
					vm,
					opts,
					destination,
					index,
					hrp,
					pubkeysSigning,
					pubkeysCore,
					pubkeysEd,
					privkeys,
					m,
					n,
				)
				testVectors = append(testVectors, multisigVectors...)
				index += len(multisigVectors)
			}
		}
	}
	return testVectors
}

func main() {
	// generate the required set of keypairs
	// do this once and use the same keys for all test vectors

	// frustratingly, we need the same list of pubkeys in multiple formats
	// https://github.com/spacemeshos/go-spacemesh/issues/6061
	pubkeysSigning := make([]signing.PublicKey, MaxKeys)
	pubkeysCore := make([]core.PublicKey, MaxKeys)
	pubkeysEd := make([]ed25519.PublicKey, MaxKeys)
	privkeys := make([]signing.PrivateKey, MaxKeys)
	for i := 0; i < MaxKeys; i++ {
		pubkeysEd[i], privkeys[i] = getKey()
		pubkeysCore[i] = types.BytesToHash(pubkeysEd[i])
		pubkeysSigning[i] = signing.PublicKey{PublicKey: pubkeysEd[i]}
	}

	testVectors := generateTestVectors(pubkeysSigning, pubkeysCore, pubkeysEd, privkeys)

	jsonData, err := json.MarshalIndent(testVectors, "", "  ")
	if err != nil {
		log.Fatal("Error marshalling test vectors: %v", err)
	}

	fmt.Println(string(jsonData))
}

// func getKey() (pub signing.PublicKey, priv signing.PrivateKey) {
func getKey() (pub ed25519.PublicKey, priv ed25519.PrivateKey) {
	// generate a random keypair
	pub, priv, err := ed25519.GenerateKey(rand.New(rand.NewSource(rand.Int63())))
	if err != nil {
		log.Fatal("failed to generate ed25519 key")
	}
	// pub = *signing.NewPublicKey(edPub)
	// pub = signing.PublicKey{PublicKey: edPub}
	// priv = signing.PrivateKey(edPriv)
	return
}
