package conversion

import (
	"testing"

	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/stretchr/testify/require"
)

func Test_convertTx(t *testing.T) {
	incomingTx := &txpb.Transaction{
		Body: &txpb.Transaction_Body{
			Payload:     []byte("payload"),
			PayloadType: "execute_action",
			Fee:         "10",
			Nonce:       1,
			Salt:        []byte("salt"),
		},
		Signature: &txpb.Signature{
			SignatureBytes: []byte("signature"),
			SignatureType:  "invalid",
		},
		Sender: []byte("sender"),
	}

	abciTx, err := ConvertToAbciTx(incomingTx)
	require.NoError(t, err, "convertToAbciTx failed")

	outgoingTx := ConvertFromAbciTx(abciTx)

	require.EqualValues(t, incomingTx, outgoingTx, "convertToAbciTx and convertFromAbciTx are not inverse")
}

func Test_convertTxWithSerialization(t *testing.T) {
	encodedPayload := []byte("payload")
	incomingTx := &txpb.Transaction{
		Body: &txpb.Transaction_Body{
			Payload:     encodedPayload,
			PayloadType: "execute_action",
			Fee:         "10",
			Nonce:       1,
			Salt:        []byte("salt"),
		},
		Signature: &txpb.Signature{
			SignatureBytes: []byte("signature"),
			SignatureType:  "invalid",
		},
		Sender: []byte("sender"),
	}

	incomingAbciTx, err := ConvertToAbciTx(incomingTx)
	require.NoError(t, err, "convertToAbciTx failed")

	abciTxBytes, err := incomingAbciTx.MarshalBinary()
	require.NoError(t, err, "incomingAbciTx.MarshalBinary failed")
	outgoingAbciTx := &transactions.Transaction{}
	err = outgoingAbciTx.UnmarshalBinary(abciTxBytes)
	require.NoError(t, err, "outgoingAbciTx.UnmarshalBinary failed")

	outgoingTx := ConvertFromAbciTx(outgoingAbciTx)

	require.EqualValues(t, incomingTx, outgoingTx, "convertToAbciTx and convertFromAbciTx are not inverse")
}
