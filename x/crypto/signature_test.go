package crypto

import (
	"crypto/ecdsa"
	"testing"

	ec "github.com/ethereum/go-ethereum/crypto"
)

func TestSign(t *testing.T) {
	pk, err := ec.HexToECDSA("4bb214b1f3a0737d758bc3828cdff371e3769fe84a2678da34700cb18d50770e")
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		data []byte
		k    *ecdsa.PrivateKey
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid_sig",
			args: args{
				data: []byte("kwil"),
				k:    pk,
			},
			want:    "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sign(tt.args.data, tt.args.k)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Sign() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckSignature(t *testing.T) {
	type args struct {
		addr string
		sig  string
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid_sig",
			args: args{
				addr: "0x995d95245698212D4Af52c8031F614C3D3127994",
				sig:  "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
				data: []byte("kwil"),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "invalid_msg",
			args: args{
				addr: "0x995d95245698212D4Af52c8031F614C3D3127994",
				sig:  "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
				data: []byte("kwill"),
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "empty_sig",
			args: args{
				addr: "0x995d95245698212D4Af52c8031F614C3D3127994",
				sig:  "",
				data: []byte("kwil"),
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "too_long_sig",
			args: args{
				addr: "0x995d95245698212D4Af52c8031F614C3D3127994",
				sig:  "0x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b425010x39fd0a5551cd0008eb45244ad3eea11fb960ff6d8d13aaad9651632b61d26ee20da867cf4f53564bc7bfa795d1efb2bb1169209d1e6f42a2d9e88cfce556b42501",
				data: []byte("kwil"),
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckSignature(tt.args.addr, tt.args.sig, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckSignature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}
