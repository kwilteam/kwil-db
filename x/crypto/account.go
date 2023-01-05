package crypto

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	ec "github.com/ethereum/go-ethereum/crypto"
)

// TODO: change getter methods (not idiomatic)
type Account interface {
	Sign(data []byte) (string, error)
	GetAddress() *common.Address
	GetPrivateKey() (*ecdsa.PrivateKey, error)
}

type account struct {
	Name    string         `json:"name" mapstructure:"name"`
	Address common.Address `json:"address" mapstructure:"address"`
	kr      KR             //keyring
}

type KR interface {
	Get(n string) ([]byte, error)
	Set(n string, pkb []byte) error
	GetAccount(n string) (*account, error)
}

func (k *keyring) GetAccount(n string) (*account, error) {
	key, err := k.Get(n)
	if err != nil {
		return nil, err
	}

	// convert to private key
	pk, err := ec.HexToECDSA(string(key))
	if err != nil {
		return nil, err
	}

	// Now we need to get the address from the private key
	addr := ec.PubkeyToAddress(pk.PublicKey)
	return &account{
		Name:    n,
		Address: addr,
		kr:      k,
	}, nil
}

func (k *keyring) GetDefaultAccount() (*account, error) {
	return k.GetAccount("kwil_main")
}

func (a *account) Sign(data []byte) (Signature, error) {
	pKey, err := a.GetPrivateKey()
	if err != nil {
		return Signature{}, err
	}

	sig, err := Sign(data, pKey)
	if err != nil {
		return Signature{}, err
	}

	// overwrite pk
	*pKey = ecdsa.PrivateKey{}

	return sig, nil
}

func (a *account) GetAddress() *common.Address {
	return &a.Address
}

func (a *account) GetPrivateKey() (*ecdsa.PrivateKey, error) {
	pkb, err := a.getPrivateKeyBytes()
	if err != nil {
		return nil, err
	}

	pkStr := string(pkb)

	// convert to private key
	pk, err := ec.HexToECDSA(pkStr)
	if err != nil {
		return nil, err
	}

	return pk, nil
}

func (a *account) getPrivateKeyBytes() ([]byte, error) {
	pkb, err := a.kr.Get(a.Name)
	if err != nil {
		return nil, err
	}

	return pkb, nil
}
