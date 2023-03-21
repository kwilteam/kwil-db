package engine_test

import (
	"kwil/pkg/engine"
	"kwil/pkg/engine/models"
	"testing"
)

func Test_Engine_CreateDataset(t *testing.T) {
	master, err := engine.Open()
	if err != nil {
		t.Fatal(err)
	}

	defer master.Close()

	// Register a new database
	err = master.CreateDataset("ownerabc123", "test")
	if err != nil {
		t.Fatal(err)
	}

	dbid := models.GenerateSchemaId("ownerabc123", "test")

	// retrieve the database
	_, ok := master.Datasets[dbid]
	if !ok {
		t.Fatal("database not found")
	}

	err = master.DeleteDataset(models.GenerateSchemaId("ownerabc123", "test"))
	if err != nil {
		t.Fatal(err)
	}
}
