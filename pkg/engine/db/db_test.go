package db_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/db"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/stretchr/testify/assert"
)

var (
	testTable = &types.Table{
		Name: "test_table",
		Columns: []*types.Column{
			{
				Name: "test_column",
				Type: types.TEXT,
				Attributes: []*types.Attribute{
					{
						Type: types.PRIMARY_KEY,
					},
				},
			},
		},
	}
)

func Test_ListTables(t *testing.T) {
	ctx := context.Background()

	datastore, err := db.NewDB(ctx, newMockDB())
	if err != nil {
		t.Fatal(err)
	}
	defer datastore.Close()

	tbls, err := datastore.ListTables(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(tbls) != 0 {
		t.Fatalf("expected 0 tables, got %d", len(tbls))
	}

	err = datastore.CreateTable(ctx, testTable)
	if err != nil {
		t.Fatal(err)
	}

	tbls, err = datastore.ListTables(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(tbls) != 1 {
		t.Fatalf("expected 1 table, got %d", len(tbls))
	}

	assert.Equal(t, testTable, tbls[0])
}
