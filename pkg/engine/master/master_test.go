package master_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/auth"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	sqlTesting "github.com/kwilteam/kwil-db/pkg/sql/testing"

	"github.com/kwilteam/kwil-db/pkg/engine/master"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

const testPrivateKey = "4a3142b366011d28c2a3ca33a678ff753c978c685178d4168bad4474ea480cc9"
const testPrivateKey2 = "99057f8ad7ba7fcd39cadff4affbf9e07880f3b885905f2c3ad47a1768ef3429"

type testCase func(*testing.T, *master.MasterDB)

func Test_Master(t *testing.T) {

	pk, err := crypto.Secp256k1PrivateKeyFromHex(testPrivateKey)
	if err != nil {
		t.Fatal(err)
	}

	ident := &types.User{
		PublicKey: pk.PubKey().Bytes(),
		AuthType:  auth.EthPersonalSignAuth,
	}

	pk2, err := crypto.Secp256k1PrivateKeyFromHex(testPrivateKey2)
	if err != nil {
		t.Fatal(err)
	}

	ident2 := &types.User{
		PublicKey: pk2.PubKey().Bytes(),
		AuthType:  auth.EthPersonalSignAuth,
	}

	tests := []struct {
		name string
		test testCase
	}{
		{
			name: "create dataset",
			test: func(t *testing.T, m *master.MasterDB) {
				ctx := context.Background()

				err := m.RegisterDataset(ctx, "testName", ident)
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

				err := m.RegisterDataset(ctx, "testName", ident)
				if err != nil {
					t.Error(err)
				}

				err = m.RegisterDataset(ctx, "testName", ident)
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

				err := m.RegisterDataset(ctx, "testName", ident)
				if err != nil {
					t.Error(err)
				}

				err = m.RegisterDataset(ctx, "testName", ident)
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
			name: "testing listing datasets by owner",
			test: func(t *testing.T, m *master.MasterDB) {
				ctx := context.Background()

				err := m.RegisterDataset(ctx, "testName", ident)
				if err != nil {
					t.Error(err)
				}

				err = m.RegisterDataset(ctx, "testName2", ident)
				if err != nil {
					t.Error(err)
				}

				err = m.RegisterDataset(ctx, "testName3", ident2)
				if err != nil {
					t.Error(err)
				}

				datasets, err := m.ListDatasetsByOwner(ctx, ident.PublicKey)
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

				err := m.RegisterDataset(ctx, "testName", ident)
				if err != nil {
					t.Error(err)
				}

				err = m.UnregisterDataset(ctx, m.DbidFunc("testName", ident.PublicKey))
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

			datastore, err := master.New(ctx, db)
			if err != nil {
				t.Fatal(err)
			}
			defer datastore.Close()

			test.test(t, datastore)
		})
	}
}
