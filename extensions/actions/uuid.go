//go:build actions_uuid || ext_test

package actions

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"strings"

	uuid "github.com/kwilteam/kwil-db/core/types"
)

const uuidName = "uuid"

func init() {
	u := &uuidExtension{}
	err := RegisterExtension("uuid", u)
	if err != nil {
		panic(err)
	}
}

type uuidExtension struct {}

func (u *uuidExtension) Execute(scope CallContext, metadata map[string]string, method string, args ...any) ([]any, error) {
	lowerMethod := strings.ToLower(method)
	if len(args) != 1 {
		return nil, fmt.Errorf("uuid: expected 1 argument, got %d", len(args))
	}
	
	// convert the first argument to a byte slice, as required by the uuid library
	var arg []byte
	switch v := args[0].(type) {
	default:
		return nil, fmt.Errorf("uuid: expected byte slice or string as first argument")
	case []byte:
		arg = v
	case string:
		arg = []byte(v)
	case int:
		buf := new(bytes.Buffer)
		// Question: should this be big endian or little endian? @brennanjl
		err := binary.Write(buf, binary.LittleEndian, v)
		if err != nil {
			return nil, fmt.Errorf("uuid: error converting int to byte slice: %w", err)
		}
		arg = buf.Bytes()
	}

	// there will be two methods on the extension:
	// - uuidv5: generates a uuidv5 from a byte slice and returns as a string
	// - uuidv5bytes: generates a uuidv5 from a byte slice and returns as a byte slice
	switch lowerMethod {
	default:
		return nil, fmt.Errorf("uuid: unknown method %s", method)
	case "uuidv5":
		return []any{uuid.NewUUIDV5(arg).String()}, nil
	}
}

// Takes no initialization parameters.
func (a *uuidExtension) Initialize(ctx context.Context, metadata map[string]string) (map[string]string, error) {
	return nil, nil
}