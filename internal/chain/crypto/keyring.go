package crypto

import (
	"github.com/99designs/keyring"
	"github.com/kwilteam/kwil-db/internal/chain/utils"
)

type Keyring struct {
	kr   keyring.Keyring
	conf config
}

type config interface {
	GetPrivKeyPath() string
	GetKeyName() string
}

func NewKeyring(c config) (*Keyring, error) {
	kr, err := keyring.Open(getKeyRingConfig("kwil"))
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
	key, err := utils.LoadFileFromRoot(k.conf.GetPrivKeyPath())
	if err != nil {
		return err
	}

	err = k.Set(k.conf.GetKeyName(), key)
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

func getKeyRingConfig(serviceName string) keyring.Config {
	return keyring.Config{ServiceName: "kwil", FileDir: "~",
		FilePasswordFunc: func(prompt string) (string, error) {
			return "test", nil
		}}
}
