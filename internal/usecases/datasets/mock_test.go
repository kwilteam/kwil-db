package datasets_test

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/engine"
	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	"github.com/kwilteam/kwil-db/pkg/engine/sqldb"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
)

type mockEngine struct {
	datasets      map[string]*mockDataset
	ownedDatasets map[string][]string
}

func newMockEngine() *mockEngine {
	return &mockEngine{
		datasets:      make(map[string]*mockDataset),
		ownedDatasets: make(map[string][]string),
	}
}

func (m *mockEngine) Close(closeAll bool) error {
	return nil
}

func (m *mockEngine) Delete(deleteAll bool) error {
	return nil
}

func (m *mockEngine) DeleteDataset(ctx context.Context, txCtx *dto.TxContext, dbid string) error {
	ownedDbs, ok := m.ownedDatasets[txCtx.Caller]
	if !ok {
		return fmt.Errorf("dataset not found")
	}

	for i, ownedDb := range ownedDbs {
		if ownedDb == dbid {
			m.ownedDatasets[txCtx.Caller] = append(ownedDbs[:i], ownedDbs[i+1:]...)
			break
		}
	}

	delete(m.datasets, dbid)

	return nil
}

func (m *mockEngine) GetDataset(dbid string) (engine.Dataset, error) {
	ds, ok := m.datasets[dbid]
	if !ok {
		return nil, fmt.Errorf("dataset not found")
	}

	return ds, nil
}

func (m *mockEngine) ListDatasets(ctx context.Context, owner string) ([]string, error) {
	return m.ownedDatasets[owner], nil
}

func (m *mockEngine) NewDataset(ctx context.Context, dsCtx *dto.DatasetContext) (engine.Dataset, error) {
	mockDs := &mockDataset{}

	id := makeDbid(dsCtx.Name, dsCtx.Owner)

	m.datasets[id] = mockDs
	m.ownedDatasets[dsCtx.Owner] = append(m.ownedDatasets[dsCtx.Owner], id)

	return newMockDataset(dsCtx.Owner, dsCtx.Name), nil
}

func makeDbid(name, owner string) string {
	return utils.GenerateDBID(name, owner)
}

type mockDataset struct {
	owner   string
	name    string
	actions map[string]*dto.Action
	tables  map[string]*dto.Table
}

func newMockDataset(owner, name string) *mockDataset {
	return &mockDataset{
		owner:   owner,
		name:    name,
		actions: make(map[string]*dto.Action),
		tables:  make(map[string]*dto.Table),
	}
}

func (m *mockDataset) CreateAction(ctx context.Context, action *dto.Action) error {
	name := strings.ToLower(action.Name)

	if _, ok := m.actions[name]; ok {
		return fmt.Errorf("action already exists")
	}

	m.actions[name] = action

	return nil
}

func (m *mockDataset) CreateTable(ctx context.Context, table *dto.Table) error {
	name := strings.ToLower(table.Name)

	if _, ok := m.tables[name]; ok {
		return fmt.Errorf("table already exists")
	}

	m.tables[name] = table

	return nil
}

func (m *mockDataset) Execute(txCtx *dto.TxContext, inputs []map[string]any) (dto.Result, error) {
	return &mockResult{}, nil
}

func (m *mockDataset) Id() string {
	return makeDbid(m.name, m.owner)
}

func (m *mockDataset) ListActions() []*dto.Action {
	actionList := make([]*dto.Action, 0)
	for _, action := range m.actions {
		actionList = append(actionList, action)
	}

	return actionList
}

func (m *mockDataset) ListTables() []*dto.Table {
	tableList := make([]*dto.Table, 0)
	for _, table := range m.tables {
		tableList = append(tableList, table)
	}

	return tableList
}

func (m *mockDataset) Name() string {
	return m.name
}

func (m *mockDataset) Owner() string {
	return m.owner
}

func (m *mockDataset) Query(ctx context.Context, stmt string, args map[string]any) (dto.Result, error) {
	return &mockResult{}, nil
}

func (m *mockDataset) Savepoint() (sqldb.Savepoint, error) {
	return &mockSavepoint{}, nil
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

type mockAccountStore struct{}

func (m *mockAccountStore) BatchCredit(creditList []*balances.Credit, chain *balances.ChainConfig) error {
	return nil
}

func (m *mockAccountStore) Close() error {
	return nil
}

func (m *mockAccountStore) GetAccount(address string) (*balances.Account, error) {
	bal, ok := new(big.Int).SetString("10000000000000000000000", 10)
	if !ok {
		return nil, fmt.Errorf("error parsing balance")
	}

	return &balances.Account{
		Address: address,
		Balance: bal,
	}, nil
}

func (m *mockAccountStore) Spend(spend *balances.Spend) error {
	return nil
}
