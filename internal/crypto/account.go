package crypto

import (
	"crypto/ecdsa"
	"encoding/json"
	"github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
	Name    string `json:"name" mapstructure:"name"`
	Address string `json:"address" mapstructure:"address"`
	kr      KR
}

type KR interface {
	Get(n string) ([]byte, error)
	Set(n string, pkb []byte) error
	GetAccount(n string) (*Account, error)
}

func (k *Keyring) GetAccount(n string) (*Account, error) {
	key, err := k.Get(n)
	if err != nil {
		return nil, err
	}

	// convert to private key
	pk, err := crypto.HexToECDSA(string(key))
	if err != nil {
		return nil, err
	}

	// Now we need to get the address from the private key
	addr := crypto.PubkeyToAddress(pk.PublicKey)
	return &Account{
		Name:    n,
		Address: addr.Hex(),
		kr:      k,
	}, nil
}

func (k *Keyring) GetDefaultAccount() (*Account, error) {
<<<<<<< HEAD
	return k.GetAccount(k.conf.GetKeyName())
=======
	return k.GetAccount(k.conf.Wallets.Ethereum.KeyName)
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
}

func (a *Account) Sign(data []byte) (string, error) {
	pKey, err := a.getPrivateKey()
	if err != nil {
		return "", err
	}

	sig, err := pKey.sign(data)
	if err != nil {
		return "", err
	}

	// overwrite pk
	*pKey.key = ecdsa.PrivateKey{}

	return sig, nil
}

func (a *Account) GetAddress() string {
	return a.Address
}

func (a *Account) getPrivateKey() (*PrivateKey, error) {
	pkb, err := a.getPrivateKeyBytes()
	if err != nil {
		return nil, err
	}

	var pk ecdsa.PrivateKey // this should be overwritten at the end of the function for security
	err = json.Unmarshal(pkb, &pk)
	if err != nil {
		return nil, err
	}

	return &PrivateKey{
		key: &pk,
	}, nil
}

func (a *Account) getPrivateKeyBytes() ([]byte, error) {
	pkb, err := a.kr.Get(a.Name)
	if err != nil {
		return nil, err
	}

	return pkb, nil
}
