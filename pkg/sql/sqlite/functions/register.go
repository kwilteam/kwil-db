package functions

import (
	"fmt"

	"github.com/kwilteam/go-sqlite"
	errorFunc "github.com/kwilteam/go-sqlite/ext/error"
)

// This file contains functionality to register custom functions with SQLite

func Register(c *sqlite.Conn) error {
	err := errorFunc.Register(c)
	if err != nil {
		return fmt.Errorf(`failed to register "ERROR" function: %w`, err)
	}

	return nil
}
