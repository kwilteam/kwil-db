package crypto

import (
	"os"
	"path/filepath"
	"runtime"
)

// Loads file from the root directory of Kwil
func loadFileFromRoot(path string) ([]byte, error) {
	// MAKE SURE THIS FILE DOES NOT MOVE OR IT WILL BREAK
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	return os.ReadFile(filepath.Join(basepath, "../../../", path))
}
