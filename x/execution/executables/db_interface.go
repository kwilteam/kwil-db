package executables

import (
	"fmt"
	"kwil/x/types/databases"
	execTypes "kwil/x/types/execution"
)

// ExecutablesInterface is an in-memory interface that makes retrieval and preparation of executables for applications easy.
// It contains executables and access control rules / roles.
// All databases will be held in memory (for now)

type ExecutablesInterface interface {
	CanExecute(wallet string, query string) bool
	Prepare(query string, caller string, inputs []*execTypes.UserInput) ([]any, error)
	ListExecutables() []*execTypes.Executable
}

// fulfills ExecutableInterface
type executableInterface struct {
	Owner        string
	Executables  map[string]*execTypes.Executable
	Access       map[string]map[string]struct{} // maps a role name to an executable
	DefaultRoles []string
}

func FromDatabase(db *databases.Database) (ExecutablesInterface, error) {
	execs, err := generateExecutables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to generate executables: %w", err)
	}

	return &executableInterface{
		Owner:        db.Owner,
		Executables:  execs,
		Access:       generateAccessParameters(db),
		DefaultRoles: db.GetDefaultRoles(),
	}, nil
}
