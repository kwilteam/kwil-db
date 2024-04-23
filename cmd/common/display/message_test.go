package display

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/types/transactions"

	"github.com/stretchr/testify/assert"
)

// NOTE: could do this for all the other tests,
// but using Example* is more handy and obvious
func Test_RespTxHash(t *testing.T) {
	resp := RespTxHash("1024")
	expectJson := `{"tx_hash":"31303234"}`
	expectText := `TxHash: 31303234`

	outText, err := resp.MarshalText()
	assert.NoError(t, err, "MarshalText should not return error")
	assert.Equal(t, expectText, string(outText), "MarshalText should return expected text")

	outJson, err := resp.MarshalJSON()
	assert.NoError(t, err, "MarshalJSON should not return error")
	assert.Equal(t, expectJson, string(outJson), "MarshalJSON should return expected json")
}

func ExampleRespTxHash_text() {
	msg := wrapMsg(RespTxHash("1024"), nil)
	prettyPrint(msg, "text", os.Stdout, os.Stderr)
	// Output:
	// TxHash: 31303234
}

func TestRespTxHash_text_withError(t *testing.T) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	err := errors.New("an error")
	msg := wrapMsg(RespTxHash("1024"), err)
	prettyPrint(msg, "text", &stdout, &stderr)

	output, err := io.ReadAll(&stdout)
	assert.NoError(t, err, "ReadAll should not return error")
	assert.Equal(t, "", string(output), "stdout should be empty")

	errput, err := io.ReadAll(&stderr)
	assert.NoError(t, err, "ReadAll should not return error")
	assert.Equal(t, "an error\n", string(errput), "stderr should contain error")
}

func ExampleRespTxHash_json() {
	msg := wrapMsg(RespTxHash("1024"), nil)
	prettyPrint(msg, "json", os.Stdout, os.Stderr)
	// Output:
	// {
	//   "result": {
	//     "tx_hash": "31303234"
	//   },
	//   "error": ""
	// }
}

func ExampleRespTxHash_json_withError() {
	err := errors.New("an error")
	msg := wrapMsg(RespTxHash("1024"), err)
	prettyPrint(msg, "json", os.Stdout, os.Stderr)
	// Output:
	// {
	//   "result": "",
	//   "error": "an error"
	// }
}

func getExampleTxQueryResponse() *transactions.TcTxQueryResponse {
	secp256k1EpSigHex := "cb3fed7f6ff36e59054c04a831b215e514052753ee353e6fe31d4b4ef736acd6155127db555d3006ba14fcb4c79bbad56c8e63b81a9896319bb053a9e253475800"
	secp256k1EpSigBytes, _ := hex.DecodeString(secp256k1EpSigHex)
	secpSig := auth.Signature{
		Signature: secp256k1EpSigBytes,
		Type:      auth.EthPersonalSignAuth,
	}

	rawPayload := transactions.ActionExecution{
		DBID:   "xf617af1ca774ebbd6d23e8fe12c56d41d25a22d81e88f67c6c6ee0d4",
		Action: "create_user",
		Arguments: [][]string{
			{"foo", "32"},
		},
	}

	payloadRLP, err := rawPayload.MarshalBinary()
	if err != nil {
		panic(err)
	}

	return &transactions.TcTxQueryResponse{
		Hash:   []byte("1024"),
		Height: 10,
		Tx: transactions.Transaction{
			Body: &transactions.TransactionBody{
				Payload:     payloadRLP,
				PayloadType: rawPayload.Type(),
				Fee:         big.NewInt(100),
				Nonce:       10,
				ChainID:     "asdf",
				Description: "This is a test transaction for cli",
			},
			Serialization: transactions.SignedMsgConcat,
			Signature:     &secpSig,
		},
		TxResult: transactions.TransactionResult{
			Code:      0,
			Log:       "This is log",
			GasUsed:   10,
			GasWanted: 10,
			Data:      nil,
			Events:    nil,
		},
	}
}

func Example_respTxQuery_text() {
	Print(&RespTxQuery{Msg: getExampleTxQueryResponse()}, nil, "text")
	// Output:
	// Transaction ID: 31303234
	// Status: success
	// Height: 10
	// Log: This is log
}

func Example_respTxQuery_json() {
	Print(&RespTxQuery{Msg: getExampleTxQueryResponse()}, nil, "json")
	// Output:
	// {
	//   "result": {
	//     "hash": "31303234",
	//     "height": 10,
	//     "tx": {
	//       "signature": {
	//         "sig": "yz/tf2/zblkFTASoMbIV5RQFJ1PuNT5v4x1LTvc2rNYVUSfbVV0wBroU/LTHm7rVbI5juBqYljGbsFOp4lNHWAA=",
	//         "type": "secp256k1_ep"
	//       },
	//       "body": {
	//         "desc": "This is a test transaction for cli",
	//         "payload": "AAH4ULg5eGY2MTdhZjFjYTc3NGViYmQ2ZDIzZThmZTEyYzU2ZDQxZDI1YTIyZDgxZTg4ZjY3YzZjNmVlMGQ0i2NyZWF0ZV91c2VyyMeDZm9vgjMy",
	//         "type": "execute_action",
	//         "fee": 100,
	//         "nonce": 10,
	//         "chain_id": "asdf"
	//       },
	//       "serialization": "concat",
	//       "sender": null
	//     },
	//     "tx_result": {
	//       "code": 0,
	//       "log": "This is log",
	//       "gas_used": 10,
	//       "gas_wanted": 10
	//     }
	//   },
	//   "error": ""
	// }
}

func Test_TxHashAndExecResponse(t *testing.T) {
	hash := []byte{1, 2, 3, 4, 5}
	qr := getExampleTxQueryResponse()
	qr.Hash = hash
	resp := &TxHashAndExecResponse{
		Hash:      hash,
		QueryResp: &RespTxQuery{Msg: qr},
	}
	expectJson := `{"tx_hash":"0102030405","exec_result":{"hash":"0102030405","height":10,"tx":` +
		`{` +
		`"signature":{"sig":"yz/tf2/zblkFTASoMbIV5RQFJ1PuNT5v4x1LTvc2rNYVUSfbVV0wBroU/LTHm7rVbI5juBqYljGbsFOp4lNHWAA=","type":"secp256k1_ep"},` +
		`"body":{"desc":"This is a test transaction for cli","payload":"AAH4ULg5eGY2MTdhZjFjYTc3NGViYmQ2ZDIzZThmZTEyYzU2ZDQxZDI1YTIyZDgxZTg4ZjY3YzZjNmVlMGQ0i2NyZWF0ZV91c2VyyMeDZm9vgjMy","type":"execute_action","fee":100,"nonce":10,"chain_id":"asdf"},` +
		`"serialization":"concat","sender":null},` +
		`"tx_result":{"code":0,"log":"This is log","gas_used":10,"gas_wanted":10}}` +
		`}`
	expectText := "TxHash: 0102030405\nStatus: success\nHeight: 10\nLog: This is log"

	outText, err := resp.MarshalText()
	assert.NoError(t, err, "MarshalText should not return error")
	assert.Equal(t, expectText, string(outText), "MarshalText should return expected text")

	outJson, err := resp.MarshalJSON()
	fmt.Println(string(outJson))
	assert.NoError(t, err, "MarshalJSON should not return error")
	assert.Equal(t, expectJson, string(outJson), "MarshalJSON should return expected json")
}
