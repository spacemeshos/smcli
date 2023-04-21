/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hashicorp/go-secure-stdlib/password"
	"github.com/spacemeshos/smcli/common"
	"github.com/spacemeshos/smcli/wallet"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strconv"
	"strings"
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
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("wallet called")
	// },
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
		reader := bufio.NewReader(os.Stdin)
		var err error
		text, err := reader.ReadString('\n')
		cobra.CheckErr(err)

		// It's critical that we trim whitespace, including CRLF. Otherwise it will get included in the mnemonic.
		text = strings.TrimSpace(text)

		var w *wallet.Wallet
		// TODO: check if we see \r\n on windows
		if text == "" {
			w, err = wallet.NewMultiWalletRandomMnemonic(n)
			cobra.CheckErr(err)
			fmt.Println("SAVE THIS MNEMONIC IN A SAFE PLACE!")
			fmt.Println(w.Mnemonic())
		} else {
			// try to use as a mnemonic
			w, err = wallet.NewMultiWalletFromMnemonic(text, n)
			cobra.CheckErr(err)
		}

		fmt.Print("Enter a secure password (optional but strongly recommended): ")
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
	Use:   "read [wallet file]",
	Short: "Reads an existing wallet file",
	Long: `This command can be used to verify whether an existing wallet file can be
successfully read and decrypted, whether the password to open the file is correct, etc.
It prints the accounts from the wallet file.`,
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
		w, err := wk.Open(f)
		cobra.CheckErr(err)

		//fmt.Println("Mnemonic:", w.Mnemonic())
		fmt.Println("Accounts:")
		for _, a := range w.Secrets.Accounts {
			err := json.NewEncoder(os.Stdout).Encode(a)
			cobra.CheckErr(err)
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(walletCmd)
	walletCmd.AddCommand(createCmd)
	walletCmd.AddCommand(readCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// walletCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// walletCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
