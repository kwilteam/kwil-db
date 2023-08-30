package addresses_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/crypto/addresses"
	"github.com/stretchr/testify/assert"
)

func init() {
	pk, err := crypto.Secp256k1PrivateKeyFromHex(secpKey)
	if err != nil {
		panic(err)
	}

	edPk, err := crypto.Ed25519PrivateKeyFromHex(edKey)
	if err != nil {
		panic(err)
	}

	secpPubKey = pk.PubKey().(*crypto.Secp256k1PublicKey)
	edPubKey = edPk.PubKey().(*crypto.Ed25519PublicKey)
}

var secpPubKey *crypto.Secp256k1PublicKey
var edPubKey *crypto.Ed25519PublicKey

func Test_Identifier(t *testing.T) {
	// testCase tests that each passed identifier can be marshaled and unmarshaled properly.
	// if an error is expected from the constructor, it should be specified in the testCase.
	type testCase struct {
		name   string
		pubkey crypto.PublicKey
		format addresses.AddressFormat
		err    error
	}

	testCases := []testCase{
		{
			name:   "secp with ethereum address - success",
			pubkey: secpPubKey,
			format: addresses.AddressFormatEthereum,
		},
		{
			name:   "secp with comet address - success",
			pubkey: secpPubKey,
			format: addresses.AddressFormatCometBFT,
		},
		{
			name:   "secp with near address - failure",
			pubkey: secpPubKey,
			format: addresses.AddressFormatNEAR,
			err:    addresses.ErrIncompatibleAddress,
		},
		{
			name:   "ed25519 with ethereum address - failure",
			pubkey: edPubKey,
			format: addresses.AddressFormatEthereum,
			err:    addresses.ErrIncompatibleAddress,
		},
		{
			name:   "ed25519 with comet address - success",
			pubkey: edPubKey,
			format: addresses.AddressFormatCometBFT,
		},
		{
			name:   "ed25519 with near address - success",
			pubkey: edPubKey,
			format: addresses.AddressFormatNEAR,
		},
		{
			name:   "invalid public key type - failure",
			pubkey: &invalidPubKey{},
			format: addresses.AddressFormatEthereum,
			err:    addresses.ErrInvalidKeyType,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			initial, err := addresses.CreateKeyIdentifier(tc.pubkey, tc.format)
			if tc.err != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tc.err)
				}
				if !errors.Is(err, tc.err) {
					t.Fatalf("expected error %v, got %v", tc.err, err)
				}
				return
			}
			assert.NoError(t, err, "unexpected error")

			bts, err := initial.MarshalBinary()
			assert.NoError(t, err, "unexpected error")

			reconstructed := &addresses.KeyIdentifier{}
			err = reconstructed.UnmarshalBinary(bts)
			assert.NoError(t, err, "unexpected error")

			assert.Equal(t, *initial, *reconstructed)

			expectedAddr, err := addresses.GenerateAddress(tc.pubkey, tc.format)
			assert.NoError(t, err, "unexpected error")

			initialAddress, err := initial.Address()
			assert.NoError(t, err, "unexpected error")

			reconstructedAddress, err := reconstructed.Address()
			assert.NoError(t, err, "unexpected error")

			assert.Equal(t, expectedAddr, initialAddress)
			assert.Equal(t, expectedAddr, reconstructedAddress)
		})
	}
}

type invalidPubKey struct{}

func (i *invalidPubKey) Address() crypto.Address {
	return nil
}

func (i *invalidPubKey) Bytes() []byte {
	return nil
}

func (i *invalidPubKey) Type() crypto.KeyType {
	return "invalid_key_type"
}

func (i *invalidPubKey) Verify(sig []byte, data []byte) error {
	return fmt.Errorf("invalid public key type - this is a test implementation of a public key")
}
