package functions

import (
	"fmt"

	"github.com/kwilteam/go-sqlite"
	errorFunc "github.com/kwilteam/go-sqlite/ext/error"
	"github.com/kwilteam/kwil-db/internal/sql/sqlite/functions/addresses"
)

// This file contains functionality to register custom functions with SQLite
func Register(c *sqlite.Conn) error {
	err := errorFunc.Register(c)
	if err != nil {
		return fmt.Errorf(`failed to register "ERROR" function: %w`, err)
	}

	err = addresses.Register(c)
	if err != nil {
		return err
	}

	return nil
}
