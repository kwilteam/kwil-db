package txsvc

import (
	"testing"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/transactions"
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

	abciTx, err := convertToAbciTx(incomingTx)
	require.NoError(t, err, "convertToAbciTx failed")

	outgoingTx, err := convertFromAbciTx(abciTx)
	require.NoError(t, err, "convertFromAbciTx failed")

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

	incomingAbciTx, err := convertToAbciTx(incomingTx)
	require.NoError(t, err, "convertToAbciTx failed")

	abciTxBytes, err := incomingAbciTx.MarshalBinary()
	require.NoError(t, err, "incomingAbciTx.MarshalBinary failed")
	outgoingAbciTx := &transactions.Transaction{}
	err = outgoingAbciTx.UnmarshalBinary(abciTxBytes)
	require.NoError(t, err, "outgoingAbciTx.UnmarshalBinary failed")

	outgoingTx, err := convertFromAbciTx(outgoingAbciTx)
	require.NoError(t, err, "convertFromAbciTx failed")

	require.EqualValues(t, incomingTx, outgoingTx, "convertToAbciTx and convertFromAbciTx are not inverse")
}
