/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/spacemeshos/smcli/config"
	"github.com/spacemeshos/smcli/resources"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start called")
		fmt.Println("Download Started")
		fileUrl := config.NodeDownloadUrl()
		zipFileName := "node.zip"
		if err := resources.DownloadFile(zipFileName, fileUrl); err != nil {
			panic(err)
		}
		fmt.Println("Download Finished")
		fmt.Println("Unpacking...")
		fmt.Println("Creating directory...")
		binPath := filepath.Join(config.DefaultConfigPath(), "bin")
		fmt.Printf("Creating directory... %s", binPath)
		if err := os.MkdirAll(binPath, 0755); err != nil {
			cobra.CheckErr(err)
		}
		fmt.Println("Creating directory...")

		if err := resources.Unzip(zipFileName, binPath); err != nil {
			cobra.CheckErr(err)
		}
		fmt.Println("Done")

		fmt.Println("Starting Node...")
		nodePath := path.Join(binPath, "Linux", "go-spacemesh")

		nodeProc := exec.Command(nodePath,
			"--listen", "/ip4/0.0.0.0/tcp/7513",
			"--config", filepath.Join(binPath, "Linux", "config.json"))
		nodeProc.Stdout = os.Stdout
		err := nodeProc.Start()

		if err != nil {
			cobra.CheckErr(err)
		}
		fmt.Printf("Just ran subprocess %d, exiting\n", nodeProc.Process.Pid)
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
