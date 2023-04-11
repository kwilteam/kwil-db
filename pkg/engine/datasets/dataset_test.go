package datasets_test

import (
	"encoding/json"
	"fmt"
	"kwil/pkg/engine/datasets"
	"kwil/pkg/engine/models"
	"kwil/pkg/engine/models/mocks"
	"kwil/pkg/engine/types"
	"os"
	"testing"
)

func Test_Dataset(t *testing.T) {
	ds, err := datasets.OpenDataset("owner", "name", getDir())
	if err != nil {
		t.Fatal(err)
	}
	ds, err = clearSet(ds)
	if err != nil {
		t.Fatal(err)
	}
	defer ds.Close()

	err = ds.ApplySchema(&mocks.MOCK_DATASET1)
	if err != nil {
		t.Fatal(err)
	}

	params := make([]map[string][]byte, 0)
	args := make(map[string][]byte)
	args["$name"] = types.NewMust("kwil").Bytes()
	args["$age"] = types.NewMust(21).Bytes()

	params = append(params, args)
	// testing execution
	_, err = ds.ExecuteAction(&models.ActionExecution{
		Action: mocks.ACTION_CREATE_USER.Name,
		Params: params,
		DBID:   ds.DBID, // not technically needed here
	}, &datasets.ExecOpts{
		Caller: "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	// read it back
	res1, err := ds.Query("SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}
	if len(res1) != 1 {
		t.Fatal(`expected 1 row, got: `, len(res1))
	}

	// read back an action with results
	res, err := ds.ExecuteAction(&models.ActionExecution{
		Action: mocks.ACTION_GET_ALL_USERS.Name,
		DBID:   ds.DBID, // not technically needed here
	}, &datasets.ExecOpts{
		Caller: "test",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(res) != 1 {
		t.Fatal(`expected 1 row, got: `, len(res))
	}
	if len(res[0]) != 4 {
		t.Fatal(`expected 4 columns, got: `, len(res[0]))
	}

	bts, err := res.Bytes()
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(bts))

	var res2 []map[string]any
	err = json.Unmarshal(bts, &res2)
	if err != nil {
		t.Fatal(err)
	}

	if len(res2) != 1 {
		t.Fatal(`expected 1 row, got: `, len(res2))
	}

	if len(res2[0]) != 4 {
		t.Fatal(`expected 4 columns, got: `, len(res2[0]))
	}
}

func Test_Read(t *testing.T) {
	ds, err := datasets.OpenDataset("owner", "name", getDir())
	if err != nil {
		t.Fatal(err)
	}
	defer ds.Close()

	// read it back
	res, err := ds.Query("SELECT * FROM users")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(res)

	// try to write to the read-only db
	_, err = ds.Query("INSERT INTO users (name, age) VALUES ('kwil', 21)")
	if err == nil {
		t.Fatal("expected an error")
	}
}

func getDir() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		dirname = "/tmp"
	}

	defaultPath := fmt.Sprintf("%s/.kwil/sqlite/", dirname)
	return defaultPath
}

func clearSet(ds *datasets.Dataset) (*datasets.Dataset, error) {
	name := ds.Name
	owner := ds.Owner

	err := ds.Wipe()
	if err != nil {
		return nil, err
	}
	err = ds.Close()
	if err != nil {
		return nil, err
	}

	return datasets.OpenDataset(owner, name, getDir())
}
