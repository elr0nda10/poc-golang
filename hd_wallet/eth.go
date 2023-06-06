package hd_wallet

import (
	"fmt"
	hdwallet "github.com/miguelmota/go-ethereum-hdwallet"
)

var (
	ErrHDWalletCreate      = fmt.Errorf("wallet creation failed")
	ErrHdWalletChildCreate = fmt.Errorf("wallet derivation failed")
)

const (
	HDEthRootPath = "m/44'/60'/0'/0/0"
)

type HDWalletEth struct {
	wallet    *hdwallet.Wallet
	masterKey *AccountKey
}

type AccountKey struct {
	PrivateKey string
	PublicKey  string
	Address    string
	HdPath     string
}

func getAccountKeyFromPath(wallet *hdwallet.Wallet, hdpath string) (res *AccountKey, resErr error) {
	defer func() {
		if r := recover(); r != nil {
			resErr = ErrHdWalletChildCreate
			res = nil
		}
	}()
	rootPath := hdwallet.MustParseDerivationPath(hdpath)
	rootAccount, err := wallet.Derive(rootPath, false)
	if err != nil {
		return nil, err
	}

	privateKey, _ := wallet.PrivateKeyHex(rootAccount)
	if err != nil {
		return nil, err
	}
	publicKey, _ := wallet.PublicKeyHex(rootAccount)
	if err != nil {
		return nil, err
	}
	address, _ := wallet.AddressHex(rootAccount)

	return &AccountKey{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
		HdPath:     hdpath,
	}, nil
}

func NewFromMnemonic(mnemonic string) (*HDWalletEth, error) {
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		return nil, ErrHDWalletCreate
	}

	accountKey, err := getAccountKeyFromPath(wallet, HDEthRootPath)
	if err != nil {
		return nil, ErrHDWalletCreate
	}

	return &HDWalletEth{
		wallet:    wallet,
		masterKey: accountKey,
	}, nil
}

func (hd *HDWalletEth) GetMasterKey() AccountKey {
	return *hd.masterKey
}

func (hd *HDWalletEth) DeriveFromPath(path string) (*AccountKey, error) {
	accountKey, err := getAccountKeyFromPath(hd.wallet, path)
	if err != nil {
		return nil, ErrHdWalletChildCreate
	}

	return accountKey, nil
}
