package executables

import (
	"fmt"
	"kwil/x/execution/dto"
)

// ExecutablesInterface is an in-memory interface that makes retrieval and preparation of executables for applications easy.
// It contains executables and access control rules / roles.
// All databases will be held in memory (for now)

type ExecutablesInterface interface {
	CanExecute(wallet string, query string) bool
	Prepare(query string, caller string, inputs []*dto.UserInput) ([]any, error)
	ListExecutables() []*dto.Executable
}

// fulfills ExecutableInterface
type executableInterface struct {
	Owner        string
	Executables  map[string]*dto.Executable
	Access       map[string]map[string]struct{} // maps a role name to an executable
	DefaultRoles []string
}

func FromDatabase(db *dto.Database) (ExecutablesInterface, error) {
	execs, err := GenerateExecutables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to generate executables: %w", err)
	}

	return &executableInterface{
		Owner:        db.Owner,
		Executables:  execs,
		Access:       GenerateAccessParameters(db),
		DefaultRoles: db.GetDefaultRoles(),
	}, nil
}
