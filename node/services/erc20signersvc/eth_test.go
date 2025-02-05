package signersvc

import (
	"context"
	"flag"
	"math/big"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

var ethRpc = flag.String("eth-rpc", os.Getenv("ETH_RPC"), "eth provider rpc")

func TestSafe_metadata(t *testing.T) {
	if *ethRpc == "" {
		t.Skip("no eth rpc configured")
	}

	blockNumber := new(big.Int).SetUint64(7660784)

	s, err := NewSafe("11155111", *ethRpc, "0x56D510E4782cDed87F8B93D260282776adEd3f4B")
	require.NoError(t, err)

	ctx := context.Background()

	got, err := s.getSafeMetadata3(ctx, blockNumber)
	require.NoError(t, err)

	got2, err := s.getSafeMetadata(ctx, blockNumber)
	require.NoError(t, err)

	require.EqualValues(t, got, got2)
}
