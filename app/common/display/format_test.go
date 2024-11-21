package display

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type demoFormat struct {
	data []byte
}

func (d *demoFormat) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Data string `json:"name_to_whatever"`
	}{
		Data: string(d.data) + "_whatever",
	})
}

func (d *demoFormat) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("Whatever format: %s", d.data)), nil
}

func Example_wrappedMsg_text() {
	msg := wrapMsg(&demoFormat{data: []byte("demo")}, nil)
	prettyPrint(msg, "text", os.Stdout, os.Stderr)
	// Output: Whatever format: demo
}

func Test_wrappedMsg_text_withError(t *testing.T) {
	var stderr bytes.Buffer
	var stdout bytes.Buffer

	err := errors.New("an error")
	msg := wrapMsg(&demoFormat{data: []byte("demo")}, err)
	prettyPrint(msg, "text", &stdout, &stderr)

	output := stdout.Bytes()
	assert.Equal(t, "", string(output), "stdout should be empty")

	errput := stderr.Bytes()
	assert.Equal(t, "an error\n", string(errput), "stderr should contain error")
}

func Example_wrappedMsg_json() {
	msg := wrapMsg(&demoFormat{data: []byte("demo")}, nil)
	prettyPrint(msg, "json", os.Stdout, os.Stderr)
	// Output: {
	//   "result": {
	//     "name_to_whatever": "demo_whatever"
	//   },
	//   "error": ""
	// }
}

func Example_wrappedMsg_json_withError() {
	err := errors.New("an error")
	msg := wrapMsg(&demoFormat{data: []byte("demo")}, err)
	prettyPrint(msg, "json", os.Stdout, os.Stderr)
	// Output:
	// {
	//   "result": null,
	//   "error": "an error"
	// }
}
