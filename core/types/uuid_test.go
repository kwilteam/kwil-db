package types_test

import (
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
