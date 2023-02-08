package executables_test

import (
	"fmt"
	"kwil/pkg/databases"
	"kwil/pkg/databases/executables"
	"kwil/pkg/databases/mocks"
	"kwil/pkg/databases/spec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DBI(t *testing.T) {
	dbi, err := executables.FromDatabase(&mocks.Db1)
	if err != nil {
		t.Error(err)
	}
	q1 := mocks.Insert1
	// default role has access, so this should be true
	canExecute := dbi.CanExecute("0xbennan", q1.Name)
	assert.True(t, canExecute)

	// default role does not have access, so this should be false
	canExecute = dbi.CanExecute("0xbennan", mocks.Insert2.Name)
	assert.False(t, canExecute)

	id := databases.GenerateSchemaName(mocks.Db1.Owner, mocks.Db1.Name)

	// check that the dbi has the correct identifier
	assert.Equal(t, id, dbi.GetDbId())

	// also assert that the identifier with name and owner is correct
	assert.Equal(t, mocks.Db1.Owner, dbi.Owner)
	assert.Equal(t, mocks.Db1.Name, dbi.Name)
}

func Test_PrepareInsert(t *testing.T) {
	dbi, err := executables.FromDatabase(&mocks.Db1)
	if err != nil {
		t.Error(err)
	}

	// make an input for query 1
	p2, err := spec.NewExplicit(5, spec.INT32)
	if err != nil {
		t.Error(err)
	}

	var inputs []*executables.UserInput
	inputs = append(inputs, &executables.UserInput{
		Name:  "param2",
		Value: p2.Bytes(),
	})

	// prepare query 1
	query, params, err := dbi.Prepare(mocks.Insert1.Name, "0xbennan", inputs)
	if err != nil {
		t.Error(err)
	}

	// assert that the params are correct
	assert.Contains(t, params, "0xbennan")
	p2int64, err := p2.AsInt64()
	if err != nil {
		t.Error(err)
	}

	assert.Contains(t, params, p2int64)

	assert.Contains(t, query, "INSERT INTO")
	assert.Contains(t, query, "VALUES ($1, $2)")
	assert.Contains(t, query, mocks.Insert1.Table)
}

func Test_PrepareUpdate(t *testing.T) {
	dbi, err := executables.FromDatabase(&mocks.Db1)
	if err != nil {
		t.Error(err)
	}

	// make an input for update 1
	p2, err := spec.NewExplicit(5, spec.INT32)
	if err != nil {
		t.Error(err)
	}

	var inputs []*executables.UserInput
	inputs = append(inputs, &executables.UserInput{
		Name:  "param2",
		Value: p2.Bytes(),
	})

	stmt, params, err := dbi.Prepare(mocks.Update1.Name, "0xbennan", inputs)
	if err != nil {
		t.Error(err)
	}

	//checking args
	assert.Contains(t, params, "0xbennan")
	p2int64, err := p2.AsInt64()
	if err != nil {
		t.Error(err)
	}

	assert.Contains(t, params, p2int64)

	//checking query
	assert.Contains(t, stmt, "UPDATE")
	assert.Contains(t, stmt, "SET")
	assert.Contains(t, stmt, "WHERE")
	assert.Contains(t, stmt, mocks.Update1.Table)
	assert.Contains(t, stmt, "$3")
}

func Test_PrepareDelete(t *testing.T) {
	dbi, err := executables.FromDatabase(&mocks.Db1)
	if err != nil {
		t.Error(err)
	}

	// make an input for update 1
	w1, err := spec.NewExplicit(true, spec.BOOLEAN)
	if err != nil {
		t.Error(err)
	}

	var inputs []*executables.UserInput
	inputs = append(inputs, &executables.UserInput{
		Name:  "where1",
		Value: w1.Bytes(),
	})

	stmt, params, err := dbi.Prepare(mocks.Delete2.Name, "0xbennan", inputs)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(params)
	fmt.Println(stmt)

	//checking args
	assert.Len(t, params, 0) // if boolean, it does not get added to params

	//checking query
	assert.Contains(t, stmt, "DELETE FROM")
	assert.Contains(t, stmt, "WHERE")
	assert.Contains(t, stmt, mocks.Delete2.Table)
	assert.Contains(t, stmt, "TRUE")
}
