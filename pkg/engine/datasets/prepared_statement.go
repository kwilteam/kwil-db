package datasets

import (
	"fmt"
	"kwil/pkg/engine/models"
	"kwil/pkg/sql/driver"
	"kwil/pkg/utils/numbers/polynomial"
	"math/big"
)

// TODO: this is a temporary interface until we have a proper parser
type aliasParser interface {
	ParseAliases(stmt string) (map[string]string, error)
}

type PreparedStatement struct {
	Stmt string
	conn *driver.Connection
	// aliases is a map of alias to table name
	// it also maps table names to themselves
	Aliases          map[string]*models.Table
	polynomial       polynomial.Expression
	requiredPolyVars []string
}

func NewPreparedStatement(conn *driver.Connection, stmt string, schema *models.Database) (*PreparedStatement, error) {
	// get aliases
	var ap aliasParser
	aliases, err := ap.ParseAliases(stmt)
	if err != nil {
		return nil, err
	}

	tableNameMapping := schema.GetTableMapping()
	for alias, tableName := range aliases {
		schemaTable := schema.GetTable(tableName)
		if schemaTable == nil {
			return nil, fmt.Errorf("table %s does not exist", tableName)
		}

		tableNameMapping[alias] = schemaTable
	}

	// plan query and get polynomial
	plan, err := conn.Plan(stmt)
	if err != nil {
		return nil, err
	}

	poly := plan.Polynomial()
	if poly == nil {
		return nil, fmt.Errorf(`polynomial could not be generated for query "%s"`, stmt)
	}

	polyVars := make([]string, 0)
	for variable := range poly.Variables() {
		polyVars = append(polyVars, variable)
	}

	return &PreparedStatement{
		Stmt:             stmt,
		conn:             conn,
		Aliases:          tableNameMapping,
		polynomial:       poly,
		requiredPolyVars: polyVars,
	}, nil

}

func (p *PreparedStatement) GetPrice() (*big.Int, error) {
	variables, err := p.getPolyVarValues()
	if err != nil {
		return nil, err
	}

	res, err := p.polynomial.Evaluate(variables)
	if err != nil {
		return nil, err
	}

	bigInt := new(big.Int)
	res.Int(bigInt)

	return bigInt, nil
}

func (p *PreparedStatement) getPolyVarValues() (map[string]*big.Float, error) {
	variables := make(map[string]*big.Float)
	for _, variable := range p.requiredPolyVars {
		rowCount, err := p.getRowCount(variable)
		if err != nil {
			return nil, err
		}

		variables[variable] = polynomial.NewFloatFromInt(rowCount)
	}

	return variables, nil
}

// getRowCount returns the number of rows for the table specified by the variable.
// If the variable is not a table name, it returns 1.
func (p *PreparedStatement) getRowCount(variable string) (int64, error) {
	table := p.Aliases[variable]
	if table == nil {
		return 1, nil
	}

	var rowCount int64
	err := p.conn.Query(fmt.Sprintf("SELECT COUNT(rowid) as row_count FROM %s;", table.Name), func(stmt *driver.Statement) error {
		rowCount = stmt.GetInt64("row_count")
		return nil
	})
	if err != nil {
		return 0, err
	}

	return rowCount, nil
}
