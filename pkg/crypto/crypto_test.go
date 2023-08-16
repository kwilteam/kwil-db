package crypto_test

import (
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"reflect"
	"testing"
)

func TestPrivateKeyFromHex(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    crypto.KeyType
		wantErr bool
	}{
		{
			name: "empty secp256k1 private key",
			args: args{
				key: "",
			},
			want:    crypto.Secp256k1,
			wantErr: true,
		},
		{
			name: "secp256k1",
			args: args{
				key: "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e",
			},
			want:    crypto.Secp256k1,
			wantErr: false,
		},
		{
			name: "secp256k1 invalid key",
			args: args{
				key: "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f",
			},
			want:    crypto.Secp256k1,
			wantErr: true,
		},
		{
			name: "empty ed25519 private key",
			args: args{
				key: "",
			},
			want:    crypto.Ed25519,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pk crypto.PrivateKey
			var err error
			switch tt.want {
			case crypto.Secp256k1:
				pk, err = crypto.Secp256k1PrivateKeyFromHex(tt.args.key)
			case crypto.Ed25519:
				pk, err = crypto.Ed25519PrivateKeyFromHex(tt.args.key)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("PrivateKeyFromHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if !reflect.DeepEqual(pk.Type(), tt.want) {
				t.Errorf("PrivateKeyFromHex() got = %v, want %v", pk.Type(), tt.want)
			}
		})
	}
}
