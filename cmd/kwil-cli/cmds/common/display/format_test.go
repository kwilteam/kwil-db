package display_test

import (
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/cmd/kwil-cli/cmds/common/display"
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

func (d *demoFormat) MarshalText() (string, error) {
	return fmt.Sprintf("Whatever format: %s", d.data), nil
}

func Example_wrappedMsg_text() {
	msg := &demoFormat{data: []byte("demo")}
	wrapped := display.WrapMsg(msg, nil)
	display.Print(wrapped, nil, "text")
	// Output: Whatever format: demo
}

func Example_wrappedMsg_json() {
	msg := &demoFormat{data: []byte("demo")}
	wrapped := display.WrapMsg(msg, nil)
	display.Print(wrapped, nil, "json")
	// Output: {
	//   "result": {
	//     "name_to_whatever": "demo_whatever"
	//   },
	//   "error": ""
	// }
}
