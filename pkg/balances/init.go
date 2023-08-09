package balances

import (
	"fmt"
	"os"
)

// this file contains sql initialization

var defaultPath string

func init() {
	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "/tmp"
	}

	defaultPath = fmt.Sprintf("%s/.kwil/sqlite/", dirname)
}

const (
	accountDBName = "accounts_db"
)
