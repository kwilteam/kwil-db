package executables

import (
	"fmt"
	"kwil/pkg/databases"
	"kwil/pkg/databases/spec"
)

// QuerySignature is the name and arguments of a query
type QuerySignature struct {
	Name string `json:"name" yaml:"name"`
	Args []*Arg `json:"args" yaml:"args"`
}

// Arg is an argument for a query (either a parameter or where predicate)
type Arg struct {
	Name string        `json:"name" yaml:"name"`
	Type spec.DataType `json:"type" yaml:"type"`
}

// DatabaseInterface provides metadata about a database and allows for the execution of queries
type DatabaseInterface struct {
	Owner        string
	Name         string
	queries      map[string]*executable
	access       map[string]map[string]struct{} // maps a role name to an executable
	defaultRoles []string
}

// FromDatabase creates a new DatabaseInterface from a database
func FromDatabase(db *databases.Database[*spec.KwilAny]) (*DatabaseInterface, error) {
	execs, err := generateExecutables(db)
	if err != nil {
		return nil, fmt.Errorf("failed to generate executables: %w", err)
	}

	return &DatabaseInterface{
		Owner:        db.Owner,
		Name:         db.Name,
		queries:      execs,
		access:       generateAccessParameters(db),
		defaultRoles: db.GetDefaultRoles(),
	}, nil
}

func (e *DatabaseInterface) GetDbId() string {
	return databases.GenerateSchemaId(e.Owner, e.Name)
}

func (e *DatabaseInterface) ListQueries() ([]*QuerySignature, error) {
	var execs []*QuerySignature
	for _, q := range e.queries {
		exec, err := q.getQuerySignature()
		if err != nil {
			return nil, fmt.Errorf("failed to get args for executable %s: %w", q.Query.Name, err)
		}
		execs = append(execs, exec)
	}

	return execs, nil
}

func (e *DatabaseInterface) Prepare(query string, caller string, inputs []*UserInput) (string, []any, error) {
	exec, ok := e.queries[query]
	if !ok {
		return "", nil, fmt.Errorf("query %s not found", query)
	}

	return exec.prepare(inputs, caller)
}

// ConvertInputs takes a map of inputs passed as strings and tries to convert them to the correct type.
// If successful, it returns a map of the inputs converted to the correct type.
// If not, or if an input is missing, it returns an error.
func (q *QuerySignature) ConvertInputs(inputs map[string]string) (map[string]*spec.KwilAny, error) {
	args := make(map[string]*spec.KwilAny)
	for _, arg := range q.Args {
		val, ok := inputs[arg.Name]
		if !ok {
			return nil, fmt.Errorf("missing input %s", arg.Name)
		}

		kwilAny, err := spec.NewExplicit(val, arg.Type)
		if err != nil {
			return nil, fmt.Errorf("error creating kwil any type with executable inputs: %w", err)
		}

		args[arg.Name] = kwilAny
	}

	return args, nil
}
