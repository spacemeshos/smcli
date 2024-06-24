package main

import (
	"crypto/ed25519"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spacemeshos/go-spacemesh/common/types"
	"github.com/spacemeshos/go-spacemesh/genvm/core"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/multisig"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/vault"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/vesting"
	"github.com/spacemeshos/go-spacemesh/genvm/templates/wallet"
	"log"
	"math/rand"
	"os"
)

const MaxPubkeys = 3

func getKey() (key core.PublicKey) {
	// generate a random pubkey and discard the private key
	pub, _, err := ed25519.GenerateKey(rand.New(rand.NewSource(rand.Int63())))
	if err != nil {
		log.Fatal("failed to generate ed25519 key")
	}
	copy(key[:], pub)
	return
}

func main() {
	t1 := table.NewWriter()
	t1.SetOutputMirror(os.Stdout)
	t1.SetTitle("address test vectors (wallet, multisig, vesting)")

	t2 := table.NewWriter()
	t2.SetOutputMirror(os.Stdout)
	t2.SetTitle("address test vectors (vault)")

	t1Rows := table.Row{
		"address",
		"network",
		"template",
		"m",
		"n",
	}

	// pregenerate keys
	keys := make([]core.PublicKey, MaxPubkeys)
	for i := range keys {
		keys[i] = getKey()
		t1Rows = append(t1Rows, fmt.Sprintf("pubkey[%d]", i))
	}
	t1.AppendHeader(t1Rows)
	t2.AppendHeader(table.Row{
		"address",
		"network",
		"template",
		"total",
		"start",
		"end",
		"owner",
	})
	runNetwork("sm", keys, t1, t2)
	runNetwork("stest", keys, t1, t2)

	t1.Render()
	t2.Render()
}

func runNetwork(hrp string, keys []core.PublicKey, t1, t2 table.Writer) {
	types.SetNetworkHRP(hrp)

	// single sig
	walletArgs := &wallet.SpawnArguments{PublicKey: keys[0]}
	walletAddress := core.ComputePrincipal(wallet.TemplateAddress, walletArgs)
	t1.AppendRow(table.Row{
		walletAddress.String(),
		hrp,
		"wallet",
		"1",
		"1",
		keys[0].String(),
	})

	// multisig and vesting wallet
	// m-of-n for n from 1 to max and m from 1 to n
	vestingAddresses := make([]types.Address, MaxPubkeys)
	for n := uint8(1); n <= MaxPubkeys; n++ {
		for m := uint8(1); m <= n; m++ {
			multisigArgs := &multisig.SpawnArguments{
				Required:   m,
				PublicKeys: keys[:n],
			}

			// multisig
			address := core.ComputePrincipal(multisig.TemplateAddress, multisigArgs)
			row := []interface{}{
				address.String(),
				hrp,
				"multisig",
				m,
				n,
			}
			for _, k := range keys[:n] {
				row = append(row, k.String())
			}
			t1.AppendRow(row)

			// vesting
			address = core.ComputePrincipal(vesting.TemplateAddress, multisigArgs)
			row = []interface{}{
				address.String(),
				hrp,
				"multisig",
				m,
				n,
			}
			for _, k := range keys[:n] {
				row = append(row, k.String())
			}
			t1.AppendRow(row)

			// save some to compute vault addresses, doesn't matter which ones we save
			vestingAddresses[n-1] = address
		}
	}

	// vault
	for _, totalAmount := range []uint64{1, 10, 100, 1000} {
		for _, vestingStart := range []types.LayerID{1, 10, 100, 1000} {
			for _, vestingEnd := range []types.LayerID{1, 10, 100, 1000} {
				for _, owner := range vestingAddresses {
					vaultArgs := &vault.SpawnArguments{
						Owner:        owner,
						TotalAmount:  totalAmount,
						VestingStart: vestingStart,
						VestingEnd:   vestingEnd,
					}
					address := core.ComputePrincipal(vault.TemplateAddress, vaultArgs)
					t2.AppendRow(table.Row{
						address.String(),
						hrp,
						"vault",
						totalAmount,
						vestingStart,
						vestingEnd,
						owner.String(),
					})
				}
			}
		}
	}
}
