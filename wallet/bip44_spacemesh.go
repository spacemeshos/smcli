package wallet

import (
	"fmt"
	"regexp"
)

// Root of the path is m/purpose' (m/44')
// https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki#purpose
// We use this to indicate that we're using the BIP44 hierarchy
func BIP44Purpose() uint32 {
	return 0x8000002C
}

// After the purpose comes the coin type (m/44'/540')
// https://github.com/satoshilabs/slips/blob/master/slip-0044.md?plain=1#L571
func BIP44SpacemeshCoinType() uint32 {
	return 0x8000021c
}

// After the coin type comes the account (m/44'/540'/account')
func BIP44Account(account uint32) uint32 {
	return BIP32HardenedKeyStart | account
}

// After the account comes the change level, BUT as of now, we don't support
// un-hardened derivation so we'll be deviating from the spec here. We'll be
// continuing with hardened derivation.
// We call it the "hardened chain" level and not the "change" level because
// these chains don't support "change" functionality in the way that Bitcoin does.
// Spacemesh isn't a UTXO-based system, so there's no concept of "change".
// We're keeping the name "chain" because it's a more general term that can be
// used to describe a sequence of addresses that are related to each other even
// after the account level.
// (m/44'/540'/account'/chain')
func BIP44HardenedChain(chain uint32) uint32 {
	return BIP32HardenedKeyStart | chain
}

// After the Hardened Chain level comes the address indeces, as of now, we don't
// support un-hardened derivation so we'll continue our deviation from the spec
// here. All addresses will be hardened.
// (m/44'/540'/account'/chain'/address_index')
func BIP44HardenedAccountIndex(hai uint32) uint32 {
	return BIP32HardenedKeyStart | hai
}

func IsPathCompletelyHardened(path HDPath) bool {
	for _, p := range path {
		if p < BIP32HardenedKeyStart {
			return false
		}
	}
	return true
}

// HDPathToString converts a BIP44 HD path to a string of the form
// "m/44'/540'/account'/chain'/address_index'"
func HDPathToString(path HDPath) string {
	s := "m"
	for _, p := range path {
		if p > BIP32HardenedKeyStart {
			s += "/" + fmt.Sprint(p-BIP32HardenedKeyStart) + "'"
		} else {
			s += "/" + fmt.Sprint(p)
		}
	}
	return s
}

func parseUint(s string) uint {
	var u uint
	fmt.Sscanf(s, "%d", &u)
	return u
}

// StringToHDPath converts a BIP44 HD path string of the form
// (m/44'/540'/account'/chain'/address_index') to its uint32 slice representation
func StringToHDPath(s string) (HDPath, error) {
	// regex of the form m/44'/540'/account'/chain'/address_index'
	rWholePath := regexp.MustCompile(`^m(\/\d+'?)+$`)
	if !rWholePath.Match([]byte(s)) {
		return nil, fmt.Errorf("invalid HD path string: %s", s)
	}
	rCrumbs := regexp.MustCompile(`\/(\d+)('?)`)
	crumbs := rCrumbs.FindAllStringSubmatch(s, -1)
	path := make(HDPath, len(crumbs))
	for i, crumb := range crumbs {
		path[i] = uint32(parseUint(crumb[1]))
		if crumb[2] == "'" {
			path[i] |= BIP32HardenedKeyStart
		}
	}
	return path, nil
}
