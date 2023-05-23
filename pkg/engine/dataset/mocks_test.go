package dataset_test

import (
	"context"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	"github.com/kwilteam/kwil-db/pkg/engine/sqldb"
)

type mockDb struct {
	tables  []*dto.Table
	actions []*dto.Action
}

func newMockDB() *mockDb {
	return &mockDb{
		tables:  make([]*dto.Table, 0),
		actions: make([]*dto.Action, 0),
	}
}

func (m *mockDb) Close() error {
	return nil
}

func (m *mockDb) CreateTable(ctx context.Context, table *dto.Table) error {
	m.tables = append(m.tables, table)
	return nil
}

func (m *mockDb) Delete() error {
	return nil
}

func (m *mockDb) ListActions(ctx context.Context) ([]*dto.Action, error) {
	return m.actions, nil
}

func (m *mockDb) ListTables(ctx context.Context) ([]*dto.Table, error) {
	return m.tables, nil
}

func (m *mockDb) Prepare(query string) (sqldb.Statement, error) {
	return &mockStatement{}, nil
}

func (m *mockDb) Query(ctx context.Context, query string, args map[string]any) (dto.Result, error) {
	return &mockResult{}, nil
}

func (m *mockDb) Savepoint() (sqldb.Savepoint, error) {
	return &mockSavepoint{}, nil
}

func (m *mockDb) StoreAction(ctx context.Context, action *dto.Action) error {
	m.actions = append(m.actions, action)
	return nil
}

type mockStatement struct {
}

func (m *mockStatement) Close() error {
	return nil
}

func (m *mockStatement) Execute(args map[string]any) (dto.Result, error) {
	return &mockResult{}, nil
}

type mockResult struct {
}

func (m *mockResult) Records() []map[string]any {
	return []map[string]any{
		{
			"foo": "bar",
			"bar": "baz",
		},
		{
			"wub": "bub",
			"wab": "bab",
		},
	}
}

type mockSavepoint struct {
}

func (m *mockSavepoint) Commit() error {
	return nil
}

func (m *mockSavepoint) Rollback() error {
	return nil
}
