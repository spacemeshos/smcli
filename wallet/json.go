package wallet

type JSONWallet struct {
	Mnemonic string `json:"mnemonic"`
	Accounts []struct {
		DisplayName string `json:"displayName"`
		Created     string `json:"created"`
		Path        string `json:"path"`
		PublicKey   string `json:"publicKey"`
		SecretKey   string `json:" secretKey"`
	} `json:"accounts"`
}

type JSONWalletMetaData struct {
	DisplayName string `json:"displayName"`
	Created     string `json:"created"`
	NetID       string `json:"netId"`
	RemoteAPI   string `json:"remoteApi"`
	HdID        string `json:"hd_id"`
	Meta        struct {
		Salt string `json:"salt"`
	} `json:"meta"`
}
type JSONWalletCryptoData struct {
	Cipher     string `json:"cipher"`
	CipherText string `json:"cipherText"`
}
type JSONWalletContactData struct {
	Nickname string `json:"nickname"`
	Address  string `json:"address"`
}
