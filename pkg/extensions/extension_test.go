package extensions_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/extensions"
)

var (
	testConfig = map[string]string{
		"eth_provider": "wss://infura.io",
	}
)

// TODO: these tests are pretty bad.
// since this is a prototype, and the package is simple, this is good for now.
func Test_Extensions(t *testing.T) {
	ctx := context.Background()
	ext := extensions.New("local:8080")

	err := ext.Connect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	instance, err := ext.CreateInstance(ctx, map[string]string{
		"token_address":  "0x12345",
		"wallet_address": "0xabcd",
	})
	if err != nil {
		t.Fatal(err)
	}

	results, err := instance.Execute(ctx, "method1", "0x12345")
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
