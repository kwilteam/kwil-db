package crypto_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/crypto"
)

func TestPrivateKeyFromHex(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		// fn is a function to create a private key from hex string
		fn      func(key string) error
		wantErr bool
	}{
		{
			name: "empty secp256k1 private key",
			args: args{
				key: "",
			},
			fn: func(key string) error {
				_, err := crypto.Secp256k1PrivateKeyFromHex(key)
				return err
			},
			wantErr: true,
		},
		{
			name: "secp256k1",
			args: args{
				key: "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e",
			},
			fn: func(key string) error {
				_, err := crypto.Secp256k1PrivateKeyFromHex(key)
				return err
			},
			wantErr: false,
		},
		{
			name: "secp256k1 invalid key",
			args: args{
				key: "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f",
			},
			fn: func(key string) error {
				_, err := crypto.Secp256k1PrivateKeyFromHex(key)
				return err
			},
			wantErr: true,
		},
		{
			name: "empty ed25519 private key",
			args: args{
				key: "",
			},
			fn: func(key string) error {
				_, err := crypto.Ed25519PrivateKeyFromHex(key)
				return err
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.args.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("PrivateKeyFromHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}
		})
	}
}
