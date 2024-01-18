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

type uuidExtension struct{}

func (u *uuidExtension) Execute(scope CallContext, metadata map[string]string, method string, args ...any) ([]any, error) {
	lowerMethod := strings.ToLower(method)

	// if no args are provided, throw error
	if len(args) == 0 {
		return nil, fmt.Errorf("uuid: expected at least 1 argument, got 0")
	}

	// convert the every argument to a byte slice, as required by the uuid library
	var arg []byte

	// iterate over the arguments and convert them to byte slices, and append them to the arg slice
	for _, v := range args {
		switch v := v.(type) {
		default:
			return nil, fmt.Errorf("uuid: expected byte slice, string, or int64 as argument, got %T", v)
		case []byte:
			arg = append(arg, v...)
		case string:
			arg = append(arg, []byte(v)...)
		case int64:
			buf := new(bytes.Buffer)
			err := binary.Write(buf, binary.LittleEndian, v)
			if err != nil {
				return nil, fmt.Errorf("uuid: error converting int to byte slice: %w", err)
			}
			arg = append(arg, buf.Bytes()...)
		}
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
