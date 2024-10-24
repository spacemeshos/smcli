package cmd

import (
	"crypto/ed25519"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"

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
		fmt.Print("First, let's collect your public keys. ")
		fmt.Print("Keys must be entered in hex format: 64 characters, without 0x prefix.\n")
		fmt.Print("Enter pub keys one at a time; press enter again when done: ")
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
		requiredSignatures := uint8(1)
		if len(keys) > 1 {
			fmt.Printf("Enter number of required signatures (between 1 and %d): ", len(keys))
			_, err := fmt.Scanln(&requiredSignatures)
			cobra.CheckErr(err)
		}

		// finally, collect amount
		var amount uint64
		fmt.Printf("Enter vault balance (denominated in SMH): ")
		_, err = fmt.Scanln(&amount)
		cobra.CheckErr(err)
		amount *= constants.OneSmesh

		vestingAddress, vaultAddress := generateAddresses(requiredSignatures, keys, amount)

		// output addresses
		fmt.Printf("Vesting address: %s\nVault address: %s\n",
			vestingAddress.String(),
			vaultAddress.String())
	},
}

var processCSVCmd = &cobra.Command{
	Use: "csv",
	Short: "Process a CSV file with account definitions " +
		"(name, amount, key1, key2, key3, key4, key5, required_signatures)",
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		csvFilePath, _ := cmd.Flags().GetString("file")
		if csvFilePath == "" {
			fmt.Println("CSV file path is required")
			return
		}

		file, err := os.Open(csvFilePath)
		if err != nil {
			fmt.Printf("Failed to open CSV file: %v\n", err)
			return
		}
		defer file.Close()

		reader := csv.NewReader(file)
		records, err := reader.ReadAll()
		if err != nil {
			fmt.Printf("Failed to read CSV file: %v\n", err)
			return
		}

		// Process each record
		for i, record := range records {
			if i == 0 {
				// Add new columns to the header
				record = append(record, "vesting_address", "vault_address")
				records[i] = record
				continue
			}

			// Extract account details from the record
			name := record[0]
			amountStr := record[1]
			keysStr := record[2:7]
			requiredSignaturesStr := record[7]

			// Convert amount and requiredSignatures to appropriate types
			amount, err := strconv.ParseUint(amountStr, 10, 64)
			if err != nil {
				log.Fatalf("Error: amount %s for %s on line %d is invalid", amountStr, name, i)
			}
			amount *= constants.OneSmesh

			requiredSignatures64, err := strconv.ParseUint(requiredSignaturesStr, 10, 8)
			if err != nil {
				log.Fatalf("Error: required signatures %s for %s on line %d is invalid", requiredSignaturesStr, name, i)
			}
			requiredSignatures := uint8(requiredSignatures64)

			// Collect the keys
			var publicKeys []core.PublicKey
			for _, keyStr := range keysStr {
				if len(keyStr) == 0 {
					break
				}
				keyBytes, err := hex.DecodeString(keyStr)
				if err != nil || len(keyBytes) != ed25519.PublicKeySize {
					log.Fatalf("Error: key %s for  %s on line %d is unreadable", keyStr, name, i)
				}
				var pubKey core.PublicKey
				copy(pubKey[:], keyBytes)
				publicKeys = append(publicKeys, pubKey)
			}

			// Generate vesting and vault addresses
			vestingAddress, vaultAddress := generateAddresses(requiredSignatures, publicKeys, amount)

			// Append new addresses to the record
			record = append(record, vestingAddress.String(), vaultAddress.String())
			records[i] = record
		}

		// Write updated records to a new CSV file
		outputFile, err := os.Create("updated_" + csvFilePath)
		if err != nil {
			fmt.Printf("Failed to create output CSV file: %v\n", err)
			return
		}
		defer outputFile.Close()

		writer := csv.NewWriter(outputFile)
		err = writer.WriteAll(records)
		if err != nil {
			fmt.Printf("Failed to write to output CSV file: %v\n", err)
			return
		}

		fmt.Println("CSV file processed successfully")
	},
}

// generateAddresses generates vesting and vault addresses based on the provided parameters.
func generateAddresses(
	requiredSignatures uint8,
	publicKeys []core.PublicKey,
	amount uint64,
) (core.Address, core.Address) {
	vestingArgs := &multisig.SpawnArguments{
		Required:   requiredSignatures,
		PublicKeys: publicKeys,
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

	return vestingAddress, vaultAddress
}

func init() {
	rootCmd.AddCommand(genesisCmd)
	genesisCmd.AddCommand(verifyCmd)
	processCSVCmd.Flags().StringP("file", "f", "",
		"Path to a CSV file with headings: name, amount, key1, key2, key3, key4, key5, required_signatures")
	genesisCmd.AddCommand(processCSVCmd)
}
