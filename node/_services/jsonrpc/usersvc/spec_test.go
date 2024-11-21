package usersvc_test

import (
	"encoding/json"
	"testing"

	"github.com/kwilteam/kwil-db/internal/services/jsonrpc/usersvc"
)

func TestSpec(t *testing.T) {
	spec := usersvc.OpenRPCSpec()
	_, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	// fmt.Println(string(b))
}
