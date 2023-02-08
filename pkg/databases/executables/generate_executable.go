package executables

import (
	"fmt"
	execution2 "kwil/pkg/databases"
	dml2 "kwil/pkg/databases/sql-builder/dml"
	"kwil/pkg/types/data_types/any_type"
	execTypes "kwil/pkg/types/execution"
)

func generateExecutables(db *execution2.Database[anytype.KwilAny]) (map[string]*execTypes.Executable, error) {
	execs := make(map[string]*execTypes.Executable)
	for _, q := range db.SQLQueries {
		e, err := generateExecutable(db, q)
		if err != nil {
			return nil, fmt.Errorf("failed to generate executable: %w", err)
		}

		execs[e.Name] = e
	}

	return execs, nil
}

func generateExecutable(db *execution2.Database[anytype.KwilAny], q *execution2.SQLQuery[anytype.KwilAny]) (*execTypes.Executable, error) {
	statement, err := generateStatement(db, q)
	if err != nil {
		return nil, fmt.Errorf("failed to generate statement: %w", err)
	}

	args, err := buildArgs(db, q)
	if err != nil {
		return nil, fmt.Errorf("failed to build args: %w", err)
	}

	return &execTypes.Executable{
		Name:       q.Name,
		Statement:  statement,
		Table:      q.Table,
		Type:       q.Type,
		Args:       args,
		UserInputs: buildInputs(args),
		Parameters: q.Params,
		Where:      q.Where,
	}, nil
}

func generateStatement(db *execution2.Database[anytype.KwilAny], q *execution2.SQLQuery[anytype.KwilAny]) (string, error) {
	switch q.Type {
	case execution2.SELECT:
		return "", fmt.Errorf("SELECT is not supported yet")
	case execution2.INSERT:
		return dml2.BuildInsert(db.GetSchemaName(), q.Table, q.ListParamColumnsAsAny())
	case execution2.UPDATE:
		return dml2.BuildUpdate(db.GetSchemaName(), q.Table, q.Params, q.Where)
	case execution2.DELETE:
		return dml2.BuildDelete(db.GetSchemaName(), q.Table, q.Where)
	}
	return "", fmt.Errorf("unknown query type")
}

// an interface for building Args from params and where clauses
type arger interface {
	GetName() string
	GetColumn() string
	GetModifier() execution2.ModifierType
	GetStatic() bool
	GetValue() *anytype.KwilAny
}

// buildsArgs will build all args and determine their position
func buildArgs(db *execution2.Database[anytype.KwilAny], q *execution2.SQLQuery[anytype.KwilAny]) ([]*execTypes.Arg, error) {
	args := []*execTypes.Arg{}
	var pos uint8 = 0
	for _, param := range q.Params {
		arg, err := buildArg(db, q, pos, param)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		pos++
	}

	for _, where := range q.Where {
		arg, err := buildArg(db, q, pos, where)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
		pos++
	}

	return args, nil
}

// buildArg will build an arg from a param or where clause
func buildArg(db *execution2.Database[anytype.KwilAny], q *execution2.SQLQuery[anytype.KwilAny], position uint8, param arger) (*execTypes.Arg, error) {
	tbl := db.GetTable(q.Table)
	if tbl == nil {
		return nil, fmt.Errorf(`table "%s" does not exist`, q.Table)
	}

	col := tbl.GetColumn(param.GetColumn())
	if col == nil {
		return nil, fmt.Errorf(`column "%s" does not exist on table "%s"`, param.GetColumn(), q.Table)
	}

	return &execTypes.Arg{
		Position: position,
		Static:   param.GetStatic(),
		Value:    param.GetValue().Value(), // this should maybe be the bytes instead?
		Type:     col.Type,
		Name:     param.GetName(),
		Modifier: param.GetModifier(),
	}, nil
}

// build inputs will identify the non-static args and build the user inputs
func buildInputs(args []*execTypes.Arg) []*execTypes.UserInput[[]byte] {
	inputs := []*execTypes.UserInput[[]byte]{}
	for _, arg := range args {
		if !arg.Static {
			inputs = append(inputs, &execTypes.UserInput[[]byte]{
				Name:  arg.Name,
				Value: nil,
			})
		}
	}
	return inputs
}
