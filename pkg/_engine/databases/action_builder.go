package databases

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

type ActionBuilder struct {
	action *Action
	stmts  []string //separating this so that we can prepare all statements at once
}

func actionBuilder(ds *Dataset) *ActionBuilder {
	return &ActionBuilder{
		action: &Action{
			dataset: ds,
			stmts:   make([]*sqlite.Statement, 0),
		},
	}
}

func (a *ActionBuilder) WithName(name string) *ActionBuilder {
	a.action.Name = name
	return a
}

func (a *ActionBuilder) WithStatements(stmts []string) *ActionBuilder {
	a.stmts = stmts
	return a
}

func (a *ActionBuilder) WithPublicity(public bool) *ActionBuilder {
	a.action.Public = public
	return a
}

func (a *ActionBuilder) WithInputs(inputs []string) *ActionBuilder {
	a.action.RequiredInputs = inputs
	return a
}

func (a *ActionBuilder) Build() (*Action, error) {
	if a.action.dataset == nil {
		return nil, fmt.Errorf("dataset is nil")
	}

	if a.action.dataset.actions[strings.ToLower(a.action.Name)] != nil {
		return nil, fmt.Errorf(`action "%s" already exists`, a.action.Name)
	}

	for _, stmt := range a.stmts {
		parsedStmt, err := parser.Parse(stmt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sql: %w", err)
		}

		sqliteStmtString, err := parsedStmt.ToSQL()
		if err != nil {
			return nil, fmt.Errorf("invalid statement: %w", err)
		}

		stmt, err := a.action.dataset.conn.Prepare(sqliteStmtString)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statement: %w", err)
		}

		a.action.stmts = append(a.action.stmts, stmt)

	}

	a.action.dataset.actions[strings.ToLower(a.action.Name)] = a.action

	return a.action, nil
}
