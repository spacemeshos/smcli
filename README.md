# smcli: Spacemesh Command-line Interface Tool

smcli is a simple tool that you can use to manage wallet files and a running Spacemesh node. It currently supports the following features:

## Wallet

smcli allows you to read encrypted wallet files (including those created using Smapp and other compatible tools), and generate new wallet files.

### Reading

To read an encrypted wallet file, run `smcli wallet read <filename>`. You'll be prompted to enter the password used to encrypt the wallet file. If you enter the correct password, you'll see the contents of the wallet printed, including the mnemonic and any accounts it contains.

### Generation

To generate a new wallet, run `smcli wallet create`. You'll be prompted to enter a [BIP39-compatible](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki) 12- or 24-word mnemonic. You can also opt to generate a new, random mnemonic. By default the wallet file will contain a single account (one derived keypair).

**NOTE: We strongly recommend only creating a new wallet on a secure, airgapped computer. You are responsible for safely storing your generated mnemonic and wallet files. There is absolutely nothing that we can do to help you recover your wallet if you misplace the file or lose your mnemonic.**
