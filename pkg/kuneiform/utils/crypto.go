package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func Sha224(data []byte) []byte {
	h := sha256.New224()
	h.Write(data)
	return h.Sum(nil)
}

func Sha224Hex(data []byte) string {
	return hex.EncodeToString(Sha224(data))
}
