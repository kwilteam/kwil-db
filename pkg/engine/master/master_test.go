package master_test

import (
	"context"
	"testing"

	sqlTesting "github.com/kwilteam/kwil-db/pkg/sql/testing"

	"github.com/kwilteam/kwil-db/pkg/engine/master"
)

type testCase func(*testing.T, *master.MasterDB)

func Test_Master(t *testing.T) {

	tests := []struct {
		name string
		test testCase
	}{
		{
			name: "create dataset",
			test: func(t *testing.T, m *master.MasterDB) {
				ctx := context.Background()

				err := m.RegisterDataset(ctx, "testName", "testOwner")
				if err != nil {
					t.Error(err)
				}

				datasets, err := m.ListDatasets(ctx)
				if err != nil {
					t.Error(err)
				}

				if len(datasets) != 1 {
					t.Errorf("expected 1 dataset, got %d", len(datasets))
				}
			},
		},
		{
			name: "create datasets with same name and owner",
			test: func(t *testing.T, m *master.MasterDB) {
				ctx := context.Background()

				err := m.RegisterDataset(ctx, "testName", "testOwner")
				if err != nil {
					t.Error(err)
				}

				err = m.RegisterDataset(ctx, "testName", "testOwner")
				if err == nil {
					t.Error("expected database to return error when creating dataset with same name and owner")
				}

				datasets, err := m.ListDatasets(ctx)
				if err != nil {
					t.Error(err)
				}

				if len(datasets) != 1 {
					t.Errorf("expected 1 dataset, got %d", len(datasets))
				}
			},
		},
		{
			name: "create datasets with same name and same owner",
			test: func(t *testing.T, m *master.MasterDB) {
				ctx := context.Background()

				err := m.RegisterDataset(ctx, "testName", "testOwner")
				if err != nil {
					t.Error(err)
				}

				err = m.RegisterDataset(ctx, "testName", "testOwner")
				if err == nil {
					t.Error("expected database to return error when creating dataset with same name and owner")
				}

				datasets, err := m.ListDatasets(ctx)
				if err != nil {
					t.Error(err)
				}

				if len(datasets) != 1 {
					t.Errorf("expected 1 dataset, got %d", len(datasets))
				}
			},
		},
		{
			name: "testing lsiting dataseets by owner",
			test: func(t *testing.T, m *master.MasterDB) {
				ctx := context.Background()

				err := m.RegisterDataset(ctx, "testName", "testOwner")
				if err != nil {
					t.Error(err)
				}

				err = m.RegisterDataset(ctx, "testName2", "testOwner")
				if err != nil {
					t.Error(err)
				}

				err = m.RegisterDataset(ctx, "testName3", "testOwner2")
				if err != nil {
					t.Error(err)
				}

				datasets, err := m.ListDatasetsByOwner(ctx, "testOwner")
				if err != nil {
					t.Error(err)
				}

				if len(datasets) != 2 {
					t.Errorf("expected 2 datasets, got %d", len(datasets))
				}
			},
		},
		{
			name: "test unregister dataset",
			test: func(t *testing.T, m *master.MasterDB) {
				ctx := context.Background()

				err := m.RegisterDataset(ctx, "testName", "testOwner")
				if err != nil {
					t.Error(err)
				}

				err = m.UnregisterDataset(ctx, m.DbidFunc("testName", "testOwner"))
				if err != nil {
					t.Error(err)
				}

				datasets, err := m.ListDatasets(ctx)
				if err != nil {
					t.Error(err)
				}

				if len(datasets) != 0 {
					t.Errorf("expected 0 datasets, got %d", len(datasets))
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			db, td, err := sqlTesting.OpenTestDB("test")
			if err != nil {
				t.Fatal(err)
			}
			defer td()

			datastore, err := master.New(ctx, &databaseAdapter{db})
			if err != nil {
				t.Fatal(err)
			}
			defer datastore.Close()

			test.test(t, datastore)
		})
	}
}

type databaseAdapter struct {
	sqlTesting.TestSqliteClient
}

func (d *databaseAdapter) Savepoint() (master.Savepoint, error) {
	return d.TestSqliteClient.Savepoint()
}
