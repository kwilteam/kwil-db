package transactions_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types/serialize"
	"github.com/kwilteam/kwil-db/core/types/transactions"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NOTE: this fails because of legacy issues with RLP itself. We should be aware
// of these issues with RLP (allows encoding nil fields, but cannot decode
// them). If we add the `rlp:"nil"` tag, it can decode, but it's a breaking
// change to transaction serialization.
func Test_TransactionMarshalUnmarshal(t *testing.T) {
	tx := &transactions.Transaction{}
	serialized, err := tx.MarshalBinary()
	require.NoError(t, err)

	tx2 := &transactions.Transaction{}
	err = tx2.UnmarshalBinary(serialized)
	require.Error(t, err)
}

// testing serialization of a transaction, since Luke found a bug
func Test_TransactionMarshal(t *testing.T) {
	tx := &transactions.Transaction{
		Signature: &auth.Signature{
			Signature: []byte("signature"),
			Type:      auth.EthPersonalSignAuth,
		},
		Body: &transactions.TransactionBody{
			Payload:     []byte("payload"),
			PayloadType: transactions.PayloadTypeDeploySchema,
			Fee:         big.NewInt(100),
			Nonce:       1,
			ChainID:     "chainIDXXX",
		},
		Sender: []byte("sender"),
	}

	serialized, err := tx.MarshalBinary()
	require.NoError(t, err)

	// For this version of the Transaction and TransactionBody, ensure the
	// serialization is stable.
	wantHash := [32]byte{0x71, 0xce, 0x11, 0xa4, 0xb1, 0xfc, 0x38, 0xe6, 0xe6, 0xe8,
		0xcd, 0x27, 0x9b, 0x2d, 0xc, 0xbd, 0x95, 0xea, 0xe, 0x8d, 0xe0, 0x46, 0x8c,
		0x35, 0x64, 0x21, 0x70, 0x31, 0x50, 0x92, 0x4, 0xf0}
	gotHash := sha256.Sum256(serialized)
	require.Equal(t, wantHash, gotHash)

	tx2 := &transactions.Transaction{}
	err = tx2.UnmarshalBinary(serialized)
	require.NoError(t, err)

	require.Equal(t, tx, tx2)
}

func Test_TransactionBodyMarshalJSON(t *testing.T) {
	txB := transactions.TransactionBody{ // not a pointer, ensure MarshalJSON method works for value
		Payload:     []byte("payload"),
		PayloadType: transactions.PayloadTypeDeploySchema,
		Fee:         big.NewInt(100),
		Nonce:       1,
		ChainID:     "chainIDXXX",
	}

	b, err := json.Marshal(txB)
	require.NoError(t, err)

	txB2 := transactions.TransactionBody{}
	err = json.Unmarshal(b, &txB2)
	require.NoError(t, err)

	require.Equal(t, txB, txB2)

	// Marshal pointer
	b, err = json.Marshal(&txB)
	require.NoError(t, err)

	txB3 := transactions.TransactionBody{}
	err = json.Unmarshal(b, &txB3)
	require.NoError(t, err)

	require.Equal(t, txB, txB3)
}

type actionExecutionV0 struct {
	DBID      string
	Action    string
	Arguments [][]*transactions.EncodedValue
	// No other optional or tail fields defined.
}

// TestActionExecution_Marshal ensures that the optional NilArg and tail Rest
// fields marshal as expected.
func TestActionExecution_Marshal(t *testing.T) {
	testRoundTrip := func(dat []byte, ae *transactions.ActionExecution) {
		var err error
		if len(dat) == 0 {
			dat, err = ae.MarshalBinary()
			require.NoError(t, err)
		}

		var ae2 transactions.ActionExecution
		err = ae2.UnmarshalBinary(dat)
		require.NoError(t, err)

		assert.EqualValues(t, ae, &ae2)
	}

	// All fields set, including optional NilArg
	ae := &transactions.ActionExecution{
		DBID:   "dbid",
		Action: "insert_thing",
		Arguments: [][]*transactions.EncodedValue{
			{
				mustDetect(nil),
				mustDetect("b"),
			},
			{
				mustDetect("c"),
				mustDetect(nil),
			},
		},
	}

	testRoundTrip(nil, ae)
	testRoundTrip(nil, ae)

	// Forward compat without the NilArg field at all. This is the main benefit
	// of optional and nil.
	aeOld := actionExecutionV0{
		DBID:      ae.DBID,
		Action:    ae.Action,
		Arguments: ae.Arguments,
		// NilArg not a field
	}

	dat, err := serialize.Encode(aeOld)
	require.NoError(t, err)

	testRoundTrip(dat, ae)
}

