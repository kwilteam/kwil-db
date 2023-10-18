package transactions_test

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallMessageBody_SerializeMsg(t *testing.T) {
	callPayload := transactions.ActionCall{
		DBID:      "xf617af1ca774ebbd6d23e8fe12c56d41d25a22d81e88f67c6c6ee0d4",
		Action:    "action0",
		Arguments: []string{"foo"},
	}

	payloadRLP, err := callPayload.MarshalBinary()
	require.NoError(t, err, "MarshalBinary()")

	defaultDescription := "By signing this message, you'll bla bla"
	longDescrption := `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
`

	type args struct {
		mst         transactions.SignedMsgSerializationType
		description string
	}

	tests := []struct {
		name    string
		args    args
		wantMsg string //hex string
		wantErr bool
	}{
		{
			name: "non support message serialization type",
			args: args{
				mst:         transactions.SignedMsgSerializationType("non support message serialization type"),
				description: defaultDescription,
			},
			wantMsg: "",
			wantErr: true,
		},
		{
			name: "description too long",
			args: args{
				mst:         transactions.SignedMsgConcat,
				description: longDescrption,
			},
			wantMsg: "",
			wantErr: true,
		},
		{
			name: "concat string",
			args: args{
				mst:         transactions.SignedMsgConcat,
				description: defaultDescription,
			},
			wantMsg: "4279207369676e696e672074686973206d6573736167652c20796f75276c6c20626c6120626c610a0a444249443a207866363137616631636137373465626264366432336538666531326335366434316432356132326438316538386636376336633665653064340a416374696f6e3a20616374696f6e300a5061796c6f61644469676573743a20383531303336393937326263643639663762636439366361353464366338346264636534326631620a0a4b77696c20f09f968b0a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			callMsgBody := transactions.CallMessageBody{
				Payload:     payloadRLP,
				Description: tt.args.description,
			}

			got, err := callMsgBody.SerializeMsg(tt.args.mst)
			if tt.wantErr {
				assert.Error(t1, err, "SerializeMsg(%v)", tt.args.mst)
				return
			}

			assert.NoError(t1, err, "SerializeMsg(%v)", tt.args.mst)
			fmt.Printf("msg to sign: \n%s\n", string(got))
			assert.Equalf(t1, tt.wantMsg, hex.EncodeToString(got), "SerializeMsg(%v)", tt.args.mst)
		})
	}
}

func TestCallMessage_Sign(t *testing.T) {
	// secp256k1
	secp2561k1PvKeyHex := "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"
	secp256k1PrivateKey, err := crypto.Secp256k1PrivateKeyFromHex(secp2561k1PvKeyHex)
	require.NoError(t, err, "error parse private secp2561k1PvKeyHex")

	ethPersonalSigner := auth.EthPersonalSigner{Secp256k1PrivateKey: *secp256k1PrivateKey}

	expectPersonalSignConcatSigHex := "fdb2360f631cad62572a413d041259c95239cab73bccea9f758425548fcca33d681b6c64fdfc1db1aa034c85a49acd561e52094710a4334ff35b30b73ea307df00"
	expectPersonalSignConcatSigBytes, _ := hex.DecodeString(expectPersonalSignConcatSigHex)
	expectPersonalSignConcatSig := &auth.Signature{
		Signature: expectPersonalSignConcatSigBytes,
		Type:      auth.EthPersonalSignAuth,
	}

	callPayload := transactions.ActionCall{
		DBID:      "xf617af1ca774ebbd6d23e8fe12c56d41d25a22d81e88f67c6c6ee0d4",
		Action:    "action0",
		Arguments: []string{"foo"},
	}

	payloadRLP, err := callPayload.MarshalBinary()
	require.NoError(t, err, "MarshalBinary()")

	type args struct {
		mst    transactions.SignedMsgSerializationType
		signer auth.Signer
	}
	tests := []struct {
		name    string
		args    args
		wantSig *auth.Signature
		wantErr bool
	}{
		{
			name: "non support message serialization type",
			args: args{
				mst:    transactions.SignedMsgSerializationType("non support message serialization type"),
				signer: &ethPersonalSigner,
			},
			wantErr: true,
		},
		{
			name: "eth personal_sign concat string",
			args: args{
				mst:    transactions.SignedMsgConcat,
				signer: &ethPersonalSigner,
			},
			wantSig: expectPersonalSignConcatSig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			msg := transactions.CallMessage{
				Body: &transactions.CallMessageBody{
					Description: "By signing this message, you'll bla bla",
					Payload:     payloadRLP,
				},
				Serialization: tt.args.mst,
			}

			err := msg.Sign(tt.args.signer)
			if tt.wantErr {
				assert.Error(t1, err, "Sign(%v)", tt.args.mst)
				return
			}

			require.NoError(t1, err, "error signing tx")
			require.Equal(t1, tt.wantSig.Type, msg.Signature.Type,
				"mismatch signature type")
			require.Equal(t1, hex.EncodeToString(tt.wantSig.Signature),
				hex.EncodeToString(msg.Signature.Signature), "mismatch signature")

			require.True(t, msg.IsSigned())
			msgBts, err := msg.SerializeMsg()
			require.NoError(t1, err, "error serializing message")

			authenticator := tt.args.signer.Authenticator()
			err = authenticator.Verify(msg.Sender, msgBts, msg.Signature.Signature)
			require.NoError(t1, err, "error verifying message")
		})
	}

}
