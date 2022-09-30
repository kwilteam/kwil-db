package utils

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"kwil/x/utils"
	"os"
	"path"
	"path/filepath"
	"runtime"
)

// Loads file from the root directory of Kwil
func LoadFileFromRoot(path string) ([]byte, error) {
	// MAKE SURE THIS FILE DOES NOT MOVE OR IT WILL BREAK
	// basepath := u.GetGoFilePathOfCallerParent()
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)

	return os.ReadFile(filepath.Join(basepath, "../../../", path))
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
	utils.PanicIfError(err)

	defer func(f *os.File) {
		_ = f.Close()
	}(f)

	h := md5.New()

	_, err = io.Copy(h, f)
	utils.PanicIfError(err)

	return hex.EncodeToString(h.Sum(nil))
}
