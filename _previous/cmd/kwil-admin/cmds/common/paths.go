package common

import (
	"os"
	"path/filepath"
	"strings"
)

func DefaultKwilAdminRoot() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kwil-admin")
}

func DefaultKwildRoot() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kwild")
}

func ExpandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}
	return filepath.Abs(path)
}
