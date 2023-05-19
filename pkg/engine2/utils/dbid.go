package utils

import (
	"strings"

	"github.com/kwilteam/kwil-db/pkg/crypto"
)

func GenerateDBID(name, owner string) string {
	return "x" + crypto.Sha224Hex(joinBytes([]byte(strings.ToLower(name)), []byte(strings.ToLower(owner))))
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
