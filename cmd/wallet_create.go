/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/hashicorp/go-secure-stdlib/password"
	"github.com/spacemeshos/smcli/wallet"
	"github.com/spf13/cobra"
)

var walletPath string

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		password, err := password.Read(os.Stdin)
		cobra.CheckErr(err)
		wk := wallet.NewWalletKey(
			wallet.WithPassword(password),
		)
		ws := wallet.NewWalletStore(wk)
		w := wallet.NewWallet()
		cobra.CheckErr(ws.Export(walletPath, w))
	},
}

func init() {
	walletCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&walletPath, "output", "o", "", "Wallet location (full path)")
}
