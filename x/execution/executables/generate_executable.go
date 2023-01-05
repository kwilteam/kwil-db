package executables

import (
	"fmt"
	"kwil/x/execution"
	"kwil/x/execution/dto"
	"kwil/x/execution/sql-builder/dml"
)

func GenerateExecutables(db *dto.Database) (map[string]*dto.Executable, error) {
	execs := make(map[string]*dto.Executable)
	for _, q := range db.SQLQueries {
		e, err := GenerateExecutable(db, q)
		if err != nil {
			return nil, fmt.Errorf("failed to generate executable: %w", err)
		}

		execs[e.Name] = e
	}

	return execs, nil
}

func GenerateExecutable(db *dto.Database, q *dto.SQLQuery) (*dto.Executable, error) {
	statement, err := generateStatement(db, q)
	if err != nil {
		return nil, fmt.Errorf("failed to generate statement: %w", err)
	}

	args, err := buildArgs(db, q)
	if err != nil {
		return nil, fmt.Errorf("failed to build args: %w", err)
	}

	return &dto.Executable{
		Name:       q.Name,
		Statement:  statement,
		Args:       args,
		UserInputs: buildInputs(args),
	}, nil
}

func generateStatement(db *dto.Database, q *dto.SQLQuery) (string, error) {
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
func buildArgs(db *dto.Database, q *dto.SQLQuery) ([]*dto.Arg, error) {
	args := []*dto.Arg{}
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
func buildArg(db *dto.Database, q *dto.SQLQuery, position uint8, param arger) (*dto.Arg, error) {
	tbl := db.GetTable(q.Table)
	if tbl == nil {
		return nil, fmt.Errorf(`table "%s" does not exist`, q.Table)
	}

	col := tbl.GetColumn(param.GetColumn())
	if col == nil {
		return nil, fmt.Errorf(`column "%s" does not exist on table "%s"`, param.GetColumn(), q.Table)
	}

	return &dto.Arg{
		Position: position,
		Static:   param.GetStatic(),
		Value:    param.GetValue(),
		Type:     col.Type,
		Name:     param.GetName(),
		Modifier: param.GetModifier(),
	}, nil
}

// build inputs will identify the non-static args and build the user inputs
func buildInputs(args []*dto.Arg) []*dto.UserInput {
	inputs := []*dto.UserInput{}
	for _, arg := range args {
		if !arg.Static {
			inputs = append(inputs, &dto.UserInput{
				Name:  arg.Name,
				Value: "",
			})
		}
	}
	return inputs
}
