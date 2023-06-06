package hd_wallet

import (
	"testing"
)

func TestEth_CreatePrivateKey(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		privKey  string
	}{{
		name:     "Test generate private key from mnemonic",
		mnemonic: "long purity dismiss tank nature cake diesel soup slim regret stomach affair",
		privKey:  "5ffb57b8dd4b8a04fbece681f0089045c616cd16113ea5865b960b66787acfbe",
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wallet, err := NewFromMnemonic(test.mnemonic)
			if err != nil {
				t.Fatalf("expect no error, but got error %v", err)
			}
			accountKey := wallet.GetMasterKey()
			if accountKey.PrivateKey != test.privKey {
				t.Fatalf("expect to get private key \"%v\", but got \"%v\"", accountKey.PrivateKey, test.privKey)
			}
		})
	}
}

func TestEth_DeriveAccount(t *testing.T) {
	tests := []struct {
		name     string
		mnemonic string
		hdpath   string
		address  string
		err      error
	}{{
		name:     "Test derive private key 1st index",
		mnemonic: "long purity dismiss tank nature cake diesel soup slim regret stomach affair",
		hdpath:   "m/44'/60'/0'/0/1",
		address:  "0xc8dA3A52f2FEb13812FEcc7b084Aa4F2Cc6e5392",
		err:      nil,
	}, {
		name:     "Test derive private key 2nd index",
		mnemonic: "long purity dismiss tank nature cake diesel soup slim regret stomach affair",
		hdpath:   "m/44'/60'/0'/0/2",
		address:  "0x3Fd00ddc8068888a255Eb2c6cCDE71D6cC36c5BF",
		err:      nil,
	}, {
		name:     "Test derive private key with invalid derivation path",
		mnemonic: "long purity dismiss tank nature cake diesel soup slim regret stomach affair",
		hdpath:   "invalid path",
		address:  "",
		err:      ErrHdWalletChildCreate,
	}}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			wallet, err := NewFromMnemonic(test.mnemonic)
			if err != nil {
				t.Fatalf("expect no error, but got error %v", err)
			}
			accountKey, err := wallet.DeriveFromPath(test.hdpath)
			if test.err != err {
				t.Fatalf("expect to get error %v, but got %v", test.err, err)
			}
			if test.err == nil && accountKey.Address != test.address {
				t.Fatalf("expect to get private key \"%v\", but got \"%v\"", accountKey.Address, test.address)
			}
		})
	}
}
