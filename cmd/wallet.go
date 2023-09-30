package cmd

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/genvm/sdk"
	"github.com/spacemeshos/go-spacemesh/signing"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strconv"
	"strings"

	api "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"google.golang.org/grpc"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hashicorp/go-secure-stdlib/password"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spacemeshos/go-spacemesh/common/types"
	walletSdk "github.com/spacemeshos/go-spacemesh/genvm/sdk/wallet"
	walletTemplate "github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"github.com/spf13/cobra"

	"github.com/spacemeshos/smcli/common"
	"github.com/spacemeshos/smcli/wallet"
)

var (
	// debug indicates that the program is in debug mode.
	debug bool

	// printPrivate indicates that private keys should be printed.
	printPrivate bool

	// printFull indicates that full keys should be printed (not abbreviated).
	printFull bool

	// printBase58 indicates that keys should be printed in base58 format.
	printBase58 bool

	// printParent indicates that the parent key should be printed.
	printParent bool

	// useLedger indicates that the Ledger device should be used.
	useLedger bool

	// hrp is the human-readable network identifier used in Spacemesh network addresses.
	hrp string

	// maxGas is the max amount to spend on gas.
	maxGas uint32

	// gasPrice is the price in smidge paid for one unit of gas.
	gasPrice uint8

	// nodeUri is the address of the node to talk to
	nodeUri string
)

// walletCmd represents the wallet command.
var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