func TestTransaction_Sign(t *testing.T) {
	// secp256k1
	secp2561k1PvKeyHex := "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"
	secp256k1PrivateKey, err := crypto.Secp256k1PrivateKeyFromHex(secp2561k1PvKeyHex)
	require.NoError(t, err, "error parse private secp2561k1PvKeyHex")

	ethPersonalSigner := auth.EthPersonalSigner{Key: *secp256k1PrivateKey}

	expectPersonalSignConcatSigHex := "e09459d0dc078f12bb176da6ec52764ac457e322644f4031a6c498979795eff16163edbfe2c68ba60e25d6a76a283f63662245555caecf68889fbfad786ae52801"
	expectPersonalSignConcatSigBytes, _ := hex.DecodeString(expectPersonalSignConcatSigHex)
	expectPersonalSignConcatSig := &auth.Signature{
		Signature: expectPersonalSignConcatSigBytes,
		Type:      auth.EthPersonalSignAuth,
	}

	rawPayload := transactions.ActionExecution{
		DBID:   "xf617af1ca774ebbd6d23e8fe12c56d41d25a22d81e88f67c6c6ee0d4",
		Action: "create_user",
		Arguments: [][]*transactions.EncodedValue{
			{
				mustDetect("foo"),
				mustDetect(32),
			},
		},
		// NOTE: With NilArg unset (and optional), the expectPersonalSignConcatSigHex
		// is the same as if it were not a defined field at all.
	}
	payloadRLP, err := rawPayload.MarshalBinary()
	require.NoError(t, err)

	type args struct {
		mst    transactions.SignedMsgSerializationType
		signer auth.Signer
	}
	tests := []struct {
		name          string
		args          args
		wantSig       *auth.Signature
		authenticator auth.Authenticator
		wantErr       bool
	}{
		{
			name: "not support message serialization type",
			args: args{
				mst:    transactions.SignedMsgSerializationType("not support message serialization type"),
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
			authenticator: &auth.EthSecp256k1Authenticator{},
			wantSig:       expectPersonalSignConcatSig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			tx := transactions.Transaction{
				Body: &transactions.TransactionBody{
					Description: "By signing this message, you'll reveal your xxx to zzz",
					Payload:     payloadRLP,
					PayloadType: rawPayload.Type(),
					Fee:         big.NewInt(100),
					Nonce:       1,
					ChainID:     "adsf",
				},
				Serialization: tt.args.mst,
			}

			err := tx.Sign(tt.args.signer)
			if tt.wantErr {
				assert.Error(t1, err, "Sign(%v)", tt.args.mst)
				return
			}

			require.NoError(t1, err, "error signing tx")
			require.Equal(t1, tt.wantSig.Type, tx.Signature.Type,
				"mismatch signature type")
			require.Equal(t1, hex.EncodeToString(tt.wantSig.Signature),
				hex.EncodeToString(tx.Signature.Signature), "mismatch signature")

			msgBts, err := tx.SerializeMsg()
			require.NoError(t1, err, "error serializing message")

			err = tt.authenticator.Verify(tx.Sender, msgBts, tx.Signature.Signature)
			require.NoError(t1, err, "error verifying message")
		})
	}
}

func TestTransactionBody_SerializeMsg(t *testing.T) {
	rawPayload := transactions.ActionExecution{
		DBID:   "xf617af1ca774ebbd6d23e8fe12c56d41d25a22d81e88f67c6c6ee0d4",
		Action: "create_user",
		Arguments: [][]*transactions.EncodedValue{
			{
				mustDetect("foo"),
				mustDetect(32),
			},
		},
	}

	payload, err := rawPayload.MarshalBinary()
	require.NoError(t, err)

	defaultDescription := "By signing this message, you'll reveal your xxx to zzz"
	longDescription := `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ
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
				description: longDescription,
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
			wantMsg: "4279207369676e696e672074686973206d6573736167652c20796f75276c6c2072657665616c20796f75722078787820746f207a7a7a0a0a5061796c6f6164547970653a20657865637574650a5061796c6f61644469676573743a20323038623838653133656336313866313836376564333534366131343861656335633835316631310a4665653a203130300a4e6f6e63653a20310a0a4b77696c20436861696e2049443a2030303030303030303030300a",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			txBody := &transactions.TransactionBody{
				Description: tt.args.description,
				Payload:     payload,
				PayloadType: rawPayload.Type(),
				Fee:         big.NewInt(100),
				Nonce:       1,
				ChainID:     "00000000000",
			}

			got, err := txBody.SerializeMsg(tt.args.mst)
			if tt.wantErr { // TODO: verify expect error
				assert.Error(t1, err, "SerializeMsg(%v)", tt.args.mst)
				return
			}

			assert.NoError(t1, err, "SerializeMsg(%v)", tt.args.mst)
			assert.Equalf(t1, tt.wantMsg, hex.EncodeToString(got), "SerializeMsg(%v)", tt.args.mst)
			fmt.Printf("msg to sign: \n%s\n", string(got))
		})
	}
}
