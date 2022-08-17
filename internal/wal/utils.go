package wal

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"io"
	"math/big"
	"os"
	"path"
)

// Will append the slice length as uint16 to the end of the byte slice
// Use this function instead of doing it manually since this uses uint16 instead of int64
func appendByteArrLength(b []byte, a []byte) []byte {
	return append(b, uint16ToBytes(uint16(len(a)))...)
}

// This function converts a big int to bytes.  The result will always be a byte slice of length 16.
func bigInt2Bytes(h *big.Int) []byte {
	b := make([]byte, 16)
	k := h.FillBytes(b)
	return k
}

func uint16ToBytes(i uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, i)
	return b
}

// Get current node key to store associated WAL
// will ensure that the WAL is correlated to the
// correct chain if reset.
func concatWithRootChainPath(homeDir, name string) string {
	chainHash := getNodeKeyHash(homeDir)
	return path.Join(homeDir+".local", chainHash, name)
}

func getNodeKeyHash(dir string) string {
	f, err := os.Open(path.Join(dir, "config", "node_key.json"))
	if err != nil {
		panic(err)
	}

	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		panic(err)
	}

	return hex.EncodeToString(h.Sum(nil))
}
