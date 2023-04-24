package executables

import (
	"fmt"
	"kwil/pkg/databases"
	"kwil/pkg/databases/spec"
)

// executable is an executable query
type executable struct {
	// the query that is being executed
	Query *databases.SQLQuery[*spec.KwilAny]

	// tableName also includes the schema name, if necessary
	TableName string

	// the table is used for validating modifiers and data types
	Columns map[string]*databases.Column[*spec.KwilAny]
}

func generateExecutables(db *databases.Database[*spec.KwilAny]) (map[string]*executable, error) {
	// first create a map of tables to columns for quick lookup and shallow copy
	// of the columns
	columns := make(map[string]map[string]*databases.Column[*spec.KwilAny])
	for _, table := range db.Tables {
		columns[table.Name] = columnsToMap(table.Columns)
	}

	// then create the executables
	executables := make(map[string]*executable)
	for _, query := range db.SQLQueries {
		cols, ok := columns[query.Table]
		if !ok {
			return nil, fmt.Errorf(`table "%s" not found when generating executable "%s"`, query.Table, query.Name)
		}

		executables[query.Name] = &executable{
			Query:     query,
			TableName: makeTableName(db.GetSchemaName(), query.Table),
			Columns:   cols,
		}
	}

	return executables, nil
}

func columnsToMap(columns []*databases.Column[*spec.KwilAny]) map[string]*databases.Column[*spec.KwilAny] {
	m := make(map[string]*databases.Column[*spec.KwilAny])
	for _, column := range columns {
		m[column.Name] = column
	}
	return m
}

func makeTableName(schemaName, table string) string {
	if schemaName != "" {
		return schemaName + "." + table
	}
	return table
}

// getQuerySignature loops over parameters and where clauses and returns a list of args and the name of the query
func (e *executable) getQuerySignature() (*QuerySignature, error) {
	var args []*Arg
	for _, param := range e.Query.Params {
		if param.Static {
			continue
		}

		column, ok := e.Columns[param.Column]
		if !ok {
			// this should never happen, but if it does, we want to know about it
			// TODO: Update to new log type log.New().Error(`column not found when generating executable args from parameters`, zap.String("table", e.TableName), zap.String("column", param.Name))
			return nil, fmt.Errorf(`column "%s" not found in table "%s"`, param.Name, e.TableName)
		}

		args = append(args, &Arg{
			Name: param.Name,
			Type: column.Type,
		})
	}

	for _, where := range e.Query.Where {
		if where.Static {
			continue
		}

		column, ok := e.Columns[where.Column]
		if !ok {
			// TODO: Update to new log typelog.New().Error(`column not found when generating executable args from wheres`, zap.String("table", e.TableName), zap.String("column", where.Name))
			return nil, fmt.Errorf(`column "%s" not found in table "%s"`, where.Name, e.TableName)
		}

		args = append(args, &Arg{
			Name: where.Name,
			Type: column.Type,
		})
	}

	return &QuerySignature{
		Name: e.Query.Name,
		Args: args,
	}, nil
}

// prepare takes the inputs and caller and returns the query and args
func (e *executable) prepare(inputs []*UserInput, caller string) (string, []any, error) {
	p := e.getPreparer(inputs, caller)
	return p.Prepare()
}

func (e *executable) getPreparer(inputs []*UserInput, caller string) *preparer {
	ins := make(map[string][]byte)
	for _, input := range inputs {
		ins[input.Name] = input.Value
	}

	return &preparer{
		executable: e,
		inputs:     ins,
		caller:     caller,
	}
}
