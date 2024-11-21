package main

import (
	"fmt"
	"testing"

	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/stretchr/testify/require"
)

func TestKey(t *testing.T) {
	priv, err := crypto.GenerateSecp256k1Key()
	require.NoError(t, err)

	fmt.Printf("Generated new key: %v\n\n", priv.Hex())

	// generate public key
	pubKeyBts := priv.PubKey().Bytes()

	pub, err := ethCrypto.UnmarshalPubkey(pubKeyBts)
	if err != nil {
		panic(err)
	}

	addr := ethCrypto.PubkeyToAddress(*pub)
	fmt.Println(addr.Hex())
}
