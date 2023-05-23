package dataset_test

import (
	"context"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	data "github.com/kwilteam/kwil-db/pkg/engine/dto/data"

	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/stretchr/testify/assert"
)

func Test_Dataset(t *testing.T) {
	ctx := context.Background()

	ds, err := dataset.NewDataset(ctx, &dto.DatasetContext{
		Name:  "testName",
		Owner: "testOwner",
	}, newMockDB())
	if err != nil {
		t.Fatal(err)
	}
	defer ds.Close()

	if ds.Id() != utils.GenerateDBID("testName", "testOwner") {
		t.Fatal("unexpected id")
	}

	err = ds.CreateTable(ctx, data.TableUsers)
	if err != nil {
		t.Fatal(err)
	}

	tableUsers := ds.GetTable(data.TableUsers.Name)

	assert.Equal(t, data.TableUsers, tableUsers)

	tableList := ds.ListTables()

	assert.Equal(t, []*dto.Table{data.TableUsers}, tableList)

	err = ds.CreateAction(ctx, data.ActionInsertUser)
	if err != nil {
		t.Fatal(err)
	}

	insertUserAction := ds.GetAction(data.ActionInsertUser.Name)

	assert.Equal(t, data.ActionInsertUser, insertUserAction)

	actionList := ds.ListActions()

	assert.Equal(t, []*dto.Action{data.ActionInsertUser}, actionList)

	// test that I cannot create a table with the same name
	err = ds.CreateTable(ctx, data.TableUsers)
	if err == nil {
		t.Fatal("expected error")
	}

	// test that I cannot create an action with the same name
	err = ds.CreateAction(ctx, data.ActionInsertUser)
	if err == nil {
		t.Fatal("expected error")
	}
}

func Test_Execution(t *testing.T) {
	ctx := context.Background()

	ds, err := dataset.NewDataset(ctx, &dto.DatasetContext{
		Name:  "testName",
		Owner: "testOwner",
	}, newMockDB())
	if err != nil {
		t.Fatal(err)
	}
	defer ds.Close()

	inputs := []map[string]any{
		{
			"age":  20,
			"name": "foo",
		},
		{
			"age":  30,
			"name": "bar",
		},
	}

	// execute non-existent action
	_, err = ds.Execute(&dto.TxContext{
		Caller: "0xbennanmode",
		Action: data.ActionInsertUser.Name,
	}, inputs)
	if err == nil {
		t.Fatal("expected error")
	}

	// create action
	err = ds.CreateAction(ctx, data.ActionInsertUser)
	if err != nil {
		t.Fatal(err)
	}

	result, err := ds.Execute(&dto.TxContext{
		Caller: "0xbennanmode",
		Action: data.ActionInsertUser.Name,
	}, inputs)
	if err != nil {
		t.Fatal(err)
	}

	records := result.Records()

	// the mock result returns 2 records
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
}
