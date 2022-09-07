package crypto

import (
	"github.com/99designs/keyring"
	"github.com/kwilteam/kwil-db/internal/utils/files"
<<<<<<< HEAD
=======
	"github.com/kwilteam/kwil-db/pkg/types"
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
)

type Keyring struct {
	kr   keyring.Keyring
<<<<<<< HEAD
	conf config
}

type config interface {
	GetPrivKeyPath() string
	GetKeyName() string
}

func NewKeyring(c config) (*Keyring, error) {
	kr, err := keyring.Open(keyring.Config{ServiceName: "kwil"})
=======
	conf *types.Config
}

func NewKeyring(c *types.Config) (*Keyring, error) {

	kr, err := keyring.Open(keyring.Config{FileDir: c.Wallets.KeyringFile})
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
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
<<<<<<< HEAD
	key, err := files.LoadFileFromRoot(k.conf.GetPrivKeyPath())
=======
	key, err := files.LoadFileFromRoot(k.conf.Wallets.Ethereum.PrivKeyPath)
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
	if err != nil {
		return err
	}

<<<<<<< HEAD
	err = k.Set(k.conf.GetKeyName(), key)
=======
	err = k.Set(k.conf.Wallets.Ethereum.KeyName, key)
>>>>>>> b64dc94cf02f1f9d814336627f167ff5d29bb7d5
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
