package executables

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/databases/spec"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
)

// Prepare is the main entry point for preparing a statement and its arguments.
// Despite preparer being privatem, Prepare is public to show that it is the
// entry point for preparing a statement.
func (p *preparer) Prepare() (string, []any, error) {
	switch p.executable.Query.Type {
	case spec.INSERT:
		return p.prepareInsert()
	case spec.UPDATE:
		return p.prepareUpdate()
	case spec.DELETE:
		return p.prepareDelete()
	}

	return "", nil, fmt.Errorf("unknown query type: %d", p.executable.Query.Type)
}

func (p *preparer) PrepareCountAll() (string, []any, error) {
	return p.prepareGetCountAll()
}

func (p *preparer) PrepareCountUpdated() (string, []any, error) {
	if p.executable.Query.Type != spec.INSERT {
		return p.prepareGetCountUpdated()
	}
	return "", nil, fmt.Errorf("unknown query type: %d", p.executable.Query.Type)
}

func (p *preparer) prepareInsert() (string, []any, error) {
	record, err := p.getRecords()
	if err != nil {
		return "", nil, err
	}

	return goqu.Dialect("postgres").Insert(p.executable.TableName).Prepared(true).Rows(record.asGoqu()).ToSQL()
}

func (p *preparer) prepareUpdate() (string, []any, error) {
	record, err := p.getRecords()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get records: %w", err)
	}

	wheres, err := p.getWhereExpression()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get where expression: %w", err)
	}

	goquWheres, err := wheres.asGoqu()
	if err != nil {
		// should never happen
		return "", nil, fmt.Errorf("failed to convert where expression to goqu: %w", err)
	}

	return goqu.Dialect("postgres").Update(p.executable.TableName).Prepared(true).Set(record).Where(goquWheres...).ToSQL()
}

func (p *preparer) prepareDelete() (string, []any, error) {
	wheres, err := p.getWhereExpression()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get where expression: %w", err)
	}

	goquWheres, err := wheres.asGoqu()
	if err != nil {
		// should never happen
		return "", nil, fmt.Errorf("failed to convert where expression to goqu: %w", err)
	}

	return goqu.Dialect("postgres").Delete(p.executable.TableName).Prepared(true).Where(goquWheres...).ToSQL()
}

func (p *preparer) prepareGetCountAll() (string, []any, error) {
	return goqu.Dialect("postgres").From(p.executable.TableName).Select(goqu.COUNT("*")).ToSQL()
}

func (p *preparer) prepareGetCountUpdated() (string, []any, error) {

	wheres, err := p.getWhereExpression()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get where expression: %w", err)
	}

	goquWheres, err := wheres.asGoqu()
	if err != nil {
		// should never happen
		return "", nil, fmt.Errorf("failed to convert where expression to goqu: %w", err)
	}

	return goqu.Dialect("postgres").From(p.executable.TableName).Select(goqu.COUNT("*")).Prepared(true).Where(goquWheres...).ToSQL()

}
