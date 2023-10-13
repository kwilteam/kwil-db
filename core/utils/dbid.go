package utils

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
)

// DB ID is a convention. This is likely to change:
// https://github.com/kwilteam/kwil-db/issues/332

func GenerateDBID(name string, ownerPubKey []byte) string {
	h := sha256.New224()
	h.Write(bytes.ToLower([]byte(name)))
	h.Write(ownerPubKey)
	return "x" + hex.EncodeToString(h.Sum(nil))
}
