package wallet

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHDPathToString(t *testing.T) {
	s := HDPathToString(HDPath{BIP32HardenedKeyStart | 44, BIP32HardenedKeyStart | 540, 0, 0, 0})
	require.Equal(t, "m/44'/540'/0/0/0", s)
}

func TestStringToHDPath(t *testing.T) {
	testVectors := []struct {
		path     string
		expected HDPath
	}{
		{
			"m/44'/540'",
			HDPath{BIP32HardenedKeyStart | 44, BIP32HardenedKeyStart | 540},
		},
		{
			"m/44'/540'/0",
			HDPath{BIP32HardenedKeyStart | 44, BIP32HardenedKeyStart | 540, 0},
		},
		{
			"m/44'/540'/0'/0'/0'",
			HDPath{
				BIP32HardenedKeyStart | 44, BIP32HardenedKeyStart | 540, BIP32HardenedKeyStart,
				BIP32HardenedKeyStart, BIP32HardenedKeyStart,
			},
		},
		{
			"m/44'/540'/0'/0/0",
			HDPath{BIP32HardenedKeyStart | 44, BIP32HardenedKeyStart | 540, BIP32HardenedKeyStart, 0, 0},
		},
		{
			"m/44'/540'/0/0'/0",
			HDPath{BIP32HardenedKeyStart | 44, BIP32HardenedKeyStart | 540, 0, BIP32HardenedKeyStart, 0},
		},
		{
			"m/44'/540'/2'/0/0",
			HDPath{BIP32HardenedKeyStart | 44, BIP32HardenedKeyStart | 540, BIP32HardenedKeyStart | 2, 0, 0},
		},
	}

	for _, tv := range testVectors {
		t.Run(tv.path, func(t *testing.T) {
			p, err := StringToHDPath(tv.path)
			require.NoError(t, err)
			require.Equal(t, tv.expected, p)
		})
	}
}
