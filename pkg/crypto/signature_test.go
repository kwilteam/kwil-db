package crypto

import (
	"encoding/hex"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSignature_Verify(t *testing.T) {
	msg := []byte("foo")
	anotherMsg := []byte("bar")

	secp256k1PubKeyHex := "04812bef44f6e7b2a19c0b01c2dca5e54ba1935a1890ffdcb93abd0c534b209c21e4f6176823fef493f7b5afaa456f31d5293363d8f801c540ebcc061812890cba"
	secp256k1PubKeyBytes, _ := hex.DecodeString(secp256k1PubKeyHex)
	secp256k1PublicKey, _ := loadSecp256k1PublicKeyFromByte(secp256k1PubKeyBytes)

	personalSignSig := "cb3fed7f6ff36e59054c04a831b215e514052753ee353e6fe31d4b4ef736acd6155127db555d3006ba14fcb4c79bbad56c8e63b81a9896319bb053a9e253475800"
	personalSignSigBytes, _ := hex.DecodeString(personalSignSig)

	cometbftSecp256k1Sig := "19a4aced02d5b9142b4f622b06442b1904445e16bd25409e6b0ff357ccc021d001d0e7824654b695b4b6e0991cb7507f487b82be4b2ed713d1e3e2cbc3d2518d"
	cometbftSecp256k1SigBytes, _ := hex.DecodeString(cometbftSecp256k1Sig)

	type fields struct {
		Signature []byte
		Type      SignatureType
	}
	type args struct {
		publicKey PublicKey
		msg       []byte
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "test secp256k1 personal_sign",
			fields: fields{
				Signature: personalSignSigBytes,
				Type:      SIGNATURE_TYPE_SECP256K1_PERSONAL,
			},
			args: args{
				publicKey: secp256k1PublicKey,
				msg:       msg,
			},
			wantErr: nil,
		},
		{
			name: "test secp256k1 personal_sign invalid signature",
			fields: fields{
				Signature: personalSignSigBytes[1:],
				Type:      SIGNATURE_TYPE_SECP256K1_PERSONAL,
			},
			args: args{
				publicKey: secp256k1PublicKey,
				msg:       msg,
			},
			wantErr: errInvalidSignature,
		},
		{
			name: "test secp256k1 personal_sign wrong signature",
			fields: fields{
				Signature: personalSignSigBytes,
				Type:      SIGNATURE_TYPE_SECP256K1_PERSONAL,
			},
			args: args{
				publicKey: secp256k1PublicKey,
				msg:       anotherMsg,
			},
			wantErr: errVerifySignatureFailed,
		},
		{
			name: "test secp256k1 cometbft",
			fields: fields{
				Signature: cometbftSecp256k1SigBytes,
				Type:      SIGNATURE_TYPE_SECP256K1_COMETBFT,
			},
			args: args{
				publicKey: secp256k1PublicKey,
				msg:       msg,
			},
			wantErr: nil,
		},
		{
			name: "test secp256k1 cometbft invalid signature",
			fields: fields{
				Signature: cometbftSecp256k1SigBytes[1:],
				Type:      SIGNATURE_TYPE_SECP256K1_COMETBFT,
			},
			args: args{
				publicKey: secp256k1PublicKey,
				msg:       msg,
			},
			wantErr: errInvalidSignature,
		},
		{
			name: "test secp256k1 cometbft wrong signature",
			fields: fields{
				Signature: cometbftSecp256k1SigBytes,
				Type:      SIGNATURE_TYPE_SECP256K1_COMETBFT,
			},
			args: args{
				publicKey: secp256k1PublicKey,
				msg:       anotherMsg,
			},
			wantErr: errVerifySignatureFailed,
		},
		{
			name: "unsupported signature type",
			fields: fields{
				Signature: nil,
				Type:      SIGNATURE_TYPE_INVALID,
			},
			args: args{
				publicKey: secp256k1PublicKey,
				msg:       msg,
			},
			wantErr: errNotSupportedSignatureType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Signature{
				Signature: tt.fields.Signature,
				Type:      tt.fields.Type,
			}
			err := s.Verify(tt.args.publicKey, tt.args.msg)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
