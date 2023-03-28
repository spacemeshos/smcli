/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"github.com/tyler-smith/go-bip39"
	"log"
	"os"

	"github.com/hashicorp/go-secure-stdlib/password"
	"github.com/spacemeshos/smcli/common"
	"github.com/spacemeshos/smcli/wallet"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create single new random wallet",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Enter 12- or 24-word mnemonic (leave blank to generate a new one): ")
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		var w *wallet.Wallet
		// TODO: check if we see \r\n on windows
		if text == "\n" {
			e, _ := bip39.NewEntropy(256)
			m, _ := bip39.NewMnemonic(e)
			w = wallet.WalletFromMnemonic(m)
			fmt.Println("SAVE THIS MNEMONIC IN A SAFE PLACE!")
			fmt.Println(m)
		} else {
			// try to use as a mnemonic
			w = wallet.WalletFromMnemonic(text)
		}

		fmt.Print("Enter a secure password (optional but strongly recommended): ")
		password, err := password.Read(os.Stdin)
		cobra.CheckErr(err)
		wk := wallet.NewKey(wallet.WithPbkdf2Password(password))
		ws := wallet.NewStore(wk)
		err = os.MkdirAll(common.DotDirectory(), 0700)
		cobra.CheckErr(err)

		// Make sure we're not overwriting an existing wallet (this should not happen)
		walletFn := common.WalletFile()
		f, _ := os.Open(walletFn)
		if f != nil {
			log.Panic("Wallet file already exists", walletFn)
		}

		// Now open for writing
		f, err = os.OpenFile(walletFn, os.O_WRONLY|os.O_CREATE, 0600)
		cobra.CheckErr(err)
		cobra.CheckErr(ws.Export(f, w))

		fmt.Printf("Wallet saved to %s. BACK UP THIS FILE NOW!\n", walletFn)
	},
}

func init() {
	walletCmd.AddCommand(createCmd)
}
