package display_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/cmd/internal/display"
	"github.com/stretchr/testify/assert"
)

// NOTE: could do this for all the other tests,
// but using Example* is more handy and obvious
func Test_RespTxHash(t *testing.T) {
	resp := display.RespTxHash("1024")
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
	msg := display.WrapMsg(display.RespTxHash("1024"), nil)
	display.PrettyPrint(msg, "text", os.Stdout, os.Stderr)
	// Output:
	// TxHash: 31303234
}

func TestRespTxHash_text_withError(t *testing.T) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	err := errors.New("an error")
	msg := display.WrapMsg(display.RespTxHash("1024"), err)
	display.PrettyPrint(msg, "text", &stdout, &stderr)

	output, err := io.ReadAll(&stdout)
	assert.NoError(t, err, "ReadAll should not return error")
	assert.Equal(t, "", string(output), "stdout should be empty")

	errput, err := io.ReadAll(&stderr)
	assert.NoError(t, err, "ReadAll should not return error")
	assert.Equal(t, "an error\n", string(errput), "stderr should contain error")
}

func ExampleRespTxHash_json() {
	msg := display.WrapMsg(display.RespTxHash("1024"), nil)
	display.PrettyPrint(msg, "json", os.Stdout, os.Stderr)
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
	msg := display.WrapMsg(display.RespTxHash("1024"), err)
	display.PrettyPrint(msg, "json", os.Stdout, os.Stderr)
	// Output:
	// {
	//   "result": "",
	//   "error": "an error"
	// }
}
