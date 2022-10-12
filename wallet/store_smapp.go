package wallet

// TODO: add a compatability implementation for the current smapp wallet store

type JSONSMAPPDecryptedWalletCipherText struct {
	Mnemonic string `json:"mnemonic"` // required
	Accounts []struct {
		DisplayName string `json:"displayName"` // optional
		Created     string `json:"created"`     // optional
		Path        string `json:"path"`        // required
		PublicKey   string `json:"publicKey"`   // optional
		SecretKey   string `json:" secretKey"`  // optional
	} `json:"accounts"` // optional
	Addresses []struct {
		Path    string `json:"path"`    // required
		Address string `json:"address"` // required
	} `json:"addresses"` // optional
}

type JSONSMAPPWalletMetaData struct {
	DisplayName string `json:"displayName"`
	Created     string `json:"created"`
	NetID       string `json:"netId"`
	RemoteAPI   string `json:"remoteApi"`
	HdID        string `json:"hd_id"`
	Meta        struct {
		Salt string `json:"salt"`
	} `json:"meta"`
}
type JSONSMAPPWalletCryptoData struct {
	Cipher     string `json:"cipher"`
	CipherText string `json:"cipherText"`
}
type JSONSMAPPWalletContactData struct {
	Nickname string `json:"nickname"`
	Address  string `json:"address"`
}
