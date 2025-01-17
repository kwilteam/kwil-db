package auth

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/core/crypto"
)

func Test_eip55ChecksumAddr(t *testing.T) {
	tests := []struct {
		name     string
		input    [20]byte
		expected string
	}{
		{
			name:     "Basic address",
			input:    [20]byte{0x5a, 0xAA, 0xfE, 0x6F, 0x8E, 0x4E, 0x44, 0xAA, 0x5d, 0x4c, 0xBd, 0x08, 0x7A, 0x63, 0x9B, 0x5E, 0x8A, 0x3E, 0xd3, 0x95},
			expected: "0x5aaaFe6F8e4E44aa5D4cBd087a639b5e8a3Ed395",
		},
		{
			name:     "All zeros",
			input:    [20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			expected: "0x0000000000000000000000000000000000000000",
		},
		{
			name:     "Mixed case address",
			input:    [20]byte{0x00, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0, 0x12, 0x34, 0x56},
			expected: "0x00123456789AbcdeF0123456789abCdef0123456",
		},
		{
			name:     "All F's",
			input:    [20]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			expected: "0xFFfFfFffFFfffFFfFFfFFFFFffFFFffffFfFFFfF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eip55ChecksumAddr(tt.input)
			if result != tt.expected {
				t.Errorf("checksumHex() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEthSecp256k1AuthenticatorVerifyPositive(t *testing.T) {
	// 	priv: a0505da852036821eb3df07e8f8ee1ebef5ce50034133ea038aee10c8b4c9111
	// pub: 02f0a13ef50767acffd58328b4f69ef36866c4f350fe3fa6172598a883d03d501d
	// priv, pub, _ := crypto.GenerateSecp256k1Key(nil)
	// fmt.Printf("priv: %x\n", priv.Bytes())
	// fmt.Printf("pub: %x\n", pub.Bytes())

	privBts, _ := hex.DecodeString("a0505da852036821eb3df07e8f8ee1ebef5ce50034133ea038aee10c8b4c9111")
	priv, _ := crypto.UnmarshalSecp256k1PrivateKey(privBts)

	msg := []byte("test message")

	signer := GetUserSigner(priv)

	authnr := GetAuthenticator(signer.AuthType()) // nolint:misspell
	if _, is := authnr.(EthSecp256k1Authenticator); !is {
		t.Errorf("expected EthSecp256k1Authenticator, got %T", authnr)
	}

	sig, err := signer.Sign(msg)
	require.NoError(t, err)

	authnr = GetAuthenticator(sig.Type)
	if _, is := authnr.(EthSecp256k1Authenticator); !is {
		t.Errorf("expected EthSecp256k1Authenticator, got %T", authnr)
	}

	err = authnr.Verify(signer.CompactID(), msg, sig.Data)
	require.NoError(t, err)

	// now pretend to be metamask
	sig.Data[crypto.RecoveryIDOffset] += 27

	err = authnr.Verify(signer.CompactID(), msg, sig.Data)
	require.NoError(t, err)
}

func TestEthSecp256k1AuthenticatorVerifyNegative(t *testing.T) {
	tests := []struct {
		name      string
		identity  []byte
		msg       []byte
		signature []byte
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "Invalid recovery ID",
			identity:  []byte{0x1, 0x2, 0x3, 0x4},
			msg:       []byte("test message"),
			signature: []byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 29}, // recovery ID 29 is invalid
			wantErr:   true,
			errMsg:    "recovery",
		},
		{
			name:      "Empty message",
			identity:  []byte{0x1, 0x2, 0x3, 0x4},
			msg:       []byte{},
			signature: make([]byte, 65),
			wantErr:   true,
			errMsg:    "invalid signature",
		},
		{
			name:      "Short signature",
			identity:  []byte{0x1, 0x2, 0x3, 0x4},
			msg:       []byte("test message"),
			signature: []byte{0x1, 0x2, 0x3},
			wantErr:   true,
			errMsg:    "invalid signature length",
		},
		{
			name:      "Identity mismatch",
			identity:  make([]byte, 20),
			msg:       []byte("test message"),
			signature: make([]byte, 65),
			wantErr:   true,
			errMsg:    "invalid signature",
		},
		{
			name:      "Empty identity",
			identity:  []byte{},
			msg:       []byte("test message"),
			signature: make([]byte, 65),
			wantErr:   true,
			errMsg:    "invalid signature",
		},
	}

	auth := EthSecp256k1Authenticator{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.Verify(tt.identity, tt.msg, tt.signature)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
