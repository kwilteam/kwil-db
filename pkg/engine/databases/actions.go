package databases

import (
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

// An Action is a single action that can be executed on a database.
type Action struct {
	Name           string
	Public         bool
	RequiredInputs []string
	stmt           *sqlite.Statement
	sqliteString   string
}

const (
	defaultCallerAddress = "0x0000000000000000000000000000000000000000"
)

// ExecOpts are options for executing an action.
// Things like caller, block height, etc. are included here.
type ExecOpts struct {
	// Caller is the wallet address of the caller.
	Caller string
}

// fillDefaults fills in default values for the options.
func (e *ExecOpts) fillDefaults() {
	if e == nil {
		e = &ExecOpts{}
	}

	if e.Caller == "" {
		e.Caller = defaultCallerAddress
	}
}

// Execute executes the action.
// It takes in a map of inputs and options.
// It returns a result set and an error.
func (a *Action) Execute(inputs map[string]any, opts *ExecOpts) (*sqlite.ResultSet, error) {
	opts.fillDefaults()
	panic("implement me")
	return nil, nil
}
