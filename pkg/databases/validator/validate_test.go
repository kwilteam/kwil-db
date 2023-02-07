package validator_test

import (
	"kwil/pkg/databases"
	"kwil/pkg/databases/mocks"
	"kwil/pkg/databases/validator"
	datatypes "kwil/pkg/types/data_types"
	anytype "kwil/pkg/types/data_types/any_type"
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

func testNames(t *testing.T, db databases.Database[anytype.KwilAny]) {
	v := validator.Validator{}

	table := &databases.Table[anytype.KwilAny]{
		Name: "SELECT",
		Columns: []*databases.Column[anytype.KwilAny]{
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

func testMaxCounts(t *testing.T, db databases.Database[anytype.KwilAny]) {
	v := validator.Validator{}

	// testing table count
	for i := 0; i < databases.MAX_TABLE_COUNT+1; i++ {
		db.Tables = append(db.Tables, &databases.Table[anytype.KwilAny]{})
	}

	err := v.Validate(&db)
	if err == nil {
		t.Errorf("expected error validating database")
	}
}

func testAttributes(t *testing.T) {
	v := validator.Validator{}

	// testing min attribute on boolean column
	col := &databases.Column[anytype.KwilAny]{
		Name: "test",
		Type: datatypes.BOOLEAN,
		Attributes: []*databases.Attribute[anytype.KwilAny]{
			{
				Type:  databases.MIN,
				Value: anytype.NewMust(4),
			},
		},
	}

	if err := v.ValidateColumn(col); err == nil {
		t.Errorf("expected error validating database")
	}

	// testing that I can't have default and unique
	col2 := &databases.Column[anytype.KwilAny]{
		Name: "test",
		Type: datatypes.STRING,
		Attributes: []*databases.Attribute[anytype.KwilAny]{
			{
				Type:  databases.DEFAULT,
				Value: anytype.NewMust("test"),
			},
			{
				Type:  databases.UNIQUE,
				Value: anytype.NewMust(nil),
			},
		},
	}

	if err := v.ValidateColumn(col2); err == nil {
		t.Errorf("expected error validating database")
	}
}

func testQueries(t *testing.T, db databases.Database[anytype.KwilAny]) {
	v := validator.New(&db)

	// testing that insert with no params is invalid
	q := &databases.SQLQuery[anytype.KwilAny]{
		Name:   "test",
		Type:   databases.INSERT,
		Table:  "table1",
		Params: []*databases.Parameter[anytype.KwilAny]{},
	}

	if err := v.ValidateQuery(q); err == nil {
		t.Errorf("expected error validating database")
	}

	// testing that delete with no wheres
	q = &databases.SQLQuery[anytype.KwilAny]{
		Name:   "test",
		Type:   databases.DELETE,
		Table:  "table1",
		Params: []*databases.Parameter[anytype.KwilAny]{},
		Where:  []*databases.WhereClause[anytype.KwilAny]{},
	}

	if err := v.ValidateQuery(q); err == nil {
		t.Errorf("expected error validating database")
	}

	// test that a insert that does not cover a table with not null columns is invalid
	table := &databases.Table[anytype.KwilAny]{
		Name: "thetable",
		Columns: []*databases.Column[anytype.KwilAny]{
			{
				Name: "col1",
				Type: datatypes.STRING,
				Attributes: []*databases.Attribute[anytype.KwilAny]{
					{
						Type:  databases.NOT_NULL,
						Value: anytype.NewMust(nil),
					},
					{
						Type:  databases.MIN_LENGTH,
						Value: anytype.NewMust(45),
					},
					{
						Type:  databases.MAX_LENGTH,
						Value: anytype.NewMust(45),
					},
				},
			},
		},
	}

	db.Tables = append(db.Tables, table)

	q = &databases.SQLQuery[anytype.KwilAny]{
		Name:   "test",
		Type:   databases.INSERT,
		Table:  "thetable",
		Params: []*databases.Parameter[anytype.KwilAny]{},
	}

	if err := v.ValidateQuery(q); err == nil {
		t.Errorf("expected error validating database")
	}

	// test that an insert with modifier CALLER will fail since the table's min length is 45 and max length is 45
	q = &databases.SQLQuery[anytype.KwilAny]{
		Name:  "test",
		Type:  databases.INSERT,
		Table: "thetable",
		Params: []*databases.Parameter[anytype.KwilAny]{
			{
				Name:     "col1",
				Column:   "col1",
				Static:   true,
				Value:    anytype.NewMust(nil),
				Modifier: databases.CALLER,
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
