package executables

import "kwil/x/execution/dto"

// ExecutablesInterface is an in-memory interface that makes retrieval and preparation of executables for applications easy.
// It contains executables and access control rules / roles.
// All databases will be held in memory (for now)

type ExecutablesInterface interface {
	CanExecute(wallet string, query string) bool
	Prepare(query string, caller string, inputs []*dto.UserInput) ([]any, error)
}

// fulfills DatabaseInterface
type databaseInterface struct {
	Owner        string
	Executables  map[string]*dto.Executable
	Access       map[string]string // maps a role name to an executable
	DefaultRoles []string
}

// NewDatabaseInterface creates a new DatabaseInterface
func NewDatabaseInterface(executables map[string]*dto.Executable, access map[string]string, defaultRoles []string, Owner string) ExecutablesInterface {
	return &databaseInterface{
		Executables:  executables,
		Access:       access,
		DefaultRoles: defaultRoles,
	}
}
