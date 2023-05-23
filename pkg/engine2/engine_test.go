package engine2_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine2"
	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
	"github.com/kwilteam/kwil-db/pkg/engine2/dto/data"

	"github.com/stretchr/testify/assert"
)

func Test_Engine(t *testing.T) {
	ctx := context.Background()

	master, err := openEngine(true)
	if err != nil {
		t.Fatal(err)
	}
	defer master.Delete(true)

	ds, err := openDataset(master)
	if err != nil {
		t.Fatal(err)
	}

	sameDs, err := master.GetDataset(ds.Id())
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, ds, sameDs)

	err = master.DeleteDataset(ctx, &dto.TxContext{
		Caller: "testOwner",
	}, ds.Id())
	if err != nil {
		t.Fatal(err)
	}

	_, err = master.GetDataset(ds.Id())
	if err == nil {
		t.Fatal("expected error")
	}
}

// testing that a dataset is persisted and reloads properly
func Test_DatasetPersistence(t *testing.T) {

	master, err := openEngine(true)
	if err != nil {
		t.Fatal(err)
	}

	ds, err := openDataset(master)
	if err != nil {
		t.Error(err)
	}

	err = addUsersTable(ds)
	if err != nil {
		t.Error(err)
	}

	err = master.Close(false)
	if err != nil {
		t.Error(err)
	}

	master, err = openEngine(false)
	if err != nil {
		t.Error(err)
	}
	defer master.Delete(true)

	ds2, err := master.GetDataset(ds.Id())
	if err != nil {
		t.Error(err)
	}

	tables := ds2.ListTables()
	if len(tables) != 1 {
		t.Error("expected 1 table")
	}
}

func openEngine(wipe bool) (engine2.Engine, error) {
	ctx := context.Background()

	opts := []engine2.EngineOpt{engine2.WithName("unittest")}
	if wipe {
		opts = append(opts, engine2.WithWipe())
	}

	master, err := engine2.Open(ctx,
		opts...,
	)
	if err != nil {
		return nil, err
	}

	return master, nil
}

func openDataset(e engine2.Engine) (engine2.Dataset, error) {
	ctx := context.Background()

	ds, err := e.NewDataset(ctx, &dto.DatasetContext{
		Name:  "testName",
		Owner: "testOwner",
	})
	if err != nil {
		return nil, err
	}

	return ds, nil
}

func addUsersTable(ds engine2.Dataset) error {
	ctx := context.Background()

	return ds.CreateTable(ctx, data.TableUsers)
}
