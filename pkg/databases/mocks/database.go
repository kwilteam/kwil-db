package mocks

import (
	execution2 "kwil/pkg/databases"
	"kwil/pkg/databases/spec"
)

var (
	// database
	Db1 = execution2.Database[*spec.KwilAny]{
		Name:  "db1",
		Owner: "0xabc",
		Tables: []*execution2.Table[*spec.KwilAny]{
			&Table1,
			&Table2,
			&Table3,
		},
		SQLQueries: []*execution2.SQLQuery[*spec.KwilAny]{
			&Insert1,
			&Insert2,
			&Insert3,
			&Update1,
			&Update2,
			&Delete1,
			&Delete2,
		},
		Roles: []*execution2.Role{
			&Role1,
			&Role2,
		},
		Indexes: []*execution2.Index{
			&Index1,
		},
	}

	// tables
	Table1 = execution2.Table[*spec.KwilAny]{
		Name:    "table1",
		Columns: []*execution2.Column[*spec.KwilAny]{&Column1, &Column2},
	}

	Table2 = execution2.Table[*spec.KwilAny]{
		Name:    "table2",
		Columns: []*execution2.Column[*spec.KwilAny]{&Column1, &Column3},
	}

	Table3 = execution2.Table[*spec.KwilAny]{
		Name:    "table3",
		Columns: []*execution2.Column[*spec.KwilAny]{&Column1, &Column4},
	}

	// columns
	Column1 = execution2.Column[*spec.KwilAny]{
		Name: "col1",
		Type: spec.STRING,
		Attributes: []*execution2.Attribute[*spec.KwilAny]{
			{
				Type:  spec.PRIMARY_KEY,
				Value: spec.NewMust(nil),
			},
		},
	}

	Column2 = execution2.Column[*spec.KwilAny]{
		Name: "col2",
		Type: spec.INT32,
		Attributes: []*execution2.Attribute[*spec.KwilAny]{
			{
				Type:  spec.MIN,
				Value: spec.NewMust(1),
			},
		},
	}

	Column3 = execution2.Column[*spec.KwilAny]{
		Name: "col3",
		Type: spec.BOOLEAN,
	}

	Column4 = execution2.Column[*spec.KwilAny]{
		Name: "col4",
		Type: spec.UUID,
		Attributes: []*execution2.Attribute[*spec.KwilAny]{
			{
				Type:  spec.UNIQUE,
				Value: spec.NewMust(nil),
			},
		},
	}

	// sql queries

	// insert1 contains a string and an int32
	Insert1 = execution2.SQLQuery[*spec.KwilAny]{
		Name:  "insert1",
		Type:  spec.INSERT,
		Table: "table1",
		Params: []*execution2.Parameter[*spec.KwilAny]{
			&Parameter1,
			&Parameter2,
		},
	}

	// Insert2 contains a string and a boolean
	Insert2 = execution2.SQLQuery[*spec.KwilAny]{
		Name:  "insert2",
		Type:  spec.INSERT,
		Table: "table2",
		Params: []*execution2.Parameter[*spec.KwilAny]{
			&Parameter1,
			&Parameter3,
		},
	}

	// insert3 contains a string and a uuid
	Insert3 = execution2.SQLQuery[*spec.KwilAny]{
		Name:  "insert3",
		Type:  spec.INSERT,
		Table: "table3",
		Params: []*execution2.Parameter[*spec.KwilAny]{
			&Parameter1,
			&Parameter4,
		},
	}

	// update
	Update1 = execution2.SQLQuery[*spec.KwilAny]{
		Name:  "update1",
		Type:  spec.UPDATE,
		Table: "table1",
		Params: []*execution2.Parameter[*spec.KwilAny]{
			&Parameter1,
			&Parameter2,
		},
		Where: []*execution2.WhereClause[*spec.KwilAny]{
			&WhereClause2,
		},
	}

	Update2 = execution2.SQLQuery[*spec.KwilAny]{
		Name:  "update2",
		Type:  spec.UPDATE,
		Table: "table2",
		Params: []*execution2.Parameter[*spec.KwilAny]{
			&Parameter1,
			&Parameter3,
		},
		Where: []*execution2.WhereClause[*spec.KwilAny]{
			&WhereClause1,
		},
	}

	// delete
	Delete1 = execution2.SQLQuery[*spec.KwilAny]{
		Name:  "delete1",
		Type:  spec.DELETE,
		Table: "table1",
		Where: []*execution2.WhereClause[*spec.KwilAny]{
			&WhereClause2,
		},
	}

	Delete2 = execution2.SQLQuery[*spec.KwilAny]{
		Name:  "delete2",
		Type:  spec.DELETE,
		Table: "table2",
		Where: []*execution2.WhereClause[*spec.KwilAny]{
			&WhereClause1,
		},
	}

	// parameters

	// parameter1 is a caller modified string
	Parameter1 = execution2.Parameter[*spec.KwilAny]{
		Name:     "param1",
		Column:   "col1",
		Static:   true,
		Value:    spec.NewMust(""),
		Modifier: spec.CALLER,
	}

	// parameter2 is an int32
	Parameter2 = execution2.Parameter[*spec.KwilAny]{
		Name:   "param2",
		Column: "col2",
	}

	// parameter3 is a boolean
	Parameter3 = execution2.Parameter[*spec.KwilAny]{
		Name:   "param3",
		Column: "col3",
		Static: false,
	}

	// parameter4 is a uuid
	Parameter4 = execution2.Parameter[*spec.KwilAny]{
		Name:   "param4",
		Column: "col4",
	}

	// where clauses

	// whereclause1 checks for EQUALS on col3 (boolean)
	WhereClause1 = execution2.WhereClause[*spec.KwilAny]{
		Name:     "where1",
		Column:   "col3",
		Static:   false,
		Operator: spec.EQUAL,
	}

	// whereclause2 checks for EQUALS on col1 (string), and is caller modified
	WhereClause2 = execution2.WhereClause[*spec.KwilAny]{
		Name:     "where2",
		Column:   "col1",
		Static:   true,
		Operator: spec.EQUAL,
		Value:    spec.NewMust(""),
		Modifier: spec.CALLER,
	}

	// roles

	// role1 has insert1, update1, and delete1 permissions
	Role1 = execution2.Role{
		Name:    "role1",
		Default: true,
		Permissions: []string{
			"insert1",
			"update1",
			"delete1",
		},
	}

	// role2 has all permissions
	Role2 = execution2.Role{
		Name: "role2",
		Permissions: []string{
			"insert1",
			"insert2",
			"insert3",
			"update1",
			"update2",
			"delete1",
			"delete2",
		},
	}

	// indexes
	Index1 = execution2.Index{
		Name:    "my_index",
		Table:   "table1",
		Columns: []string{"col1", "col2"},
		Using:   spec.BTREE,
	}
)
