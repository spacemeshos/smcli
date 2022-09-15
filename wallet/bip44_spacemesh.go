package wallet

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
	return 0x80000000 | account
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
	return 0x80000000 | chain
}

// After the Hardened Chain level comes the address indeces, as of now, we don't
// support un-hardened derivation so we'll continue our deviation from the spec
// here. All addresses will be hardened.
// (m/44'/540'/account'/chain'/address_index')
func BIP44HardenedAccountIndex(hai uint32) uint32 {
	return 0x80000000 | hai
}
