package wallet_test

import (
	"testing"

	"github.com/spacemeshos/smcli/wallet"
	"github.com/stretchr/testify/require"
)

func TestHDPathToString(t *testing.T) {
	s := wallet.HDPathToString(wallet.HDPath{wallet.BIP32HardenedKeyStart | 44, wallet.BIP32HardenedKeyStart | 540, 0, 0, 0})
	require.Equal(t, "m/44'/540'/0/0/0", s)
}

func TestStringToHDPath(t *testing.T) {
	testVectors := []struct {
		path     string
		expected wallet.HDPath
	}{
		{"m/44'/540'", wallet.HDPath{wallet.BIP32HardenedKeyStart | 44, wallet.BIP32HardenedKeyStart | 540}},
		{"m/44'/540'/0", wallet.HDPath{wallet.BIP32HardenedKeyStart | 44, wallet.BIP32HardenedKeyStart | 540, 0}},
		{"m/44'/540'/0'/0'/0'", wallet.HDPath{wallet.BIP32HardenedKeyStart | 44, wallet.BIP32HardenedKeyStart | 540, wallet.BIP32HardenedKeyStart, wallet.BIP32HardenedKeyStart, wallet.BIP32HardenedKeyStart}},
		{"m/44'/540'/0'/0/0", wallet.HDPath{wallet.BIP32HardenedKeyStart | 44, wallet.BIP32HardenedKeyStart | 540, wallet.BIP32HardenedKeyStart, 0, 0}},
		{"m/44'/540'/0/0'/0", wallet.HDPath{wallet.BIP32HardenedKeyStart | 44, wallet.BIP32HardenedKeyStart | 540, 0, wallet.BIP32HardenedKeyStart, 0}},
		{"m/44'/540'/2'/0/0", wallet.HDPath{wallet.BIP32HardenedKeyStart | 44, wallet.BIP32HardenedKeyStart | 540, wallet.BIP32HardenedKeyStart | 2, 0, 0}},
	}

	for _, tv := range testVectors {
		t.Run(tv.path, func(t *testing.T) {
			p, err := wallet.StringToHDPath(tv.path)
			require.NoError(t, err)
			require.Equal(t, tv.expected, p)
		})
	}
}
