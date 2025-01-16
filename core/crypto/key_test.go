package crypto_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/core/crypto"
	"github.com/stretchr/testify/require"
)

func Test_RegisterKeys(t *testing.T) {
	// Test that the keys are registered
	err := crypto.RegisterKeyType(100, "testInvalidKT")
	require.Error(t, err)

	// Test Valid key
	validKeyType := crypto.KeyType(70000)
	validKTStr := "testKT"
	err = crypto.RegisterKeyType(validKeyType, validKTStr)
	require.NoError(t, err)
	require.Equal(t, validKTStr, validKeyType.String())

	kt, err := crypto.ParseKeyType(validKTStr)
	require.NoError(t, err)
	require.Equal(t, validKeyType, kt)

	// Register already registered keyType under different name
	duplicateKTStr := "testKT2"
	err = crypto.RegisterKeyType(validKeyType, duplicateKTStr)
	require.Error(t, err)

	// Register new keyType under already registered name
	newKT := crypto.KeyType(70001)
	err = crypto.RegisterKeyType(newKT, validKTStr)
	require.Error(t, err)
}
