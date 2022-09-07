package crypto

import (
	"github.com/99designs/keyring"
	types "github.com/kwilteam/kwil-db/pkg/types/chain"
)

type Keyring struct {
	kr   keyring.Keyring
	conf *types.Config
}

func NewKeyring(c *types.Config) (*Keyring, error) {

	kr, err := keyring.Open(keyring.Config{FileDir: c.Wallets.KeyringFile})
	if err != nil {
		return nil, err
	}

	nkr := &Keyring{
		kr:   kr,
		conf: c,
	}

	err = nkr.importConfigKey()
	if err != nil {
		return nkr, err
	}

	return nkr, nil
}

func (k *Keyring) importConfigKey() error {
	key, err := loadFileFromRoot(k.conf.Wallets.Ethereum.PrivKeyPath)
	if err != nil {
		return err
	}

	err = k.Set(k.conf.Wallets.Ethereum.KeyName, key)
	if err != nil {
		return err
	}

	return nil
}

func (k *Keyring) Set(name string, key []byte) error {
	return k.kr.Set(keyring.Item{
		Key:  name,
		Data: key,
	})
}

func (k *Keyring) Get(name string) ([]byte, error) {
	item, err := k.kr.Get(name)
	return item.Data, err
}
