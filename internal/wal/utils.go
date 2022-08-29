package wal

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/kwilteam/kwil-db/internal/utils/errs"
	"io"
	"os"
	"path"
)

// Get current node key to store associated WAL
// will ensure that the WAL is correlated to the
// correct chain if reset.
func concatWithRootChainPath(homeDir, name string) string {
	chainHash := getNodeKeyHash(homeDir)
	return path.Join(homeDir+".local", chainHash, name)
}

func getNodeKeyHash(dir string) string {
	f, err := os.Open(path.Join(dir, "config", "node_key.json"))
	errs.PanicIfError(err)

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	h := md5.New()

	_, err = io.Copy(h, f)
	errs.PanicIfError(err)

	return hex.EncodeToString(h.Sum(nil))
}