// createCmd represents the create command.
var createCmd = &cobra.Command{
	Use:   "create [--ledger] [numaccounts]",
	Short: "Generate a new wallet file from a BIP-39-compatible mnemonic or Ledger device",
	Long: `Create a new wallet file containing one or more accounts using a BIP-39-compatible mnemonic
or a Ledger hardware wallet. If using a mnemonic you can choose to use an existing mnemonic or generate
a new, random mnemonic.

Add --ledger to instead read the public key from a Ledger device. If using a Ledger device please make
sure the device is connected, unlocked, and the Spacemesh app is open.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// get the number of accounts to create
		n := 1
		if len(args) > 0 {
			tmpN, err := strconv.ParseInt(args[0], 10, 16)
			cobra.CheckErr(err)
			n = int(tmpN)
		}

		var w *wallet.Wallet
		var err error

		// Short-circuit and check for a ledger device
		if useLedger {
			w, err = wallet.NewMultiWalletFromLedger(n)
			cobra.CheckErr(err)
			fmt.Println("Note that, when using a hardware wallet, the wallet file I'm about to produce won't " +
				"contain any private keys or mnemonics, but you may still choose to encrypt it to protect privacy.")
		} else {
			// get or generate the mnemonic
			fmt.Print("Enter a BIP-39-compatible mnemonic (or leave blank to generate a new one): ")
			text, err := password.Read(os.Stdin)
			fmt.Println()
			cobra.CheckErr(err)
			fmt.Println("Note: This application does not yet support BIP-39-compatible optional passwords. Support will be added soon.")

			// It's critical that we trim whitespace, including CRLF. Otherwise it will get included in the mnemonic.
			text = strings.TrimSpace(text)

			if text == "" {
				w, err = wallet.NewMultiWalletRandomMnemonic(n)
				cobra.CheckErr(err)
				fmt.Println("\nThis is your mnemonic (seed phrase). Write it down and store it safely. It is the ONLY way to restore your wallet.")
				fmt.Println("Neither Spacemesh nor anyone else can help you restore your wallet without this mnemonic.")
				fmt.Println("\n***********************************\nSAVE THIS MNEMONIC IN A SAFE PLACE!\n***********************************")
				fmt.Println()
				fmt.Println(w.Mnemonic())
				fmt.Println("\nPress enter when you have securely saved your mnemonic.")
				_, _ = fmt.Scanln()
			} else {
				// try to use as a mnemonic
				w, err = wallet.NewMultiWalletFromMnemonic(text, n)
				cobra.CheckErr(err)
			}
		}

		fmt.Print("Enter a secure password used to encrypt the wallet file (optional but strongly recommended): ")
		password, err := password.Read(os.Stdin)
		fmt.Println()
		cobra.CheckErr(err)
		wk := wallet.NewKey(wallet.WithRandomSalt(), wallet.WithPbkdf2Password([]byte(password)))
		err = os.MkdirAll(common.DotDirectory(), 0o700)
		cobra.CheckErr(err)

		// Make sure we're not overwriting an existing wallet (this should not happen)
		walletFn := common.WalletFile()
		_, err = os.Stat(walletFn)
		switch {
		case errors.Is(err, os.ErrNotExist):
			// all fine
		case err == nil:
			log.Fatalln("Wallet file already exists")
		default:
			log.Fatalf("Error opening %s: %v\n", walletFn, err)
		}

		// Now open for writing
		f2, err := os.OpenFile(walletFn, os.O_WRONLY|os.O_CREATE, 0o600)
		cobra.CheckErr(err)
		defer f2.Close()
		cobra.CheckErr(wk.Export(f2, w))

		fmt.Printf("Wallet saved to %s. BACK UP THIS FILE NOW!\n", walletFn)
	},
}

// sendCmd initiates a simple coin send transaction
var sendCmd = &cobra.Command{
	Use:   "send [wallet file] [sender] [recipient] [amount (in smidge)] [--maxGas N] [--gasPrice N] [--node N]",
	Short: "Sends coins from the specified wallet and account",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		walletFn := args[0]
		senderPubkeyHex := args[1]
		recipientAddressStr := args[2]
		amount, err := strconv.ParseUint(args[3], 10, 64)
		cobra.CheckErr(err)

		// make sure the file exists
		f, err := os.Open(walletFn)
		cobra.CheckErr(err)
		defer f.Close()

		// get the password
		fmt.Print("Enter wallet password: ")
		password, err := password.Read(os.Stdin)
		fmt.Println()
		cobra.CheckErr(err)

		// attempt to read it
		wk := wallet.NewKey(wallet.WithPasswordOnly([]byte(password)))
		w, err := wk.Open(f, debug)
		cobra.CheckErr(err)

		// parse the sender and make sure the wallet contains it
		var keypair *wallet.EDKeyPair
		for _, kp := range w.Secrets.Accounts {
			if hex.EncodeToString(kp.Public) == senderPubkeyHex {
				keypair = kp
				break
			}
		}
		if keypair == nil {
			log.Fatalln("pubkey not found in wallet file")
		}

		// parse principal address
		spawnargs := walletTemplate.SpawnArguments{}
		copy(spawnargs.PublicKey[:], signing.Public(signing.PrivateKey(keypair.Private)))
		principal := core.ComputePrincipal(walletTemplate.TemplateAddress, &spawnargs)
		fmt.Printf("principal: %s\n", principal.String())

		// parse the recipient address
		recipientAddress, err := types.StringToAddress(recipientAddressStr)
		cobra.CheckErr(err)

		// establish grpc link and read principal nonce and balance
		cc, err := grpc.Dial(nodeUri, grpc.WithTransportCredentials(insecure.NewCredentials()))
		cobra.CheckErr(err)
		defer cc.Close()

		meshClient := api.NewMeshServiceClient(cc)
		meshResp, err := meshClient.GenesisID(cmd.Context(), &api.GenesisIDRequest{})
		cobra.CheckErr(err)
		var genesisId types.Hash20
		copy(genesisId[:], meshResp.GenesisId)

		client := api.NewNodeServiceClient(cc)
		statusResp, err := client.Status(cmd.Context(), &api.StatusRequest{})
		cobra.CheckErr(err)
		fmt.Printf("Synced: %v\nPeers: %d\nSyncedLayer: %d\nTopLayer: %d\nVerifiedLayer: %d\nGenesisID: %s\n",
			statusResp.Status.IsSynced,
			statusResp.Status.ConnectedPeers,
			statusResp.Status.SyncedLayer.GetNumber(),
			statusResp.Status.TopLayer.GetNumber(),
			statusResp.Status.VerifiedLayer.GetNumber(),
			genesisId.String(),
		)

		gstate := api.NewGlobalStateServiceClient(cc)
		resp, err := gstate.Account(cmd.Context(), &api.AccountRequest{AccountId: &api.AccountId{Address: principal.String()}})
		cobra.CheckErr(err)
		nonce := resp.AccountWrapper.StateProjected.Counter
		balance := resp.AccountWrapper.StateProjected.Balance
		fmt.Printf("Sender nonce %d balance %d\n", nonce, balance.Value)

		// generate the tx
		tx := walletSdk.Spend(signing.PrivateKey(keypair.Private), recipientAddress, amount, nonce+1,
			sdk.WithGenesisID(genesisId),
			//sdk.WithGasPrice(0),
		)
		fmt.Printf("Generated signed tx: %s\n", hex.EncodeToString(tx))

		// parse it
		txService := api.NewTransactionServiceClient(cc)
		txResp, err := txService.ParseTransaction(cmd.Context(), &api.ParseTransactionRequest{Transaction: tx})
		cobra.CheckErr(err)
		fmt.Printf("parsed tx: principal: %s, gasprice: %d, maxgas: %d, nonce: %d\n",
			txResp.Tx.Principal.Address, txResp.Tx.GasPrice, txResp.Tx.MaxGas, txResp.Tx.Nonce.Counter)

		// broadcast it
		//txResp, err := txService.ParseTransaction(cmd.Context(), &api.ParseTransactionRequest{Transaction: tx})
		sendResp, err := txService.SubmitTransaction(cmd.Context(), &api.SubmitTransactionRequest{Transaction: tx})
		cobra.CheckErr(err)
		//fmt.Printf("status: %s, txid: %s, tx state: %s",
		//	txResp.String(), txResp.Tx.Id, txResp.Tx.String())

		// return the txid
		fmt.Printf("status code: %d, txid: %s, tx state: %s\n",
			sendResp.Status.Code, hex.EncodeToString(sendResp.Txstate.Id.Id), sendResp.Txstate.State.String())
	},
}

// readCmd reads an existing wallet file.
var readCmd = &cobra.Command{
	Use:   "read [wallet file] [--full/-f] [--private/-p] [--base58]",
	Short: "Reads an existing wallet file",
	Long: `This command can be used to verify whether an existing wallet file can be
successfully read and decrypted, whether the password to open the file is correct, etc.
It prints the accounts from the wallet file. By default it does not print private keys.
Add --private to print private keys. Add --full to print full keys. Add --base58 to print
keys in base58 format rather than hexadecimal. Add --parent to print parent key (and not
only child keys).`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		walletFn := args[0]

		// make sure the file exists
		f, err := os.Open(walletFn)
		cobra.CheckErr(err)
		defer f.Close()

		// get the password
		fmt.Print("Enter wallet password: ")
		password, err := password.Read(os.Stdin)
		fmt.Println()
		cobra.CheckErr(err)

		// attempt to read it
		wk := wallet.NewKey(wallet.WithPasswordOnly([]byte(password)))
		w, err := wk.Open(f, debug)
		cobra.CheckErr(err)

		widthEnforcer := func(col string, maxLen int) string {
			if len(col) <= maxLen {
				return col
			}
			if maxLen <= 7 {
				return col[:maxLen]
			}
			return fmt.Sprintf("%s..%s", col[:maxLen-7], col[len(col)-5:])
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle("Wallet Contents")
		caption := ""
		if printPrivate {
			caption = fmt.Sprintf("Mnemonic: %s", w.Mnemonic())
		}
		if !printFull {
			if printPrivate {
				caption += "\n"
			}
			caption += "To print full keys, use the --full flag."
		}
		t.SetCaption(caption)
		maxWidth := 20
		if printFull {
			// full key is 64 bytes which is 128 chars in hex, need to print at least this much
			maxWidth = 150
		}
		if printPrivate {
			t.AppendHeader(table.Row{
				"address",
				"pubkey",
				"privkey",
				"path",
				"name",
				"created",
			})
			t.SetColumnConfigs([]table.ColumnConfig{
				{Number: 2, WidthMax: maxWidth, WidthMaxEnforcer: widthEnforcer},
				{Number: 3, WidthMax: maxWidth, WidthMaxEnforcer: widthEnforcer},
			})
		} else {
			t.AppendHeader(table.Row{
				"address",
				"pubkey",
				"path",
				"name",
				"created",
			})
			t.SetColumnConfigs([]table.ColumnConfig{
				{Number: 2, WidthMax: maxWidth, WidthMaxEnforcer: widthEnforcer},
			})
		}

		// set the encoder
		encoder := hex.EncodeToString
		if printBase58 {
			encoder = base58.Encode
		}

		privKeyEncoder := func(privKey []byte) string {
			if len(privKey) == 0 {
				return "(none)"
			}
			return encoder(privKey)
		}

		// print the master account
		if printParent {
			master := w.Secrets.MasterKeypair
			if master != nil {
				if printPrivate {
					t.AppendRow(table.Row{
						"N/A",
						encoder(master.Public),
						privKeyEncoder(master.Private),
						master.Path.String(),
						master.DisplayName,
						master.Created,
					})
				} else {
					t.AppendRow(table.Row{
						"N/A",
						encoder(master.Public),
						master.Path.String(),
						master.DisplayName,
						master.Created,
					})
				}
			}
		}

		// print child accounts
		for _, a := range w.Secrets.Accounts {
			if printPrivate {
				t.AppendRow(table.Row{
					wallet.PubkeyToAddress(a.Public, hrp),
					encoder(a.Public),
					privKeyEncoder(a.Private),
					a.Path.String(),
					a.DisplayName,
					a.Created,
				})
			} else {
				t.AppendRow(table.Row{
					wallet.PubkeyToAddress(a.Public, hrp),
					encoder(a.Public),
					a.Path.String(),
					a.DisplayName,
					a.Created,
				})
			}
		}
		t.Render()
	},
}

func init() {
	rootCmd.AddCommand(walletCmd)
	walletCmd.AddCommand(createCmd)
	walletCmd.AddCommand(readCmd)
	walletCmd.AddCommand(sendCmd)
	sendCmd.Flags().Uint8Var(&gasPrice, "gasPrice", 1, "Set gas price")
	sendCmd.Flags().Uint32Var(&maxGas, "maxGas", 0, "Set max gas")
	sendCmd.Flags().StringVar(&nodeUri, "nodeUri", "localhost:9092", "Node URI")
	readCmd.Flags().BoolVarP(&printPrivate, "private", "p", false, "Print private keys")
	readCmd.Flags().BoolVarP(&printFull, "full", "f", false, "Print full keys (no abbreviation)")
	readCmd.Flags().BoolVar(&printBase58, "base58", false, "Print keys in base58 (rather than hex)")
	readCmd.Flags().BoolVar(&printParent, "parent", false, "Print parent key (not only child keys)")
	readCmd.Flags().StringVar(&hrp, "hrp", types.NetworkHRP(), "Set human-readable address prefix")
	readCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")
	createCmd.Flags().BoolVarP(&useLedger, "ledger", "l", false, "Create a wallet using a Ledger device")
}
