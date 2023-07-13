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
	"github.com/cosmos/btcutil/bech32"
	"github.com/hashicorp/go-secure-stdlib/password"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spf13/cobra"

	"github.com/spacemeshos/smcli/cmd/internal"
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

	// printHex indicates that keys should be printed in Hex format.
	printHex bool

	// printParent indicates that the parent key should be printed.
	printParent bool

	// useLedger indicates that the Ledger device should be used.
	useLedger bool

	// noAddress indicates that the address should not be shown when printing a key.
	// This matches the old behavior of the read cmd.
	noAddress bool
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
	Use:                   "read [wallet file] [--full/-f] [--private/-p] [--parent] [--base58] [--hex] [--no-address]",
	DisableFlagsInUseLine: true,
	Short:                 "Reads an existing wallet file",
	Long: `This command can be used to verify whether an existing wallet file can be
successfully read and decrypted, whether the password to open the file is correct, etc.
It prints the accounts from the wallet file. By default it does not print private keys.
Add --private to print private keys. Add --full to print full keys. Add --base58 to print
keys in base58 format or --hex for hexdecimal rather than bech32. Add --parent to print parent key (and not
only child keys). Add --no-address to not print the address.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		w, err := internal.LoadWallet(args[0], debug)
		cobra.CheckErr(err)

		caption := make([]string, 0, 2)
		maxWidth := 20
		widthEnforcer := func(col string, maxLen int) string {
			if len(col) <= maxLen {
				return col
			}
			if maxLen <= 7 {
				return col[:maxLen]
			}
			return fmt.Sprintf("%s..%s", col[:maxLen-7], col[len(col)-5:])
		}

		header := table.Row{"pubkey", "path", "name", "created"}

		if printFull {
			// full key is 64 bytes which is 128 chars in hex, need to print at least this much
			maxWidth = 150
		} else {
			caption = append(caption, "To print full keys, use the --full flag.")
		}

		colCfgs := []table.ColumnConfig{
			{Number: 1, WidthMax: maxWidth, WidthMaxEnforcer: widthEnforcer},
		}

		if !noAddress {
			header = append(header[:3], header[2:]...)
			header[2] = "address"
		}

		if printPrivate {
			caption = append(caption, fmt.Sprintf("Mnemonic: %s", w.Mnemonic()))
			header = append(header[:2], header[1:]...)
			header[1] = "privkey"
			colCfgs = append(colCfgs, table.ColumnConfig{
				Number: 2, WidthMax: maxWidth, WidthMaxEnforcer: widthEnforcer,
			})
		}

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle("Wallet Contents")
		t.SetCaption(strings.Join(caption, "\n"))
		t.AppendHeader(header)
		t.SetColumnConfigs(colCfgs)

		// set the encoder
		var encoder func([]byte) string
		switch {
		case printBase58:
			encoder = base58.Encode
		case printHex:
			encoder = hex.EncodeToString
		default:
			encoder = func(data []byte) string {
				dataConverted, err := bech32.ConvertBits(data, 8, 5, true)
				cobra.CheckErr(err)
				encoded, err := bech32.Encode(types.NetworkHRP(), dataConverted)
				cobra.CheckErr(err)
				return encoded
			}
		}

		addRow := func(account *wallet.EDKeyPair) {
			row := make([]any, 0, 6) // Row len is 4 w/o address, up to 6 w/ priv key.
			row = append(row, encoder(account.Public))

			if printPrivate {
				privKey := "(none)"
				if len(account.Private) > 0 {
					privKey = encoder(account.Private)
				}

				row = append(row, privKey)
			}

			row = append(row, account.Path.String())

			if !noAddress {
				row = append(row, types.GenerateAddress(account.Public).String())
			}

			row = append(row, account.DisplayName, account.Created)

			t.AppendRow(row)
		}

		// print the master account
		if printParent {
			if master := w.Secrets.MasterKeypair; master != nil {
				addRow(master)
			}
		}

		// print child accounts
		for _, a := range w.Secrets.Accounts {
			addRow(a)
		}

		t.Render()
	},
}

var addrCmd = &cobra.Command{
	Use:                   "address [wallet file] [--parent]",
	DisableFlagsInUseLine: true,
	Short:                 "Show wallet addresses",
	Long:                  "Show the addresses associated with the given wallet",
	Args:                  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		w, err := internal.LoadWallet(args[0], debug)
		cobra.CheckErr(err)

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.SetTitle("Wallet Addresses")
		t.AppendHeader(table.Row{"address", "name"})

		if printParent {
			if master := w.Secrets.MasterKeypair; master != nil {
				t.AppendRow(table.Row{
					types.GenerateAddress(master.Public).String(),
					master.DisplayName,
				})
			}
		}

		for _, account := range w.Secrets.Accounts {
			t.AppendRow(table.Row{
				types.GenerateAddress(account.Public).String(),
				account.DisplayName,
			})
		}

		t.Render()
	},
}

func init() {
	rootCmd.AddCommand(walletCmd)
	walletCmd.AddCommand(createCmd)
	walletCmd.AddCommand(readCmd)
	walletCmd.AddCommand(addrCmd)
	readCmd.Flags().BoolVarP(&printPrivate, "private", "p", false, "Print private keys")
	readCmd.Flags().BoolVarP(&printFull, "full", "f", false, "Print full keys (no abbreviation)")
	readCmd.Flags().BoolVar(&printBase58, "base58", false, "Print keys in base58 (rather than bech32)")
	readCmd.Flags().BoolVar(&printHex, "hex", false, "Print keys in hex (rather than bech32)")
	readCmd.Flags().BoolVar(&printParent, "parent", false, "Print parent key (not only child keys)")
	readCmd.Flags().BoolVar(&noAddress, "no-address", false, "Do not print the address associated with the key")
	readCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")
	createCmd.Flags().BoolVarP(&useLedger, "ledger", "l", false, "Create a wallet using a Ledger device")
	addrCmd.Flags().BoolVar(&printParent, "parent", false, "Print parent address (not only child addresses)")
}
