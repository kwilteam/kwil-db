package driver

import (
	"os"
	"path/filepath"
)

func createDirIfNeeded(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, os.ModePerm)
}
