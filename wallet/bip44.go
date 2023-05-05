package wallet

import (
	"encoding/json"
	"fmt"
	"regexp"
)

//lint:file-ignore SA4016 ignore ineffective bitwise operations to aid readability

// BIP32HardenedKeyStart: keys with index >= this must be hardened as per BIP32.
// https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki#extended-keys
const BIP32HardenedKeyStart uint32 = 0x80000000

const (
	HDPurposeSegment  = 0
	HDCoinTypeSegment = 1
	HDAccountSegment  = 2
	HDChainSegment    = 3
	HDIndexSegment    = 4
)

type HDPath []uint32

func (p *HDPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *HDPath) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return
	}
	*p, err = StringToHDPath(s)
	return
}

func (p *HDPath) String() string {
	return HDPathToString(*p)
}

func (p *HDPath) Purpose() uint32 {
	return (*p)[HDPurposeSegment]
}

func (p *HDPath) CoinType() uint32 {
	return (*p)[HDCoinTypeSegment]
}

func (p *HDPath) Account() uint32 {
	return (*p)[HDAccountSegment]
}

func (p *HDPath) Chain() uint32 {
	return (*p)[HDChainSegment]
}

func (p *HDPath) Index() uint32 {
	return (*p)[HDIndexSegment]
}

func (p *HDPath) Extend(idx uint32) HDPath {
	return append(*p, idx)
}

// Root of the path is m/purpose' (m/44')
// https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki#purpose
// We use this to indicate that we're using the BIP44 hierarchy.
func BIP44Purpose() uint32 {
	return 0x8000002C
}

// After the purpose comes the coin type (m/44'/540')
// https://github.com/satoshilabs/slips/blob/master/slip-0044.md?plain=1#L571
func BIP44SpacemeshCoinType() uint32 {
	return 0x8000021c
}

// After the coin type comes the account (m/44'/540'/account')
// For now we only support account 0'.
func BIP44Account() uint32 {
	//nolint:staticcheck // ignore ineffective bitwise operations to aid readability
	return BIP32HardenedKeyStart | 0
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
// For now we only support "chain" 0. We may want to use a different chain for testnet.
func BIP44HardenedChain() uint32 {
	//nolint:staticcheck // ignore ineffective bitwise operations to aid readability
	return BIP32HardenedKeyStart | 0
}

// After the Hardened Chain level comes the address indices, as of now, we don't
// support un-hardened derivation so we'll continue our deviation from the spec
// here. All addresses will be hardened.
// (m/44'/540'/account'/chain'/address_index').
func BIP44HardenedAccountIndex(hai uint32) uint32 {
	return BIP32HardenedKeyStart | hai
}

func DefaultPath() HDPath {
	return HDPath{
		BIP44Purpose(),
		BIP44SpacemeshCoinType(),
		BIP44Account(),
		BIP44HardenedChain(),
	}
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
// "m/44'/540'/account'/chain'/address_index'".
func HDPathToString(path HDPath) string {
	s := "m"
	for _, p := range path {
		if p >= BIP32HardenedKeyStart {
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
// (m/44'/540'/account'/chain'/address_index') to its uint32 slice representation.
func StringToHDPath(s string) (HDPath, error) {
	// regex of the form m/44'/540'/account'/chain'/address_index'
	rWholePath := regexp.MustCompile(`^m(/\d+'?)+$`)
	if !rWholePath.Match([]byte(s)) {
		return nil, fmt.Errorf("invalid HD path string: %s", s)
	}
	rCrumbs := regexp.MustCompile(`/(\d+)('?)`)
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
