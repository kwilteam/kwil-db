package mocks

import (
	"kwil/x/execution"
	"kwil/x/execution/dto"
)

var (
	// database
	Db1 = dto.Database{
		Name:  "db1",
		Owner: "0xabc",
		Tables: []*dto.Table{
			&Table1,
			&Table2,
		},
		SQLQueries: []*dto.SQLQuery{
			&Insert1,
			&Insert2,
			&Update1,
			&Update2,
			&Delete1,
			&Delete2,
		},
		Roles: []*dto.Role{
			&Role1,
			&Role2,
		},
		Indexes: []*dto.Index{
			&Index1,
		},
	}

	// tables
	Table1 = dto.Table{
		Name:    "table1",
		Columns: []*dto.Column{&Column1, &Column2},
	}

	Table2 = dto.Table{
		Name:    "table2",
		Columns: []*dto.Column{&Column1, &Column3},
	}

	// columns
	Column1 = dto.Column{
		Name: "col1",
		Type: execution.STRING,
		Attributes: []*dto.Attribute{
			{
				Type:  execution.PRIMARY_KEY,
				Value: nil,
			},
		},
	}

	Column2 = dto.Column{
		Name: "col2",
		Type: execution.INT32,
		Attributes: []*dto.Attribute{
			{
				Type:  execution.MIN,
				Value: 0,
			},
		},
	}

	Column3 = dto.Column{
		Name: "col3",
		Type: execution.BOOLEAN,
	}

	// sql queries

	// insert
	Insert1 = dto.SQLQuery{
		Name:  "insert1",
		Type:  execution.INSERT,
		Table: "table1",
		Params: []*dto.Parameter{
			&Parameter1,
			&Parameter2,
		},
	}

	Insert2 = dto.SQLQuery{
		Name:  "insert2",
		Type:  execution.INSERT,
		Table: "table2",
		Params: []*dto.Parameter{
			&Parameter1,
			&Parameter3,
		},
	}

	// update
	Update1 = dto.SQLQuery{
		Name:  "update1",
		Type:  execution.UPDATE,
		Table: "table1",
		Params: []*dto.Parameter{
			&Parameter1,
			&Parameter2,
		},
		Where: []*dto.WhereClause{
			&WhereClause2,
		},
	}

	Update2 = dto.SQLQuery{
		Name:  "update2",
		Type:  execution.UPDATE,
		Table: "table2",
		Params: []*dto.Parameter{
			&Parameter1,
			&Parameter3,
		},
		Where: []*dto.WhereClause{
			&WhereClause1,
		},
	}

	// delete
	Delete1 = dto.SQLQuery{
		Name:  "delete1",
		Type:  execution.DELETE,
		Table: "table1",
		Where: []*dto.WhereClause{
			&WhereClause2,
		},
	}

	Delete2 = dto.SQLQuery{
		Name:  "delete2",
		Type:  execution.DELETE,
		Table: "table2",
		Where: []*dto.WhereClause{
			&WhereClause1,
		},
	}

	// parameters

	Parameter1 = dto.Parameter{
		Name:     "param1",
		Column:   "col1",
		Static:   true,
		Value:    "",
		Modifier: execution.CALLER,
	}

	Parameter2 = dto.Parameter{
		Name:   "param2",
		Column: "col2",
	}

	Parameter3 = dto.Parameter{
		Name:   "param3",
		Column: "col3",
		Static: false,
	}

	WhereClause1 = dto.WhereClause{
		Name:     "where1",
		Column:   "col3",
		Static:   false,
		Operator: execution.EQUAL,
	}

	WhereClause2 = dto.WhereClause{
		Name:     "where2",
		Column:   "col1",
		Static:   true,
		Operator: execution.EQUAL,
		Value:    "",
		Modifier: execution.CALLER,
	}

	// roles
	Role1 = dto.Role{
		Name:    "role1",
		Default: true,
		Permissions: []string{
			"insert1",
			"update1",
			"delete1",
		},
	}

	Role2 = dto.Role{
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
	Index1 = dto.Index{
		Name:    "my_index",
		Table:   "table1",
		Columns: []string{"col1", "col2"},
		Using:   1,
	}
)
