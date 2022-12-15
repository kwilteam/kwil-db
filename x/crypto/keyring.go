package crypto

import (
	"runtime"

	"kwil/x/cfgx"
	"kwil/x/utils"

	kr "github.com/99designs/keyring"
)

type Keyring interface {
	Get(string) ([]byte, error)
	Set(string, []byte) error
	GetAccount(string) (*account, error)
}

type keyring struct {
	kr   kr.Keyring
	conf cfgx.Config
}

func NewKeyring(c cfgx.Config) (*keyring, error) {
	kr, err := kr.Open(getKeyRingConfig("kwil"))
	if err != nil {
		return nil, err
	}

	nkr := &keyring{
		kr:   kr,
		conf: c,
	}

	err = nkr.importConfigKey()
	if err != nil {
		return nkr, err
	}

	return nkr, nil
}

func (k *keyring) importConfigKey() (err error) {
	key := []byte(k.conf.String("keys.wallet-key"))
	if len(key) > 0 {
		return k.Set("kwil_main", key)
	}

	key, err = utils.LoadFileFromRoot(k.conf.String("keys.key-path"))
	if err != nil {
		return err
	}

	return k.Set("kwil_main", key)
}

func (k *keyring) Set(name string, key []byte) error {
	return k.kr.Set(kr.Item{
		Key:  name,
		Data: key,
	})
}

func (k *keyring) Get(name string) ([]byte, error) {
	item, err := k.kr.Get(name)
	return item.Data, err
}

func getKeyRingConfig(serviceName string) kr.Config {
	if runtime.GOOS == "darwin" {
		return kr.Config{ServiceName: serviceName, KeychainName: "kwil",
			KeychainTrustApplication: true}
	}
	return kr.Config{ServiceName: serviceName, FileDir: "~",
		FilePasswordFunc: func(prompt string) (string, error) {
			return "test", nil
		}}
}
