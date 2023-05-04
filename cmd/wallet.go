/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"github.com/hashicorp/go-secure-stdlib/password"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spacemeshos/smcli/common"
	"github.com/spacemeshos/smcli/wallet"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strconv"
	"strings"
)

var (
	// debug indicates that the program is in debug mode
	debug,

	// printPrivate indicates that private keys should be printed
	printPrivate,

	// printFull indicates that full keys should be printed (not abbreviated)
	printFull,

	// printBase58 indicates that keys should be printed in base58 format
	printBase58 bool
)

// walletCmd represents the wallet command
var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [numaccounts]",
	Short: "Generate a new wallet file from a BIP-39-compatible mnemonic",
	Long: `Create a new wallet file containing one or more accounts using a BIP-39-compatible mnemonic.
You can choose to use an existing mnemonic or generate a new, random mnemonic.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// get the number of accounts to create
		n := 1
		if len(args) > 0 {
			tmpN, err := strconv.ParseInt(args[0], 10, 16)
			cobra.CheckErr(err)
			n = int(tmpN)
		}

		// get or generate the mnemonic
		fmt.Print("Enter a BIP-39-compatible mnemonic (or leave blank to generate a new one): ")
		text, err := password.Read(os.Stdin)
		fmt.Println()
		cobra.CheckErr(err)
		fmt.Println("Note: This application does not yet support BIP-39-compatible optional passwords. Support will be added soon.")

		// It's critical that we trim whitespace, including CRLF. Otherwise it will get included in the mnemonic.
		text = strings.TrimSpace(text)

		var w *wallet.Wallet
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

		fmt.Print("Enter a secure password used to encrypt the wallet file (optional but strongly recommended): ")
		password, err := password.Read(os.Stdin)
		fmt.Println()
		cobra.CheckErr(err)
		wk := wallet.NewKey(wallet.WithRandomSalt(), wallet.WithPbkdf2Password([]byte(password)))
		err = os.MkdirAll(common.DotDirectory(), 0700)
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
		f2, err := os.OpenFile(walletFn, os.O_WRONLY|os.O_CREATE, 0600)
		cobra.CheckErr(err)
		defer f2.Close()
		cobra.CheckErr(wk.Export(f2, w))

		fmt.Printf("Wallet saved to %s. BACK UP THIS FILE NOW!\n", walletFn)
	},
}

// readCmd reads an existing wallet file
var readCmd = &cobra.Command{
	Use:   "read [wallet file] [--full/-f] [--private/-p] [--base58]",
	Short: "Reads an existing wallet file",
	Long: `This command can be used to verify whether an existing wallet file can be
successfully read and decrypted, whether the password to open the file is correct, etc.
It prints the accounts from the wallet file. By default it does not print private keys.
Add --private to print private keys. Add --full to print full keys. Add --base58 to print
keys in base58 format rather than hexidecimal.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		walletFn := args[0]

		// make sure the file exists
		f, err := os.Open(walletFn)
		defer f.Close()
		cobra.CheckErr(err)

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
		// TODO: add spacemesh address format (bech32)
		// https://github.com/spacemeshos/smcli/issues/38
		if printPrivate {
			t.AppendHeader(table.Row{
				"pubkey",
				"privkey",
				"path",
				"name",
				"created",
			})
			t.SetColumnConfigs([]table.ColumnConfig{
				{Number: 1, WidthMax: maxWidth, WidthMaxEnforcer: widthEnforcer},
				{Number: 2, WidthMax: maxWidth, WidthMaxEnforcer: widthEnforcer},
			})
		} else {
			t.AppendHeader(table.Row{
				"pubkey",
				"path",
				"name",
				"created",
			})
			t.SetColumnConfigs([]table.ColumnConfig{
				{Number: 1, WidthMax: maxWidth, WidthMaxEnforcer: widthEnforcer},
			})
		}

		// print the master keypair
		master := w.Secrets.MasterKeypair
		encoder := hex.EncodeToString
		if printBase58 {
			encoder = base58.Encode
		}
		if master != nil {
			if printPrivate {
				t.AppendRow(table.Row{
					encoder(master.Public),
					encoder(master.Private),
					master.Path.String(),
					master.DisplayName,
					master.Created,
				})
			} else {
				t.AppendRow(table.Row{
					encoder(master.Public),
					master.Path.String(),
					master.DisplayName,
					master.Created,
				})
			}
		}

		for _, a := range w.Secrets.Accounts {
			if printPrivate {
				t.AppendRow(table.Row{
					encoder(a.Public),
					encoder(a.Private),
					a.Path.String(),
					a.DisplayName,
					a.Created,
				})
			} else {
				t.AppendRow(table.Row{
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
	readCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debug mode")
}
