package engine_test

import (
	"kwil/pkg/engine"
	"kwil/pkg/engine/datasets"
	"kwil/pkg/engine/models"
	"kwil/pkg/engine/models/mocks"
	"kwil/pkg/engine/types"
	"testing"
)

func Test_Engine_Deploy(t *testing.T) {
	master, err := engine.Open()
	if err != nil {
		t.Fatal(err)
	}
	defer master.Close()

	db, err := wipeAndDeploy(t, master, &mocks.MOCK_DATASET1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = db.ExecuteAction(&models.ActionExecution{
		Action: mocks.ACTION_CREATE_USER.Name,
		Params: []map[string][]byte{
			{
				"$name": types.NewMust("kwil").Bytes(),
				"$age":  types.NewMust(21).Bytes(),
			},
		},
		DBID: db.DBID,
	}, &datasets.ExecOpts{
		Caller: "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	res, err := db.Query(`SELECT * FROM users`)
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 1 {
		t.Fatal("expected 1 row")
	}

	err = master.DropDataset(db.DBID)
	if err != nil {
		t.Fatal(err)
	}

	ds, err := master.GetDataset(db.DBID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if ds != nil {
		t.Fatal("expected nil, got dataset")
	}

}

// wipe and deploy guarantees that the database is brand new with a fresh schema
func wipeAndDeploy(t *testing.T, master *engine.Engine, schema *models.Dataset) (*datasets.Dataset, error) {
	master.DropDataset(schema.ID())

	err := master.Deploy(schema)
	if err != nil {
		return nil, err
	}

	return master.GetDataset(schema.ID())
}
