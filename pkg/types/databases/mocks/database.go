package mocks

import (
	execution2 "kwil/pkg/execution"
	"kwil/pkg/types/data_types"
	"kwil/pkg/types/data_types/any_type"
	databases2 "kwil/pkg/types/databases"
)

var (
	// database
	Db1 = databases2.Database[anytype.KwilAny]{
		Name:  "db1",
		Owner: "0xabc",
		Tables: []*databases2.Table[anytype.KwilAny]{
			&Table1,
			&Table2,
		},
		SQLQueries: []*databases2.SQLQuery[anytype.KwilAny]{
			&Insert1,
			&Insert2,
			&Update1,
			&Update2,
			&Delete1,
			&Delete2,
		},
		Roles: []*databases2.Role{
			&Role1,
			&Role2,
		},
		Indexes: []*databases2.Index{
			&Index1,
		},
	}

	// tables
	Table1 = databases2.Table[anytype.KwilAny]{
		Name:    "table1",
		Columns: []*databases2.Column[anytype.KwilAny]{&Column1, &Column2},
	}

	Table2 = databases2.Table[anytype.KwilAny]{
		Name:    "table2",
		Columns: []*databases2.Column[anytype.KwilAny]{&Column1, &Column3},
	}

	// columns
	Column1 = databases2.Column[anytype.KwilAny]{
		Name: "col1",
		Type: datatypes.STRING,
		Attributes: []*databases2.Attribute[anytype.KwilAny]{
			{
				Type:  execution2.PRIMARY_KEY,
				Value: anytype.NewMust(nil),
			},
		},
	}

	Column2 = databases2.Column[anytype.KwilAny]{
		Name: "col2",
		Type: datatypes.INT32,
		Attributes: []*databases2.Attribute[anytype.KwilAny]{
			{
				Type:  execution2.MIN,
				Value: anytype.NewMust(1),
			},
		},
	}

	Column3 = databases2.Column[anytype.KwilAny]{
		Name: "col3",
		Type: datatypes.BOOLEAN,
	}

	// sql queries

	// insert
	Insert1 = databases2.SQLQuery[anytype.KwilAny]{
		Name:  "insert1",
		Type:  execution2.INSERT,
		Table: "table1",
		Params: []*databases2.Parameter[anytype.KwilAny]{
			&Parameter1,
			&Parameter2,
		},
	}

	Insert2 = databases2.SQLQuery[anytype.KwilAny]{
		Name:  "insert2",
		Type:  execution2.INSERT,
		Table: "table2",
		Params: []*databases2.Parameter[anytype.KwilAny]{
			&Parameter1,
			&Parameter3,
		},
	}

	// update
	Update1 = databases2.SQLQuery[anytype.KwilAny]{
		Name:  "update1",
		Type:  execution2.UPDATE,
		Table: "table1",
		Params: []*databases2.Parameter[anytype.KwilAny]{
			&Parameter1,
			&Parameter2,
		},
		Where: []*databases2.WhereClause[anytype.KwilAny]{
			&WhereClause2,
		},
	}

	Update2 = databases2.SQLQuery[anytype.KwilAny]{
		Name:  "update2",
		Type:  execution2.UPDATE,
		Table: "table2",
		Params: []*databases2.Parameter[anytype.KwilAny]{
			&Parameter1,
			&Parameter3,
		},
		Where: []*databases2.WhereClause[anytype.KwilAny]{
			&WhereClause1,
		},
	}

	// delete
	Delete1 = databases2.SQLQuery[anytype.KwilAny]{
		Name:  "delete1",
		Type:  execution2.DELETE,
		Table: "table1",
		Where: []*databases2.WhereClause[anytype.KwilAny]{
			&WhereClause2,
		},
	}

	Delete2 = databases2.SQLQuery[anytype.KwilAny]{
		Name:  "delete2",
		Type:  execution2.DELETE,
		Table: "table2",
		Where: []*databases2.WhereClause[anytype.KwilAny]{
			&WhereClause1,
		},
	}

	// parameters

	Parameter1 = databases2.Parameter[anytype.KwilAny]{
		Name:     "param1",
		Column:   "col1",
		Static:   true,
		Value:    anytype.NewMust(""),
		Modifier: execution2.CALLER,
	}

	Parameter2 = databases2.Parameter[anytype.KwilAny]{
		Name:   "param2",
		Column: "col2",
	}

	Parameter3 = databases2.Parameter[anytype.KwilAny]{
		Name:   "param3",
		Column: "col3",
		Static: false,
	}

	WhereClause1 = databases2.WhereClause[anytype.KwilAny]{
		Name:     "where1",
		Column:   "col3",
		Static:   false,
		Operator: execution2.EQUAL,
	}

	WhereClause2 = databases2.WhereClause[anytype.KwilAny]{
		Name:     "where2",
		Column:   "col1",
		Static:   true,
		Operator: execution2.EQUAL,
		Value:    anytype.NewMust(""),
		Modifier: execution2.CALLER,
	}

	// roles
	Role1 = databases2.Role{
		Name:    "role1",
		Default: true,
		Permissions: []string{
			"insert1",
			"update1",
			"delete1",
		},
	}

	Role2 = databases2.Role{
		Name: "role2",
		Permissions: []string{
			"insert1",
			"insert2",
			"update1",
			"update2",
			"delete1",
			"delete2",
		},
	}

	// indexes
	Index1 = databases2.Index{
		Name:    "my_index",
		Table:   "table1",
		Columns: []string{"col1", "col2"},
		Using:   1,
	}
)
