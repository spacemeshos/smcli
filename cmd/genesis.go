package cmd

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/spacemeshos/economics/constants"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/multisig"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/vault"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/vesting"
	"github.com/spf13/cobra"
)

// genesisCmd represents the wallet command.
var genesisCmd = &cobra.Command{
	Use:   "genesis",
	Short: "Genesis-related utilities",
}

// createCmd represents the create command.
var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify a genesis ledger account",
	Args:  cobra.MaximumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		var err error

		// first, collect the keys
		var keys []core.PublicKey
		fmt.Print("First, let's collect your public keys. Keys must be entered in hex format: 64 characters, without 0x prefix.\n")
		fmt.Print("Enter pubkeys one at a time; press enter again when done: ")
		for {
			var keyStr string
			_, err := fmt.Scanln(&keyStr)
			if err != nil {
				break
			}
			keyBytes, err := hex.DecodeString(keyStr)
			if err != nil || len(keyBytes) != ed25519.PublicKeySize {
				log.Fatalln("Error: key is unreadable")
			}
			key := [ed25519.PublicKeySize]byte{}
			copy(key[:], keyBytes)
			keys = append(keys, key)
			fmt.Printf("[enter next key or just press enter to end] > ")
		}
		if len(keys) == 0 {
			log.Fatalln("Error: must enter at least one key")
		}

		// next collect multisig params
		m := uint8(1)
		if len(keys) > 1 {
			fmt.Printf("Enter number of required signatures (between 1 and %d): ", len(keys))
			_, err := fmt.Scanln(&m)
			cobra.CheckErr(err)
		}

		// finally, collect amount
		var amount uint64
		fmt.Printf("Enter vault balance (denominated in SMH): ")
		_, err = fmt.Scanln(&amount)
		cobra.CheckErr(err)
		amount *= constants.OneSmesh

		// calculate keys
		vestingArgs := &multisig.SpawnArguments{
			Required:   m,
			PublicKeys: keys,
		}
		if int(vestingArgs.Required) > len(vestingArgs.PublicKeys) {
			log.Fatalf("requires more signatures (%d) than public keys (%d) in the wallet\n",
				vestingArgs.Required,
				len(vestingArgs.PublicKeys),
			)
		}
		vestingAddress := core.ComputePrincipal(vesting.TemplateAddress, vestingArgs)
		vaultArgs := &vault.SpawnArguments{
			Owner:               vestingAddress,
			TotalAmount:         amount,
			InitialUnlockAmount: amount / 4,
			VestingStart:        types.LayerID(constants.VestStart),
			VestingEnd:          types.LayerID(constants.VestEnd),
		}
		vaultAddress := core.ComputePrincipal(vault.TemplateAddress, vaultArgs)

		// output addresses
		fmt.Printf("Vesting address: %s\nVault address: %s\n",
			vestingAddress.String(),
			vaultAddress.String())
	},
}

func init() {
	rootCmd.AddCommand(genesisCmd)
	genesisCmd.AddCommand(verifyCmd)
}
