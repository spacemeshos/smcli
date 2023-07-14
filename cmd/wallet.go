package cmd

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hashicorp/go-secure-stdlib/password"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"

	"github.com/spacemeshos/go-spacemesh/common/types"
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

	// hrp is the human-readable network identifier used in Spacemesh network addresses
	hrp string
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
	readCmd.Flags().BoolVarP(&printPrivate, "private", "p", false, "Print private keys")
	readCmd.Flags().BoolVarP(&printFull, "full", "f", false, "Print full keys (no abbreviation)")
	readCmd.Flags().BoolVar(&printBase58, "base58", false, "Print keys in base58 (rather than hex)")
	readCmd.Flags().BoolVar(&printParent, "parent", false, "Print parent key (not only child keys)")
	readCmd.Flags().StringVar(&hrp, "hrp", types.NetworkHRP(), "Set human-readable address prefix")
	readCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")
	createCmd.Flags().BoolVarP(&useLedger, "ledger", "l", false, "Create a wallet using a Ledger device")
}
