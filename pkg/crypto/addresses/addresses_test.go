package addresses_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/crypto/addresses"
)

const (
	secpKey = "f1aa5a7966c3863ccde3047f6a1e266cdc0c76b399e256b8fede92b1c69e4f4e"
	edKey   = "7c67e60fce0c403ff40193a3128e5f3d8c2139aed36d76d7b5f1e70ec19c43f00aa611bf555596912bc6f9a9f169f8785918e7bab9924001895798ff13f05842"
)

// Test_Addresses tests all types of addresses supported.
// This is mostly useful as a way of testing for our underlying dependencies
// changing, since many are not stable.
// It detects the key type based on the length of the private key.
func Test_Addresses(t *testing.T) {
	type testCase struct {
		name   string
		key    string // hex encoded private key
		output string // hex encoded address
		fn     addressFunc
	}

	testCases := []testCase{
		{
			name:   "ethereum",
			key:    secpKey,
			output: "0xc89D42189f0450C2b2c3c61f58Ec5d628176A1E7",
			fn: func(pk crypto.PrivateKey) (crypto.Address, error) {
				return addresses.CreateEthereumAddress(pk.PubKey().(*crypto.Secp256k1PublicKey))
			},
		},
		{
			name:   "cosmos_secp256k1",
			key:    secpKey,
			output: "6E741B9E60A1DFB6FE40B53069CFBD00A6C1FC88",
			fn: func(pk crypto.PrivateKey) (crypto.Address, error) {
				return addresses.CreateCosmosAddress(pk.PubKey())
			},
		},
		{
			name:   "cosmos_ed25519",
			key:    edKey,
			output: "E7620EFA925B10DBE5DDFA8D68B404AB338E98EC",
			fn: func(pk crypto.PrivateKey) (crypto.Address, error) {
				return addresses.CreateCosmosAddress(pk.PubKey())
			},
		},
		{
			name:   "near",
			key:    edKey,
			output: "0aa611bf555596912bc6f9a9f169f8785918e7bab9924001895798ff13f05842",
			fn: func(pk crypto.PrivateKey) (crypto.Address, error) {
				return addresses.CreateNearAddress(pk.PubKey().(*crypto.Ed25519PublicKey))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var pk crypto.PrivateKey
			var err error
			switch len(tc.key) {
			case 64:
				pk, err = crypto.Secp256k1PrivateKeyFromHex(tc.key)
			case 128:
				pk, err = crypto.Ed25519PrivateKeyFromHex(tc.key)
			}
			if err != nil {
				t.Fatal(err)
			}

			addr, err := tc.fn(pk)
			if err != nil {
				t.Fatal(err)
			}

			if addr.String() != tc.output {
				t.Errorf("Expected %s, got %s", tc.output, addr.String())
			}
		})
	}
}

type addressFunc func(crypto.PrivateKey) (crypto.Address, error)
