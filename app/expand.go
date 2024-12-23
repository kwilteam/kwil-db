package app

import (
	"errors"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// ExpandPath expands paths that start with ~ or ~user.
// It handles both ~/ and ~user/ correctly and is platform-agnostic.
func ExpandPath(path string) (string, error) {
	if path == "" {
		return "", errors.New("path cannot be empty")
	}

	if !strings.HasPrefix(path, "~") {
		return filepath.Abs(path)
	}

	// Split path into potential user and remaining path
	parts := strings.SplitN(path, string(os.PathSeparator), 2)
	prefix := parts[0] // e.g., "~user" or "~"

	var homeDir string
	var err error

	if prefix == "~" {
		homeDir, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	} else if len(prefix) > 1 && prefix[1] != '/' && prefix[1] != '\\' {
		userName := prefix[1:]
		usr, err := user.Lookup(userName)
		if err != nil {
			return "", err
		}
		homeDir = usr.HomeDir
	} else {
		return "", errors.New("invalid path prefix")
	}

	if len(parts) > 1 {
		return filepath.Abs(filepath.Join(homeDir, parts[1]))
	}
	return filepath.Abs(homeDir)
}
