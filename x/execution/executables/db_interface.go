package executables

import (
	"fmt"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
	execTypes "kwil/x/types/execution"
)

// ExecutablesInterface is an in-memory interface that makes retrieval and preparation of executables for applications easy.
// It contains executables and access control rules / roles.
// All databases will be held in memory (for now)

type ExecutablesInterface interface {
	CanExecute(wallet string, query string) bool
	Prepare(query string, caller string, inputs []*execTypes.UserInput[anytype.KwilAny]) (string, []any, error)
	ListExecutables() []*execTypes.Executable
	GetIdentifier() *databases.DatabaseIdentifier
}

// fulfills ExecutableInterface
type executableInterface struct {
	Owner        string
	Name         string
	Executables  map[string]*execTypes.Executable
	Access       map[string]map[string]struct{} // maps a role name to an executable
	DefaultRoles []string
}

func FromDatabase(db *databases.Database[anytype.KwilAny]) (ExecutablesInterface, error) {
	execs, err := generateExecutables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to generate executables: %w", err)
	}

	return &executableInterface{
		Owner:        db.Owner,
		Name:         db.Name,
		Executables:  execs,
		Access:       generateAccessParameters(db),
		DefaultRoles: db.GetDefaultRoles(),
	}, nil
}

func (e *executableInterface) GetIdentifier() *databases.DatabaseIdentifier {
	return &databases.DatabaseIdentifier{
		Owner: e.Owner,
		Name:  e.Name,
	}
}

func (e *executableInterface) getDbId() string {
	return databases.GenerateSchemaName(e.Owner, e.Name)
}
