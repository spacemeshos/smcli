/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var hwCmd = &cobra.Command{
	Use:     "hardware",
	Aliases: []string{"hw"},
	Short:   "Use a hardware wallet",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

var hwListCmd = &cobra.Command{
	Use:   "list",
	Short: "List connected hardware devices",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

func init() {
	walletCmd.AddCommand(hwCmd)
	hwCmd.AddCommand(hwListCmd)
}
