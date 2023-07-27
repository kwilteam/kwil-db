package arweave_test

import (
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/arweave"
	"github.com/kwilteam/kwil-db/pkg/crypto"
)

// since these are essentially integration tests, we don't want to run them in CI
// bundlr is totally optional and experimental
const runTest = false

const (
	bundlrEndpoint = "https://node1.bundlr.network"
	privKey        = ""
)

func Test_Bundlr(t *testing.T) {
	if !runTest {
		t.Skip()
	}

	ecdsaPrivKey, err := crypto.ECDSAFromHex(privKey)
	if err != nil {
		t.Fatal(err)
	}

	pubkey := crypto.AddressFromPrivateKey(ecdsaPrivKey)
	fmt.Println("pubkey", pubkey)

	client, err := arweave.NewBundlrClient(bundlrEndpoint, ecdsaPrivKey)
	if err != nil {
		t.Fatal(err)
	}

	res, err := client.StoreItem([]byte("hello wsorsfld"))
	if err != nil {
		t.Fatal(err)
	}

	t.Log(res)
}
