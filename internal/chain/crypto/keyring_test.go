package crypto

import (
	"testing"

	"github.com/99designs/keyring"
	kconf "github.com/kwilteam/kwil-db/internal/chain/config/test"
)

type MockKeyRing struct {
}

func (k *MockKeyRing) Get(n string) ([]byte, error) {
	return []byte("4bb214b1f3a0737d758bc3828cdff371e3769fe84a2678da34700cb18d50770e"), nil
}

func (k *MockKeyRing) Set(n string, p []byte) error {
	return nil
}

func (k *MockKeyRing) GetAccount(n string) (*Account, error) {
	return &Account{
		Name:    "brennan",
		Address: "0x9f8f72a0007c9c62c1fd76f972b9d5d7a9c0dbf9",
		kr:      k,
	}, nil
}

func TestNewKeyring(t *testing.T) {
	type args struct {
		c config
	}
	tests := []struct {
		name    string
		args    args
		want    *Keyring
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				c: kconf.GetTestConfig(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewKeyring(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestKeyring_ImportConfigKey(t *testing.T) {
	kr, err := keyring.Open(getKeyRingConfig("kwil"))

	if err != nil {
		t.Error(err)
	}

	type fields struct {
		kr   keyring.Keyring
		conf config
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				kr:   kr,
				conf: kconf.GetTestConfig(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Keyring{
				kr:   tt.fields.kr,
				conf: tt.fields.conf,
			}
			if err := k.importConfigKey(); (err != nil) != tt.wantErr {
				t.Errorf("Keyring.ImportConfigKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
