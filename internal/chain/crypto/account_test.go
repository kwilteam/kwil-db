package crypto

import (
	"reflect"
	"testing"

	types "github.com/kwilteam/kwil-db/pkg/types/chain"

	kconf "github.com/kwilteam/kwil-db/internal/chain/config/test"
)

func TestKeyring_GetAccount(t *testing.T) {
	k := MockKeyRing{}
	type fields struct {
		kr   MockKeyRing
		conf *types.Config
	}
	type args struct {
		n string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Account
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				kr:   MockKeyRing{},
				conf: kconf.GetTestConfig(),
			},
			args: args{
				n: "brennan",
			},
			want: &Account{
				Name:    "brennan",
				Address: "0x9f8f72a0007c9c62c1fd76f972b9d5d7a9c0dbf9",
				kr:      &k,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := MockKeyRing{}
			got, err := k.GetAccount(tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("Keyring.GetAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Keyring.GetAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}
