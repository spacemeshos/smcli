/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/hashicorp/go-secure-stdlib/password"
	"github.com/spacemeshos/smcli/common"
	"github.com/spacemeshos/smcli/wallet"
	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
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
		password, err := password.Read(os.Stdin)
		cobra.CheckErr(err)
		wk := wallet.NewKey(
			wallet.WithArgon2idPassword(password),
		)
		ws := wallet.NewStore(wk)
		e, _ := bip39.NewEntropy(256)
		m, _ := bip39.NewMnemonic(e)
		w := wallet.WalletFromMnemonic(m)
		err = os.MkdirAll(common.DotDirectory(), 0700)
		cobra.CheckErr(err)
		f, err := os.OpenFile(common.WalletFile(), os.O_WRONLY|os.O_CREATE, 0600)
		cobra.CheckErr(err)
		cobra.CheckErr(
			ws.Export(f, w),
		)
	},
}

func init() {
	walletCmd.AddCommand(createCmd)
}
