/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"syscall"

	"github.com/shirou/gopsutil/v3/process"
	"github.com/spacemeshos/smcli/common"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running node.",
	Long:  `Stop a running node.`,
	Run: func(cmd *cobra.Command, args []string) {
		common.InitDotDir()
		sp := common.NewStateProvider()
		if running := sp.NodeIsRunning(); !running {
			fmt.Println("Node is not running")
			return
		}
		// TODO(jonZlotnik): add more thorough checks for node state
		// like ps -a | grep go-spacemesh
		pid := sp.GetNodePid()
		fmt.Println("Stopping Node with pid:", pid)
		sp.UpdateNodePid(-1)
		p, err := process.NewProcess(int32(pid))
		cobra.CheckErr(err)
		err = p.SendSignal(syscall.SIGINT)
		cobra.CheckErr(err)

	},
}

func init() {
	nodeCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// stopCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// stopCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
