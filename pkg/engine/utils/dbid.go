package utils

import (
	"strings"

	"github.com/kwilteam/kwil-db/pkg/crypto"
)

func GenerateDBID(name string, ownerPubKey []byte) string {
	return "x" + crypto.Sha224Hex(joinBytes([]byte(strings.ToLower(name)), ownerPubKey))
}

// joinBytes is a helper function to join multiple byte slices into one
func joinBytes(s ...[]byte) []byte {
	n := 0
	for _, v := range s {
		n += len(v)
	}

	b, i := make([]byte, n), 0
	for _, v := range s {
		i += copy(b[i:], v)
	}
	return b
}
