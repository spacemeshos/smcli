/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	smapi "github.com/spacemeshos/api/release/go/spacemesh/v1"
	"github.com/spacemeshos/smcli/common"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		cc, _ := grpc.Dial(common.GetGRPCServerAddr(), grpc.WithInsecure())
		defer cc.Close()
		client := smapi.NewNodeServiceClient(cc)

		statusResp, err := client.Status(cmd.Context(), &smapi.StatusRequest{})
		cobra.CheckErr(err)
		fmt.Printf("Synced: %v\nPeers: %d\nSyncedLayer: %d\nTopLayer: %d\nVerifiedLayer: %d\n",
			statusResp.Status.IsSynced,
			statusResp.Status.ConnectedPeers,
			statusResp.Status.SyncedLayer.GetNumber(),
			statusResp.Status.TopLayer.GetNumber(),
			statusResp.Status.VerifiedLayer.GetNumber(),
		)
	},
}

func init() {
	nodeCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
