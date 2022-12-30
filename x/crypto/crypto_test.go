package crypto

import (
	"encoding/hex"
	"reflect"
	"testing"
)

// Using private key 4bb214b1f3a0737d758bc3828cdff371e3769fe84a2678da34700cb18d50770e
// Public:
// Public Bytes: [4 197 141 51 158 16 36 14 57 147 17 68 175 224 209 17 1 128 241 107 124 249 138 4 140 195 17 175 251 164 131 87 37 187 20 25 78 94 105 159 107 221 221 213 105 170 169 248 255 206 112 253 139 14 195 102 158 15 104 246 110 146 154 137 171]
// Address: 0x995d95245698212D4Af52c8031F614C3D3127994

func TestSha384(t *testing.T) {
	res, err := hex.DecodeString("82835f0f3732e85736f1372184640199c9155a81980f562b4418aadabe2a21f57cb580b48f2f06b439bdf204f4b3dcb7")
	if err != nil {
		t.Errorf("Error with TestSha384")
	}

	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "test1",
			args: args{
				data: []byte("kwil"),
			},
			want: res,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sha384(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sha384() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSha384Str(t *testing.T) {

	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test1",
			args: args{
				data: []byte("kwil"),
			},
			want: "82835f0f3732e85736f1372184640199c9155a81980f562b4418aadabe2a21f57cb580b48f2f06b439bdf204f4b3dcb7",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Sha384Str(tt.args.data); got != tt.want {
				t.Errorf("Sha384Str() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Sha256(t *testing.T) {
	res := "0c9e4969977f81d845cb959915b563a09a857e8d16911ce5a2780f38a5985410"
	hash := Sha256Str([]byte("kwil"))
	if hash != res {
		t.Errorf("Error with Sha256")
	}
}

func Test_Sha224(t *testing.T) {
	res := "9651608f63f583b7684c3c856d3ef1e3c9d70e2c05b4fa0080989c4d"
	hash := Sha224Str([]byte("kwil"))
	if hash != res {
		t.Errorf("Error with Sha224")
	}
}
