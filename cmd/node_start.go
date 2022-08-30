/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spacemeshos/smcli/common"
	"github.com/spacemeshos/smcli/util"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a node.",
	Long: `Start a spacemesh node. Will automatically download and install 
the latest version of go-spacemesh if not already installed.`,
	Run: func(cmd *cobra.Command, args []string) {
		common.InitDotDir()

		// Check if node is running according to state file
		sp := common.NewStateProvider()
		if running := sp.NodeIsRunning(); running {
			fmt.Println("Node is already running with pid:", sp.GetNodePid())
			return
		}
		// Check if node bin file exists
		if _, err := os.Stat(common.NodeBin()); os.IsNotExist(err) {
			fmt.Println("Download Started")
			fileUrl := common.NodeDownloadUrl()
			zipFileName := "node.zip"
			if err := util.DownloadFile(zipFileName, fileUrl); err != nil {
				panic(err)
			}
			fmt.Println("Download Finished")
			fmt.Println("Unpacking...")
			fmt.Printf("Creating directory %s", common.BinDirectory())

			err := util.Unzip(zipFileName, common.BinDirectory())
			cobra.CheckErr(err)
			err = os.Remove(zipFileName)
			cobra.CheckErr(err)

			fmt.Println("\nDone")
		}

		fmt.Println("Starting Node...")

		nodePath := common.NodeBin()

		nodeProc := exec.Command(nodePath,
			"--listen", "/ip4/0.0.0.0/tcp/7513", // TODO(jonZlotnik): passthrough port flag
			"--config", common.NodeConfigFile(),
			"--data-folder", common.NodeDataDirectory())
		nodeLogFile, err := common.OpenNodeLogFile()
		cobra.CheckErr(err)

		nodeProc.Stdout = nodeLogFile
		nodeProc.Stderr = nodeLogFile

		err = nodeProc.Start()
		cobra.CheckErr(err)
		sp.UpdateNodePid(nodeProc.Process.Pid)
		fmt.Printf("Just launched %d, exiting\n", nodeProc.Process.Pid)
	},
}

func init() {
	nodeCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
