package datasets

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/engine/models"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/sql/driver"
	"github.com/kwilteam/kwil-db/pkg/utils/numbers/polynomial"
	"math/big"
)

type PreparedAction struct {
	Name       string
	Statements []*PreparedStatement
	Public     bool
	Inputs     []string
}

func NewPreparedAction(conn *driver.Connection, a *models.Action, tables map[string]*models.Table) (*PreparedAction, error) {
	stmts := make([]*PreparedStatement, 0)
	for _, stmt := range a.Statements {
		s, err := NewPreparedStatement(conn, stmt, tables)
		if err != nil {
			return nil, fmt.Errorf("error preparing statement %s in action %s: %w", stmt, a.Name, err)
		}
		stmts = append(stmts, s)
	}
	return &PreparedAction{
		Name:       a.Name,
		Statements: stmts,
		Public:     a.Public,
		Inputs:     a.Inputs,
	}, nil
}

func (a *PreparedAction) GetPrice() (*big.Int, error) {
	price := big.NewInt(0)
	for _, stmt := range a.Statements {
		p, err := stmt.GetPrice()
		if err != nil {
			return nil, err
		}
		price.Add(price, p)
	}
	return price, nil
}

func (p *PreparedAction) GetAction() *models.Action {
	stmts := make([]string, 0)
	for _, stmt := range p.Statements {
		stmts = append(stmts, stmt.Stmt)
	}
	return &models.Action{
		Name:       p.Name,
		Statements: stmts,
		Public:     p.Public,
		Inputs:     p.Inputs,
	}
}

func (p *PreparedAction) Prepare(exec *models.ActionExecution, opts *ExecOpts) (finalRecords []map[string]any, err error) {
	for _, record := range exec.Params {
		finalRecord, err := p.prepareSingle(record)
		if err != nil {
			return nil, fmt.Errorf("error preparing action inputs %s: %w", p.Name, err)
		}

		finalRecord[callerVar] = opts.Caller

		finalRecords = append(finalRecords, finalRecord)
	}
	if len(finalRecords) == 0 {
		finalRecords = append(finalRecords, map[string]any{callerVar: opts.Caller})
	}

	return finalRecords, nil
}

// prepareSingle prepares a single record
func (p *PreparedAction) prepareSingle(record map[string][]byte) (map[string]any, error) {
	finalRecord := make(map[string]any)
	for _, input := range p.Inputs {
		val, ok := record[input]
		if !ok {
			return nil, fmt.Errorf(`missing input "%s"`, input)
		}

		concrete, err := types.NewFromSerial(val)
		if err != nil {
			return nil, fmt.Errorf("error converting serialized input %s: %w", input, err)
		}

		if err := concrete.Ok(); err != nil {
			return nil, fmt.Errorf("invalid serialized input %s: %w", input, err)
		}

		finalRecord[input] = concrete.Value()
	}

	return finalRecord, nil
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

func NewPreparedStatement(conn *driver.Connection, stmt string, tables map[string]*models.Table) (*PreparedStatement, error) {
	// get aliases
	aliases, err := parseAliases(stmt)
	if err != nil {
		return nil, err
	}

	for alias, tableName := range aliases {
		schemaTable, ok := tables[tableName]
		if !ok {
			return nil, fmt.Errorf("table %s does not exist", tableName)
		}

		tables[alias] = schemaTable
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

	err = conn.Prepare(stmt)
	if err != nil {
		return nil, err
	}

	return &PreparedStatement{
		Stmt:             stmt,
		conn:             conn,
		Aliases:          tables,
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
	table, ok := p.Aliases[variable]
	if !ok {
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
