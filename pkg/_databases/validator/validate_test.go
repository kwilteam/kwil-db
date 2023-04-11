package validator_test

import (
	"kwil/pkg/databases"
	"kwil/pkg/databases/mocks"
	"kwil/pkg/databases/spec"
	"kwil/pkg/databases/validator"
	"testing"

	"github.com/mitchellh/copystructure"
)

func Test_Validate(t *testing.T) {
	db := mocks.Db1
	v := validator.Validator{}
	err := v.Validate(&db)
	if err != nil {
		t.Errorf("error validating database: %v", err)
	}

	// testing names
	copyDB := mustCopy(db)
	testNames(t, copyDB)
	testMaxCounts(t, copyDB)
	testAttributes(t)
	testQueries(t, db)

}

func testNames(t *testing.T, db databases.Database[*spec.KwilAny]) {
	v := validator.Validator{}

	table := &databases.Table[*spec.KwilAny]{
		Name: "SELECT",
		Columns: []*databases.Column[*spec.KwilAny]{
			&mocks.Column1, // just need something to pass validation
		},
	}
	db.Tables = append(db.Tables, table)

	err := v.Validate(&db)
	if err == nil {
		t.Errorf("expected error validating database")
	}

	table.Name = "select"
	db.Name = "select"
	err = v.Validate(&db)
	if err == nil {
		t.Errorf("expected error validating database")
	}

	db.Name = "SELECT * FROM table"
	err = v.Validate(&db)
	if err == nil {
		t.Errorf("expected error validating database")
	}

}

func testMaxCounts(t *testing.T, db databases.Database[*spec.KwilAny]) {
	v := validator.Validator{DB: &db}

	// testing table count
	for i := 0; i < validator.MAX_TABLE_COUNT+1; i++ {
		db.Tables = append(db.Tables, &databases.Table[*spec.KwilAny]{})
	}

	err := v.Validate(&db)
	if err == nil {
		t.Errorf("expected error validating database")
	}
}

func testAttributes(t *testing.T) {
	v := validator.Validator{}

	// testing min attribute on boolean column
	col := &databases.Column[*spec.KwilAny]{
		Name: "test",
		Type: spec.BOOLEAN,
		Attributes: []*databases.Attribute[*spec.KwilAny]{
			{
				Type:  spec.MIN,
				Value: spec.NewMust(4),
			},
		},
	}

	if err := v.ValidateColumn(col); err == nil {
		t.Errorf("expected error validating database")
	}

	// testing that I can't have default and unique
	col2 := &databases.Column[*spec.KwilAny]{
		Name: "test",
		Type: spec.STRING,
		Attributes: []*databases.Attribute[*spec.KwilAny]{
			{
				Type:  spec.DEFAULT,
				Value: spec.NewMust("test"),
			},
			{
				Type:  spec.UNIQUE,
				Value: spec.NewMust(nil),
			},
		},
	}

	if err := v.ValidateColumn(col2); err == nil {
		t.Errorf("expected error validating database")
	}
}

func testQueries(t *testing.T, db databases.Database[*spec.KwilAny]) {
	v := validator.Validator{DB: &db}

	// testing that insert with no params is invalid
	q := &databases.SQLQuery[*spec.KwilAny]{
		Name:   "test",
		Type:   spec.INSERT,
		Table:  "table1",
		Params: []*databases.Parameter[*spec.KwilAny]{},
	}

	if err := v.ValidateQuery(q); err == nil {
		t.Errorf("expected error validating database")
	}

	// testing that delete with no wheres
	q = &databases.SQLQuery[*spec.KwilAny]{
		Name:   "test",
		Type:   spec.DELETE,
		Table:  "table1",
		Params: []*databases.Parameter[*spec.KwilAny]{},
		Where:  []*databases.WhereClause[*spec.KwilAny]{},
	}

	if err := v.ValidateQuery(q); err == nil {
		t.Errorf("expected error validating database")
	}

	// test that a insert that does not cover a table with not null columns is invalid
	table := &databases.Table[*spec.KwilAny]{
		Name: "thetable",
		Columns: []*databases.Column[*spec.KwilAny]{
			{
				Name: "col1",
				Type: spec.STRING,
				Attributes: []*databases.Attribute[*spec.KwilAny]{
					{
						Type:  spec.NOT_NULL,
						Value: spec.NewMust(nil),
					},
					{
						Type:  spec.MIN_LENGTH,
						Value: spec.NewMust(45),
					},
					{
						Type:  spec.MAX_LENGTH,
						Value: spec.NewMust(45),
					},
				},
			},
		},
	}

	db.Tables = append(db.Tables, table)

	q = &databases.SQLQuery[*spec.KwilAny]{
		Name:   "test",
		Type:   spec.INSERT,
		Table:  "thetable",
		Params: []*databases.Parameter[*spec.KwilAny]{},
	}

	if err := v.ValidateQuery(q); err == nil {
		t.Errorf("expected error validating database")
	}

	// test that an insert with modifier CALLER will fail since the table's min length is 45 and max length is 45
	q = &databases.SQLQuery[*spec.KwilAny]{
		Name:  "test",
		Type:  spec.INSERT,
		Table: "thetable",
		Params: []*databases.Parameter[*spec.KwilAny]{
			{
				Name:     "col1",
				Column:   "col1",
				Static:   true,
				Value:    spec.NewMust(nil),
				Modifier: spec.CALLER,
			},
		},
	}

	if err := v.ValidateQuery(q); err == nil {
		t.Errorf("expected error validating database")
	}
}

func mustCopy[T any](i T) T {
	c, err := copystructure.Copy(i)
	if err != nil {
		panic(err)
	}
	return c.(T)
}
