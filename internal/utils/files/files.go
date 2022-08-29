package files

import (
	"os"
	"path"
	"path/filepath"
	"runtime"
)

// Loads file from the root directory of Kwil
func LoadFileFromRoot(path string) ([]byte, error) {
	// MAKE SURE THIS FILE DOES NOT MOVE OR IT WILL BREAK
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	return os.ReadFile(filepath.Join(basepath, "../../", path))
}

func GetCurrentPath() string {
	_, filename, _, _ := runtime.Caller(1)

	return path.Dir(filename)
}
