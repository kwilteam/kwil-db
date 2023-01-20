package executables

import (
	"fmt"
	"kwil/x/execution"
	"kwil/x/execution/sql-builder/dml"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
	execTypes "kwil/x/types/execution"
)

func generateExecutables(db *databases.Database[anytype.KwilAny]) (map[string]*execTypes.Executable, error) {
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

func generateExecutable(db *databases.Database[anytype.KwilAny], q *databases.SQLQuery) (*execTypes.Executable, error) {
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
		Args:       args,
		UserInputs: buildInputs(args),
	}, nil
}

func generateStatement(db *databases.Database[anytype.KwilAny], q *databases.SQLQuery) (string, error) {
	switch q.Type {
	case execution.SELECT:
		return "", fmt.Errorf("SELECT is not supported yet")
	case execution.INSERT:
		return dml.BuildInsert(db.GetSchemaName(), q.Table, q.ListParamColumnsAsAny())
	case execution.UPDATE:
		return dml.BuildUpdate(db.GetSchemaName(), q.Table, q.Params, q.Where)
	case execution.DELETE:
		return dml.BuildDelete(db.GetSchemaName(), q.Table, q.Where)
	}
	return "", fmt.Errorf("unknown query type")
}

// an interface for building Args from params and where clauses
type arger interface {
	GetName() string
	GetColumn() string
	GetModifier() execution.ModifierType
	GetStatic() bool
	GetValue() any
}

// buildsArgs will build all args and determine their position
func buildArgs(db *databases.Database[anytype.KwilAny], q *databases.SQLQuery) ([]*execTypes.Arg, error) {
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
func buildArg(db *databases.Database[anytype.KwilAny], q *databases.SQLQuery, position uint8, param arger) (*execTypes.Arg, error) {
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
		Value:    param.GetValue(),
		Type:     col.Type,
		Name:     param.GetName(),
		Modifier: param.GetModifier(),
	}, nil
}

// build inputs will identify the non-static args and build the user inputs
func buildInputs(args []*execTypes.Arg) []*execTypes.UserInput {
	inputs := []*execTypes.UserInput{}
	for _, arg := range args {
		if !arg.Static {
			inputs = append(inputs, &execTypes.UserInput{
				Name:  arg.Name,
				Value: "",
			})
		}
	}
	return inputs
}
