package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// DB ID is a convention. This is likely to change:
// https://github.com/kwilteam/kwil-db/issues/332

func GenerateDBID(name string, ownerID []byte) string {
	h := sha256.New224()
	h.Write([]byte(strings.ToLower(name)))
	h.Write(ownerID)
	return "x" + hex.EncodeToString(h.Sum(nil))
}
