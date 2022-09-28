package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
	kconf "kwil/x/chain/config/test"
	"kwil/x/chain/crypto"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Accepting all requests
	},
}

type MockAccount struct {
}

func (m *MockAccount) GetAddress() string {
	return "0x995d95245698212D4Af52c8031F614C3D3127994"
}

func (m *MockAccount) Sign(msg []byte) (string, error) {
	pk, err := crypto.ECDSAFromHex("4bb214b1f3a0737d758bc3828cdff371e3769fe84a2678da34700cb18d50770e")
	if err != nil {
		return "", err
	}
	return crypto.Sign(msg, pk)
}

var pa = func(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	tc := kconf.GetTestConfig()

	ma := newAuthenticator(tc)

	err = ma.Authenticate(conn)
	if err != nil {
		panic(err)
	}
}

func TestAuthClient_RequestAuth(t *testing.T) {

	s := httptest.NewServer(http.HandlerFunc(pa))
	defer s.Close()

	u := strings.TrimPrefix(s.URL, "http://")

	type fields struct {
		keys map[string]string
		acc  account
		log  zerolog.Logger
	}
	type args struct {
		ip string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				keys: make(map[string]string),
				acc:  &MockAccount{},
				log:  zerolog.Nop(),
			},
			args: args{
				ip: u,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ac := &authClient{
				keys: tt.fields.keys,
				acc:  tt.fields.acc,
				log:  tt.fields.log,
			}
			got, err := ac.RequestAuth(tt.args.ip)
			if (err != nil) != tt.wantErr {
				t.Errorf("authClient.RequestAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("authClient.RequestAuth() = %v, want %v", got, tt.want)
			}
		})
	}
}
