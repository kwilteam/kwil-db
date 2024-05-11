package types_test

import (
	"encoding/json"
	"testing"

	"github.com/kwilteam/kwil-db/core/types"
)

func Test_UUID(t *testing.T) {
	seed := []byte("test")

	uuid := types.NewUUIDV5(seed)

	if uuid.String() != "24aa70cf-0e18-57c9-b449-da8c9db37821" {
		t.Errorf("expected uuid to be 24aa70cf-0e18-57c9-b449-da8c9db37821, got %s", uuid.String())
	}
}

func Test_UUIDJSON(t *testing.T) {
	seed := []byte("test")

	uuid := types.NewUUIDV5(seed)

	b, err := json.Marshal(uuid)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(uuid)

	var uuid3 types.UUID
	err = json.Unmarshal(b, &uuid3)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(uuid3) // 00000000-0000-0000-0000-000000000000
}
