package common

import (
	"os"
	"path/filepath"
)

func DefaultKwilAdminRoot() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kwil-admin")
}

func DefaultKwildRoot() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".kwild")
}
