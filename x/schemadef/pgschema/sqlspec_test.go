package pgschema

import (
	"fmt"
	"strconv"
	"testing"

	"kwil/x/schemadef/sqlschema"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	f := `schema "public" {}

table "users" {
	schema = sqlschema.public
    column "id" {
        type = int
    }
    column "name" {
        type = varchar(255)
    }
}

query "add_user" {
    statement = <<SQL
    INSERT INTO users (name) VALUES (@name)
    SQL
}
`
	var s sqlschema.Schema
	err := EvalHCLBytes([]byte(f), &s, nil)
	require.NoError(t, err)
}

func TestSQLSpec(t *testing.T) {
	f := `
schema "schema" {
}

table "table" {
	schema = sqlschema.schema
	column "col" {
		type = integer
		comment = "column comment"
	}
	column "age" {
		type = integer
	}
	column "price" {
		type = int
	}
	column "account_name" {
		type = varchar(32)
	}
	column "varchar_length_is_not_required" {
		type = varchar
	}
	column "character_varying_length_is_not_required" {
		type = character_varying
	}
	column "tags" {
		type = hstore
	}
	column "created_at" {
		type    = timestamp(4)
		default = sql("current_timestamp(4)")
	}
	column "updated_at" {
		type    = time
		default = sql("current_time")
	}
	primary_key {
		columns = [table.table.column.col]
	}
	index "index" {
		type = HASH
		unique = true
		columns = [
			table.table.column.col,
			table.table.column.age,
		]
		where = "active"
		comment = "index comment"
	}
	foreign_key "accounts" {
		columns = [
			table.table.column.account_name,
		]
		ref_columns = [
			table.accounts.column.name,
		]
		on_delete = SET_NULL
	}
	check "positive price" {
		expr = "price > 0"
	}
	comment = "table comment"
}

table "accounts" {
	schema = sqlschema.schema
	column "name" {
		type = varchar(32)
	}
	column "type" {
		type = enum.account_type
	}
	primary_key {
		columns = [table.accounts.column.name]
	}
}

enum "account_type" {
	schema = sqlschema.schema
	values = ["private", "business"]
}
`
	var s sqlschema.Schema
	err := EvalHCLBytes([]byte(f), &s, nil)
	require.NoError(t, err)
	exp := &sqlschema.Schema{
		Name: "schema",
	}
	exp.Tables = []*sqlschema.Table{
		{
			Name:   "table",
			Schema: exp,
			Columns: []*sqlschema.Column{
				{
					Name: "col",
					Type: &sqlschema.ColumnType{
						Type: &sqlschema.IntegerType{
							T: "integer",
						},
					},
					Attrs: []sqlschema.Attr{
						&sqlschema.Comment{Text: "column comment"},
					},
				},
				{
					Name: "age",
					Type: &sqlschema.ColumnType{
						Type: &sqlschema.IntegerType{
							T: "integer",
						},
					},
				},
				{
					Name: "price",
					Type: &sqlschema.ColumnType{
						Type: &sqlschema.IntegerType{
							T: TypeInt,
						},
					},
				},
				{
					Name: "account_name",
					Type: &sqlschema.ColumnType{
						Type: &sqlschema.StringType{
							T:    "varchar",
							Size: 32,
						},
					},
				},
				{
					Name: "varchar_length_is_not_required",
					Type: &sqlschema.ColumnType{
						Type: &sqlschema.StringType{
							T:    "varchar",
							Size: 0,
						},
					},
				},
				{
					Name: "character_varying_length_is_not_required",
					Type: &sqlschema.ColumnType{
						Type: &sqlschema.StringType{
							T:    "character varying",
							Size: 0,
						},
					},
				},
				{
					Name: "tags",
					Type: &sqlschema.ColumnType{
						Type: &UserDefinedType{
							T: "hstore",
						},
					},
				},
				{
					Name: "created_at",
					Type: &sqlschema.ColumnType{
						Type: typeTime(TypeTimestamp, 4),
					},
					Default: &sqlschema.RawExpr{X: "current_timestamp(4)"},
				},
				{
					Name: "updated_at",
					Type: &sqlschema.ColumnType{
						Type: typeTime(TypeTime, 6),
					},
					Default: &sqlschema.RawExpr{X: "current_time"},
				},
			},
			Attrs: []sqlschema.Attr{
				&sqlschema.Check{
					Name: "positive price",
					Expr: "price > 0",
				},
				&sqlschema.Comment{Text: "table comment"},
			},
		},
		{
			Name:   "accounts",
			Schema: exp,
			Columns: []*sqlschema.Column{
				{
					Name: "name",
					Type: &sqlschema.ColumnType{
						Type: &sqlschema.StringType{
							T:    "varchar",
							Size: 32,
						},
					},
				},
				{
					Name: "type",
					Type: &sqlschema.ColumnType{
						Type: &sqlschema.EnumType{
							T:      "account_type",
							Values: []string{"private", "business"},
							Schema: exp,
						},
					},
				},
			},
		},
	}
	exp.Tables[0].PrimaryKey = &sqlschema.Index{
		Table: exp.Tables[0],
		Parts: []*sqlschema.IndexPart{
			{Seq: 0, Column: exp.Tables[0].Columns[0]},
		},
	}
	exp.Tables[0].Indexes = []*sqlschema.Index{
		{
			Name:   "index",
			Table:  exp.Tables[0],
			Unique: true,
			Parts: []*sqlschema.IndexPart{
				{Seq: 0, Column: exp.Tables[0].Columns[0]},
				{Seq: 1, Column: exp.Tables[0].Columns[1]},
			},
			Attrs: []sqlschema.Attr{
				&sqlschema.Comment{Text: "index comment"},
				&IndexType{T: IndexTypeHash},
				&IndexPredicate{Predicate: "active"},
			},
		},
	}
	exp.Tables[0].ForeignKeys = []*sqlschema.ForeignKey{
		{
			Name:       "accounts",
			Table:      exp.Tables[0],
			Columns:    []*sqlschema.Column{exp.Tables[0].Columns[3]},
			RefTable:   exp.Tables[1],
			RefColumns: []*sqlschema.Column{exp.Tables[1].Columns[0]},
			OnDelete:   sqlschema.SetNull,
		},
	}
	exp.Tables[1].PrimaryKey = &sqlschema.Index{
		Table: exp.Tables[1],
		Parts: []*sqlschema.IndexPart{
			{Seq: 0, Column: exp.Tables[1].Columns[0]},
		},
	}
	exp.Realm = sqlschema.NewRealm(exp)
	exp.Enums = []*sqlschema.Enum{
		{
			Name:   "account_type",
			Schema: exp,
			Values: []string{"private", "business"},
		},
	}

	require.EqualValues(t, exp, &s)
}

func TestUnmarshalSpec_IndexType(t *testing.T) {
	f := `
schema "s" {}
table "t" {
	schema = sqlschema.s
	column "c" {
		type = int
	}
	index "i" {
		type = %s
		columns = [column.c]
	}
}
`
	t.Run("Invalid", func(t *testing.T) {
		f := fmt.Sprintf(f, "UNK")
		err := EvalHCLBytes([]byte(f), &sqlschema.Schema{}, nil)
		require.Error(t, err)
	})
	t.Run("Valid", func(t *testing.T) {
		var (
			s sqlschema.Schema
			f = fmt.Sprintf(f, "HASH")
		)
		err := EvalHCLBytes([]byte(f), &s, nil)
		require.NoError(t, err)
		idx := s.Tables[0].Indexes[0]
		require.Equal(t, IndexTypeHash, idx.Attrs[0].(*IndexType).T)
	})
}

func TestUnmarshalSpec_BRINIndex(t *testing.T) {
	f := `
schema "s" {}
table "t" {
	schema = sqlschema.s
	column "c" {
		type = int
	}
	index "i" {
		type = BRIN
		columns = [column.c]
		page_per_range = 2
	}
}
`
	var s sqlschema.Schema
	err := EvalHCLBytes([]byte(f), &s, nil)
	require.NoError(t, err)
	idx := s.Tables[0].Indexes[0]
	require.Equal(t, IndexTypeBRIN, idx.Attrs[0].(*IndexType).T)
	require.EqualValues(t, 2, idx.Attrs[1].(*IndexStorageParams).PagesPerRange)
}

func TestUnmarshalSpec_Partitioned(t *testing.T) {
	t.Run("Columns", func(t *testing.T) {
		var (
			s = &sqlschema.Schema{}
			f = `
schema "test" {}
table "logs" {
	schema = sqlschema.test
	column "name" {
		type = text
	}
	partition {
		type = HASH
		columns = [
			column.name
		]
	}
}
`
		)
		err := EvalHCLBytes([]byte(f), s, nil)
		require.NoError(t, err)
		c := sqlschema.NewStringColumn("name", "text")
		expected := sqlschema.New("test").
			AddTables(sqlschema.NewTable("logs").AddColumns(c).AddAttrs(&Partition{T: PartitionTypeHash, Parts: []*PartitionPart{{Column: c.Name}}}))
		expected.SetRealm(sqlschema.NewRealm(expected))
		require.Equal(t, expected, s)
	})

	t.Run("Parts", func(t *testing.T) {
		var (
			s = &sqlschema.Schema{}
			f = `
schema "test" {}
table "logs" {
	schema = sqlschema.test
	column "name" {
		type = text
	}
	partition {
		type = RANGE
		by {
			column = column.name
		}
		by {
			expr = "lower(name)"
		}
	}
}
`
		)
		err := EvalHCLBytes([]byte(f), s, nil)
		require.NoError(t, err)
		c := sqlschema.NewStringColumn("name", "text")
		expected := sqlschema.New("test").
			AddTables(sqlschema.NewTable("logs").AddColumns(c).AddAttrs(&Partition{T: PartitionTypeRange, Parts: []*PartitionPart{{Column: c.Name}, {Expr: &sqlschema.RawExpr{X: "lower(name)"}}}}))
		expected.SetRealm(sqlschema.NewRealm(expected))
		require.Equal(t, expected, s)
	})

	t.Run("Invalid", func(t *testing.T) {
		err := EvalHCLBytes([]byte(`
			schema "test" {}
			table "logs" {
				schema = sqlschema.test
				column "name" { type = text }
				partition {
					columns = [column.name]
				}
			}
		`), &sqlschema.Schema{}, nil)
		require.EqualError(t, err, "missing attribute logs.partition.type")

		err = EvalHCLBytes([]byte(`
			schema "test" {}
			table "logs" {
				schema = sqlschema.test
				column "name" { type = text }
				partition {
					type = HASH
				}
			}
		`), &sqlschema.Schema{}, nil)
		require.EqualError(t, err, `missing columns or expressions for logs.partition`)

		err = EvalHCLBytes([]byte(`
			schema "test" {}
			table "logs" {
				schema = sqlschema.test
				column "name" { type = text }
				partition {
					type = HASH
					columns = [column.name]
					by { column = column.name }
				}
			}
		`), &sqlschema.Schema{}, nil)
		require.EqualError(t, err, `multiple definitions for logs.partition, use "columns" or "by"`)
	})
}

func TestMarshalSpec_Partitioned(t *testing.T) {
	t.Run("Columns", func(t *testing.T) {
		c := sqlschema.NewStringColumn("name", "text")
		s := sqlschema.New("test").
			AddTables(sqlschema.NewTable("logs").AddColumns(c).AddAttrs(&Partition{T: PartitionTypeHash, Parts: []*PartitionPart{{Column: c.Name}}}))
		buf, err := MarshalHCL(s)
		require.NoError(t, err)
		require.Equal(t, `table "logs" {
  schema = sqlschema.test
  column "name" {
    null = false
    type = text
  }
  partition {
    type    = HASH
    columns = [column.name]
  }
}
schema "test" {
}
`, string(buf))
	})

	t.Run("Parts", func(t *testing.T) {
		c := sqlschema.NewStringColumn("name", "text")
		s := sqlschema.New("test").
			AddTables(sqlschema.NewTable("logs").AddColumns(c).AddAttrs(&Partition{T: PartitionTypeHash, Parts: []*PartitionPart{{Column: c.Name}, {Expr: &sqlschema.RawExpr{X: "lower(name)"}}}}))
		buf, err := MarshalHCL(s)
		require.NoError(t, err)
		require.Equal(t, `table "logs" {
  schema = sqlschema.test
  column "name" {
    null = false
    type = text
  }
  partition {
    type = HASH
    by {
      column = column.name
    }
    by {
      expr = "lower(name)"
    }
  }
}
schema "test" {
}
`, string(buf))
	})
}

func TestMarshalSpec_IndexPredicate(t *testing.T) {
	s := &sqlschema.Schema{
		Name: "test",
		Tables: []*sqlschema.Table{
			{
				Name: "users",
				Columns: []*sqlschema.Column{
					{
						Name: "id",
						Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "int"}},
					},
				},
			},
		},
	}
	s.Tables[0].Schema = s
	s.Tables[0].Schema = s
	s.Tables[0].Indexes = []*sqlschema.Index{
		{
			Name:   "index",
			Table:  s.Tables[0],
			Unique: true,
			Parts: []*sqlschema.IndexPart{
				{Seq: 0, Column: s.Tables[0].Columns[0]},
			},
			Attrs: []sqlschema.Attr{
				&IndexPredicate{Predicate: "id <> 0"},
			},
		},
	}
	buf, err := MarshalSpec(s, hclState)
	require.NoError(t, err)
	const expected = `table "users" {
  schema = sqlschema.test
  column "id" {
    null = false
    type = int
  }
  index "index" {
    unique  = true
    columns = [column.id]
    where   = "id <> 0"
  }
}
schema "test" {
}
`
	require.EqualValues(t, expected, string(buf))
}

func TestMarshalSpec_BRINIndex(t *testing.T) {
	s := &sqlschema.Schema{
		Name: "test",
		Tables: []*sqlschema.Table{
			{
				Name: "users",
				Columns: []*sqlschema.Column{
					{
						Name: "id",
						Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "int"}},
					},
				},
			},
		},
	}
	s.Tables[0].Schema = s
	s.Tables[0].Schema = s
	s.Tables[0].Indexes = []*sqlschema.Index{
		{
			Name:   "index",
			Table:  s.Tables[0],
			Unique: true,
			Parts: []*sqlschema.IndexPart{
				{Seq: 0, Column: s.Tables[0].Columns[0]},
			},
			Attrs: []sqlschema.Attr{
				&IndexType{T: IndexTypeBRIN},
				&IndexStorageParams{PagesPerRange: 2},
			},
		},
	}
	buf, err := MarshalSpec(s, hclState)
	require.NoError(t, err)
	const expected = `table "users" {
  schema = sqlschema.test
  column "id" {
    null = false
    type = int
  }
  index "index" {
    unique         = true
    columns        = [column.id]
    type           = BRIN
    page_per_range = 2
  }
}
schema "test" {
}
`
	require.EqualValues(t, expected, string(buf))
}

func TestUnmarshalSpec_Identity(t *testing.T) {
	f := `
schema "s" {}
table "t" {
	schema = sqlschema.s
	column "c" {
		type = int
		identity {
			generated = %s
			start = 10
		}
	}
}
`
	t.Run("Invalid", func(t *testing.T) {
		f := fmt.Sprintf(f, "UNK")
		err := EvalHCLBytes([]byte(f), &sqlschema.Schema{}, nil)
		require.Error(t, err)
	})
	t.Run("Valid", func(t *testing.T) {
		var (
			s sqlschema.Schema
			f = fmt.Sprintf(f, "ALWAYS")
		)
		err := EvalHCLBytes([]byte(f), &s, nil)
		require.NoError(t, err)
		id := s.Tables[0].Columns[0].Attrs[0].(*Identity)
		require.Equal(t, GeneratedTypeAlways, id.Generation)
		require.EqualValues(t, 10, id.Sequence.Start)
		require.Zero(t, id.Sequence.Increment)
	})
}

func TestUnmarshalSpec_IndexInclude(t *testing.T) {
	f := `
schema "s" {}
table "t" {
	schema = sqlschema.s
	column "c" {
		type = int
	}
	column "d" {
		type = int
	}
	index "c" {
		columns = [
			column.c,
		]
		include = [
			column.d,
		]
	}
}
`
	var s sqlschema.Schema
	err := EvalHCLBytes([]byte(f), &s, nil)
	require.NoError(t, err)
	require.Len(t, s.Tables, 1)
	require.Len(t, s.Tables[0].Columns, 2)
	require.Len(t, s.Tables[0].Indexes, 1)
	idx, ok := s.Tables[0].Index("c")
	require.True(t, ok)
	require.Len(t, idx.Parts, 1)
	require.Len(t, idx.Attrs, 1)
	var include IndexInclude
	require.True(t, sqlschema.Has(idx.Attrs, &include))
	require.Len(t, include.Columns, 1)
	require.Equal(t, "d", include.Columns[0])
}

func TestMarshalSpec_IndexInclude(t *testing.T) {
	s := &sqlschema.Schema{
		Name: "test",
		Tables: []*sqlschema.Table{
			{
				Name: "users",
				Columns: []*sqlschema.Column{
					{
						Name: "c",
						Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "int"}},
					},
					{
						Name: "d",
						Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "int"}},
					},
				},
			},
		},
	}
	s.Tables[0].Schema = s
	s.Tables[0].Schema = s
	s.Tables[0].Indexes = []*sqlschema.Index{
		{
			Name:  "index",
			Table: s.Tables[0],
			Parts: []*sqlschema.IndexPart{
				{Seq: 0, Column: s.Tables[0].Columns[0]},
			},
			Attrs: []sqlschema.Attr{
				&IndexInclude{Columns: columnNames(s.Tables[0].Columns[1:])},
			},
		},
	}
	buf, err := MarshalSpec(s, hclState)
	require.NoError(t, err)
	const expected = `table "users" {
  schema = sqlschema.test
  column "c" {
    null = false
    type = int
  }
  column "d" {
    null = false
    type = int
  }
  index "index" {
    columns = [column.c]
    include = [column.d]
  }
}
schema "test" {
}
`
	require.EqualValues(t, expected, string(buf))
}

func TestMarshalSpec_GeneratedColumn(t *testing.T) {
	s := sqlschema.New("test").
		AddTables(
			sqlschema.NewTable("users").
				AddColumns(
					sqlschema.NewIntColumn("c1", "int").
						SetGeneratedExpr(&sqlschema.GeneratedExpr{Expr: "c1 * 2"}),
					sqlschema.NewIntColumn("c2", "int").
						SetGeneratedExpr(&sqlschema.GeneratedExpr{Expr: "c3 * c4", Type: "STORED"}),
				),
		)
	buf, err := MarshalSpec(s, hclState)
	require.NoError(t, err)
	const expected = `table "users" {
  schema = sqlschema.test
  column "c1" {
    null = false
    type = int
    as {
      expr = "c1 * 2"
      type = STORED
    }
  }
  column "c2" {
    null = false
    type = int
    as {
      expr = "c3 * c4"
      type = STORED
    }
  }
}
schema "test" {
}
`
	require.EqualValues(t, expected, string(buf))
}

func TestUnmarshalSpec_GeneratedColumns(t *testing.T) {
	var (
		s sqlschema.Schema
		f = `
schema "test" {}
table "users" {
	schema = sqlschema.test
	column "c1" {
		type = int
		as = "1"
	}
	column "c2" {
		type = int
		as {
			expr = "2"
		}
	}
	column "c3" {
		type = int
		as {
			expr = "3"
			type = STORED
		}
	}
}
`
	)
	err := EvalHCLBytes([]byte(f), &s, nil)
	require.NoError(t, err)
	expected := sqlschema.New("test").
		AddTables(
			sqlschema.NewTable("users").
				AddColumns(
					sqlschema.NewIntColumn("c1", "int").SetGeneratedExpr(&sqlschema.GeneratedExpr{Expr: "1", Type: "STORED"}),
					sqlschema.NewIntColumn("c2", "int").SetGeneratedExpr(&sqlschema.GeneratedExpr{Expr: "2", Type: "STORED"}),
					sqlschema.NewIntColumn("c3", "int").SetGeneratedExpr(&sqlschema.GeneratedExpr{Expr: "3", Type: "STORED"}),
				),
		)
	expected.SetRealm(sqlschema.NewRealm(expected))
	require.EqualValues(t, expected, &s)
}

func TestMarshalSpec_Enum(t *testing.T) {
	s := sqlschema.New("test").
		AddTables(
			sqlschema.NewTable("account").
				AddColumns(
					sqlschema.NewEnumColumn("account_type",
						sqlschema.EnumName("account_type"),
						sqlschema.EnumValues("private", "business"),
					),
					sqlschema.NewColumn("account_states").
						SetType(&ArrayType{
							T: "states[]",
							Type: &sqlschema.EnumType{
								T:      "state",
								Values: []string{"on", "off"},
							},
						}),
				),
			sqlschema.NewTable("table2").
				AddColumns(
					sqlschema.NewEnumColumn("account_type",
						sqlschema.EnumName("account_type"),
						sqlschema.EnumValues("private", "business"),
					),
				),
		)
	buf, err := MarshalSpec(s, hclState)
	require.NoError(t, err)
	const expected = `table "account" {
  schema = sqlschema.test
  column "account_type" {
    null = false
    type = enum.account_type
  }
  column "account_states" {
    null = false
    type = sql("states[]")
  }
}
table "table2" {
  schema = sqlschema.test
  column "account_type" {
    null = false
    type = enum.account_type
  }
}
enum "account_type" {
  schema = sqlschema.test
  values = ["private", "business"]
}
enum "state" {
  schema = sqlschema.test
  values = ["on", "off"]
}
schema "test" {
}
`
	require.EqualValues(t, expected, string(buf))
}

func TestMarshalSpec_TimePrecision(t *testing.T) {
	s := sqlschema.New("test").
		AddTables(
			sqlschema.NewTable("times").
				AddColumns(
					sqlschema.NewTimeColumn("t_time_def", TypeTime),
					sqlschema.NewTimeColumn("t_time_with_time_zone", TypeTimeTZ, sqlschema.TimePrecision(2)),
					sqlschema.NewTimeColumn("t_time_without_time_zone", TypeTime, sqlschema.TimePrecision(2)),
					sqlschema.NewTimeColumn("t_timestamp", TypeTimestamp, sqlschema.TimePrecision(2)),
					sqlschema.NewTimeColumn("t_timestamptz", TypeTimestampTZ, sqlschema.TimePrecision(2)),
				),
		)
	buf, err := MarshalSpec(s, hclState)
	require.NoError(t, err)
	const expected = `table "times" {
  schema = sqlschema.test
  column "t_time_def" {
    null = false
    type = time
  }
  column "t_time_with_time_zone" {
    null = false
    type = timetz(2)
  }
  column "t_time_without_time_zone" {
    null = false
    type = time(2)
  }
  column "t_timestamp" {
    null = false
    type = timestamp(2)
  }
  column "t_timestamptz" {
    null = false
    type = timestamptz(2)
  }
}
schema "test" {
}
`
	require.EqualValues(t, expected, string(buf))
}

func TestTypes(t *testing.T) {
	p := func(i int) *int { return &i }
	for _, tt := range []struct {
		typeExpr string
		expected sqlschema.Type
	}{
		{
			typeExpr: "bit(10)",
			expected: &BitType{T: TypeBit, Size: 10},
		},
		{
			typeExpr: `hstore`,
			expected: &UserDefinedType{T: "hstore"},
		},
		{
			typeExpr: "bit_varying(10)",
			expected: &BitType{T: TypeBitVar, Size: 10},
		},
		{
			typeExpr: "boolean",
			expected: &sqlschema.BoolType{T: TypeBoolean},
		},
		{
			typeExpr: "bool",
			expected: &sqlschema.BoolType{T: TypeBool},
		},
		{
			typeExpr: "bytea",
			expected: &sqlschema.BinaryType{T: TypeBytea},
		},
		{
			typeExpr: "varchar(255)",
			expected: &sqlschema.StringType{T: TypeVarChar, Size: 255},
		},
		{
			typeExpr: "char(255)",
			expected: &sqlschema.StringType{T: TypeChar, Size: 255},
		},
		{
			typeExpr: "character(255)",
			expected: &sqlschema.StringType{T: TypeCharacter, Size: 255},
		},
		{
			typeExpr: "text",
			expected: &sqlschema.StringType{T: TypeText},
		},
		{
			typeExpr: "smallint",
			expected: &sqlschema.IntegerType{T: TypeSmallInt},
		},
		{
			typeExpr: "integer",
			expected: &sqlschema.IntegerType{T: TypeInteger},
		},
		{
			typeExpr: "bigint",
			expected: &sqlschema.IntegerType{T: TypeBigInt},
		},
		{
			typeExpr: "int",
			expected: &sqlschema.IntegerType{T: TypeInt},
		},
		{
			typeExpr: "int2",
			expected: &sqlschema.IntegerType{T: TypeInt2},
		},
		{
			typeExpr: "int4",
			expected: &sqlschema.IntegerType{T: TypeInt4},
		},
		{
			typeExpr: "int8",
			expected: &sqlschema.IntegerType{T: TypeInt8},
		},
		{
			typeExpr: "cidr",
			expected: &NetworkType{T: TypeCIDR},
		},
		{
			typeExpr: "inet",
			expected: &NetworkType{T: TypeInet},
		},
		{
			typeExpr: "macaddr",
			expected: &NetworkType{T: TypeMACAddr},
		},
		{
			typeExpr: "macaddr8",
			expected: &NetworkType{T: TypeMACAddr8},
		},
		{
			typeExpr: "circle",
			expected: &sqlschema.SpatialType{T: TypeCircle},
		},
		{
			typeExpr: "line",
			expected: &sqlschema.SpatialType{T: TypeLine},
		},
		{
			typeExpr: "lseg",
			expected: &sqlschema.SpatialType{T: TypeLseg},
		},
		{
			typeExpr: "box",
			expected: &sqlschema.SpatialType{T: TypeBox},
		},
		{
			typeExpr: "path",
			expected: &sqlschema.SpatialType{T: TypePath},
		},
		{
			typeExpr: "point",
			expected: &sqlschema.SpatialType{T: TypePoint},
		},
		{
			typeExpr: "date",
			expected: &sqlschema.TimeType{T: TypeDate},
		},
		{
			typeExpr: "time",
			expected: typeTime(TypeTime, 6),
		},
		{
			typeExpr: "time(4)",
			expected: typeTime(TypeTime, 4),
		},
		{
			typeExpr: "timetz",
			expected: typeTime(TypeTimeTZ, 6),
		},
		{
			typeExpr: "timestamp",
			expected: typeTime(TypeTimestamp, 6),
		},
		{
			typeExpr: "timestamp(4)",
			expected: typeTime(TypeTimestamp, 4),
		},
		{
			typeExpr: "timestamptz",
			expected: typeTime(TypeTimestampTZ, 6),
		},
		{
			typeExpr: "timestamptz(4)",
			expected: typeTime(TypeTimestampTZ, 4),
		},
		{
			typeExpr: "interval",
			expected: &IntervalType{T: "interval"},
		},
		{
			typeExpr: "interval(1)",
			expected: &IntervalType{T: "interval", Precision: p(1)},
		},
		{
			typeExpr: "second",
			expected: &IntervalType{T: "interval", F: "second"},
		},
		{
			typeExpr: "minute_to_second",
			expected: &IntervalType{T: "interval", F: "minute to second"},
		},
		{
			typeExpr: "minute_to_second(2)",
			expected: &IntervalType{T: "interval", F: "minute to second", Precision: p(2)},
		},
		{
			typeExpr: "real",
			expected: &sqlschema.FloatType{T: TypeReal, Precision: 24},
		},
		{
			typeExpr: "float",
			expected: &sqlschema.FloatType{T: TypeFloat},
		},
		{
			typeExpr: "float(1)",
			expected: &sqlschema.FloatType{T: TypeFloat, Precision: 1},
		},
		{
			typeExpr: "float(25)",
			expected: &sqlschema.FloatType{T: TypeFloat, Precision: 25},
		},
		{
			typeExpr: "float8",
			expected: &sqlschema.FloatType{T: TypeFloat8, Precision: 53},
		},
		{
			typeExpr: "float4",
			expected: &sqlschema.FloatType{T: TypeFloat4, Precision: 24},
		},
		{
			typeExpr: "numeric",
			expected: &sqlschema.DecimalType{T: TypeNumeric},
		},
		{
			typeExpr: "numeric(10)",
			expected: &sqlschema.DecimalType{T: TypeNumeric, Precision: 10},
		},
		{
			typeExpr: "numeric(10, 2)",
			expected: &sqlschema.DecimalType{T: TypeNumeric, Precision: 10, Scale: 2},
		},
		{
			typeExpr: "decimal",
			expected: &sqlschema.DecimalType{T: TypeDecimal},
		},
		{
			typeExpr: "decimal(10)",
			expected: &sqlschema.DecimalType{T: TypeDecimal, Precision: 10},
		},
		{
			typeExpr: "decimal(10,2)",
			expected: &sqlschema.DecimalType{T: TypeDecimal, Precision: 10, Scale: 2},
		},
		{
			typeExpr: "smallserial",
			expected: &SerialType{T: TypeSmallSerial},
		},
		{
			typeExpr: "serial",
			expected: &SerialType{T: TypeSerial},
		},
		{
			typeExpr: "bigserial",
			expected: &SerialType{T: TypeBigSerial},
		},
		{
			typeExpr: "serial2",
			expected: &SerialType{T: TypeSerial2},
		},
		{
			typeExpr: "serial4",
			expected: &SerialType{T: TypeSerial4},
		},
		{
			typeExpr: "serial8",
			expected: &SerialType{T: TypeSerial8},
		},
		{
			typeExpr: "xml",
			expected: &XMLType{T: TypeXML},
		},
		{
			typeExpr: "json",
			expected: &sqlschema.JSONType{T: TypeJSON},
		},
		{
			typeExpr: "jsonb",
			expected: &sqlschema.JSONType{T: TypeJSONB},
		},
		{
			typeExpr: "uuid",
			expected: &UUIDType{T: TypeUUID},
		},
		{
			typeExpr: "money",
			expected: &CurrencyType{T: TypeMoney},
		},
		{
			typeExpr: `sql("int[]")`,
			expected: &ArrayType{Type: &sqlschema.IntegerType{T: "int"}, T: "int[]"},
		},
		{
			typeExpr: `sql("int[2]")`,
			expected: &ArrayType{Type: &sqlschema.IntegerType{T: "int"}, T: "int[]"},
		},
		{
			typeExpr: `sql("text[][]")`,
			expected: &ArrayType{Type: &sqlschema.StringType{T: "text"}, T: "text[]"},
		},
		{
			typeExpr: `sql("integer [3][3]")`,
			expected: &ArrayType{Type: &sqlschema.IntegerType{T: "integer"}, T: "integer[]"},
		},
		{
			typeExpr: `sql("integer  ARRAY[4]")`,
			expected: &ArrayType{Type: &sqlschema.IntegerType{T: "integer"}, T: "integer[]"},
		},
		{
			typeExpr: `sql("integer ARRAY")`,
			expected: &ArrayType{Type: &sqlschema.IntegerType{T: "integer"}, T: "integer[]"},
		},
		{
			typeExpr: `sql("character varying(255) [1][2]")`,
			expected: &ArrayType{Type: &sqlschema.StringType{T: "character varying", Size: 255}, T: "character varying(255)[]"},
		},
		{
			typeExpr: `sql("character varying ARRAY[2]")`,
			expected: &ArrayType{Type: &sqlschema.StringType{T: "character varying"}, T: "character varying[]"},
		},
		{
			typeExpr: `sql("varchar(2) [ 2 ] [  ]")`,
			expected: &ArrayType{Type: &sqlschema.StringType{T: "varchar", Size: 2}, T: "varchar(2)[]"},
		},
	} {
		t.Run(tt.typeExpr, func(t *testing.T) {
			var test sqlschema.Schema
			doc := fmt.Sprintf(`table "test" {
	schema = sqlschema.test
	column "test" {
		null = false
		type = %s
	}
}
schema "test" {
}
`, tt.typeExpr)
			err := EvalHCLBytes([]byte(doc), &test, nil)
			require.NoError(t, err)
			colspec := test.Tables[0].Columns[0]
			require.EqualValues(t, tt.expected, colspec.Type.Type)
			spec, err := MarshalHCL(&test)
			require.NoError(t, err)
			var after sqlschema.Schema
			err = EvalHCLBytes(spec, &after, nil)
			require.NoError(t, err)
			require.EqualValues(t, tt.expected, after.Tables[0].Columns[0].Type.Type)
		})
	}
}

func typeTime(t string, p int) sqlschema.Type {
	return &sqlschema.TimeType{T: t, Precision: &p}
}

func TestParseType_Time(t *testing.T) {
	for _, tt := range []struct {
		typ      string
		expected sqlschema.Type
	}{
		{
			typ:      "timestamptz",
			expected: typeTime(TypeTimestampTZ, 6),
		},
		{
			typ:      "timestamptz(0)",
			expected: typeTime(TypeTimestampTZ, 0),
		},
		{
			typ:      "timestamptz(6)",
			expected: typeTime(TypeTimestampTZ, 6),
		},
		{
			typ:      "timestamp with time zone",
			expected: typeTime(TypeTimestampTZ, 6),
		},
		{
			typ:      "timestamp(1) with time zone",
			expected: typeTime(TypeTimestampTZ, 1),
		},
		{
			typ:      "timestamp",
			expected: typeTime(TypeTimestamp, 6),
		},
		{
			typ:      "timestamp(0)",
			expected: typeTime(TypeTimestamp, 0),
		},
		{
			typ:      "timestamp(6)",
			expected: typeTime(TypeTimestamp, 6),
		},
		{
			typ:      "timestamp without time zone",
			expected: typeTime(TypeTimestamp, 6),
		},
		{
			typ:      "timestamp(1) without time zone",
			expected: typeTime(TypeTimestamp, 1),
		},
		{
			typ:      "time",
			expected: typeTime(TypeTime, 6),
		},
		{
			typ:      "time(3)",
			expected: typeTime(TypeTime, 3),
		},
		{
			typ:      "time without time zone",
			expected: typeTime(TypeTime, 6),
		},
		{
			typ:      "time(3) without time zone",
			expected: typeTime(TypeTime, 3),
		},
		{
			typ:      "timetz",
			expected: typeTime(TypeTimeTZ, 6),
		},
		{
			typ:      "timetz(4)",
			expected: typeTime(TypeTimeTZ, 4),
		},
		{
			typ:      "time with time zone",
			expected: typeTime(TypeTimeTZ, 6),
		},
		{
			typ:      "time(4) with time zone",
			expected: typeTime(TypeTimeTZ, 4),
		},
	} {
		t.Run(tt.typ, func(t *testing.T) {
			typ, err := ParseType(tt.typ)
			require.NoError(t, err)
			require.Equal(t, tt.expected, typ)
		})
	}
}

func TestFormatType_Interval(t *testing.T) {
	p := func(i int) *int { return &i }
	for i, tt := range []struct {
		typ *IntervalType
		fmt string
	}{
		{
			typ: &IntervalType{T: "interval"},
			fmt: "interval",
		},
		{
			typ: &IntervalType{T: "interval", Precision: p(6)},
			fmt: "interval",
		},
		{
			typ: &IntervalType{T: "interval", Precision: p(3)},
			fmt: "interval(3)",
		},
		{
			typ: &IntervalType{T: "interval", F: "DAY"},
			fmt: "interval day",
		},
		{
			typ: &IntervalType{T: "interval", F: "HOUR TO SECOND"},
			fmt: "interval hour to second",
		},
		{
			typ: &IntervalType{T: "interval", F: "HOUR TO SECOND", Precision: p(2)},
			fmt: "interval hour to second(2)",
		},
		{
			typ: &IntervalType{T: "interval", F: "DAY TO HOUR", Precision: p(6)},
			fmt: "interval day to hour",
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			f, err := FormatType(tt.typ)
			require.NoError(t, err)
			require.Equal(t, tt.fmt, f)
		})
	}
}
func TestParseType_Interval(t *testing.T) {
	p := func(i int) *int { return &i }
	for i, tt := range []struct {
		typ    string
		parsed *IntervalType
	}{
		{
			typ:    "interval",
			parsed: &IntervalType{T: "interval", Precision: p(6)},
		},
		{
			typ:    "interval(2)",
			parsed: &IntervalType{T: "interval", Precision: p(2)},
		},
		{
			typ:    "interval day",
			parsed: &IntervalType{T: "interval", F: "day", Precision: p(6)},
		},
		{
			typ:    "interval day to second(2)",
			parsed: &IntervalType{T: "interval", F: "day to second", Precision: p(2)},
		},
		{
			typ:    "interval day to second (2)",
			parsed: &IntervalType{T: "interval", F: "day to second", Precision: p(2)},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			p, err := ParseType(tt.typ)
			require.NoError(t, err)
			require.Equal(t, tt.parsed, p)
		})
	}
}

func TestMarshalRealm(t *testing.T) {
	t1 := sqlschema.NewTable("t1").
		AddColumns(sqlschema.NewIntColumn("id", "int"))
	t2 := sqlschema.NewTable("t2").
		SetComment("Qualified with s1").
		AddColumns(sqlschema.NewIntColumn("oid", "int"))
	t2.AddForeignKeys(sqlschema.NewForeignKey("oid2id").AddColumns(t2.Columns[0]).SetRefTable(t1).AddRefColumns(t1.Columns[0]))

	t3 := sqlschema.NewTable("t3").
		AddColumns(sqlschema.NewIntColumn("id", "int"))
	t4 := sqlschema.NewTable("t2").
		SetComment("Qualified with s2").
		AddColumns(sqlschema.NewIntColumn("oid", "int"))
	t4.AddForeignKeys(sqlschema.NewForeignKey("oid2id").AddColumns(t4.Columns[0]).SetRefTable(t3).AddRefColumns(t3.Columns[0]))
	t5 := sqlschema.NewTable("t5").
		AddColumns(sqlschema.NewIntColumn("oid", "int"))
	t5.AddForeignKeys(sqlschema.NewForeignKey("oid2id1").AddColumns(t5.Columns[0]).SetRefTable(t1).AddRefColumns(t1.Columns[0]))
	// Reference is qualified with s1.
	t5.AddForeignKeys(sqlschema.NewForeignKey("oid2id2").AddColumns(t5.Columns[0]).SetRefTable(t2).AddRefColumns(t2.Columns[0]))

	r := sqlschema.NewRealm(
		sqlschema.New("s1").AddTables(t1, t2),
		sqlschema.New("s2").AddTables(t3, t4, t5),
	)
	got, err := MarshalHCL.MarshalSpec(r)
	require.NoError(t, err)
	require.Equal(
		t,
		`table "t1" {
  schema = sqlschema.s1
  column "id" {
    null = false
    type = int
  }
}
table "s1" "t2" {
  schema  = sqlschema.s1
  comment = "Qualified with s1"
  column "oid" {
    null = false
    type = int
  }
  foreign_key "oid2id" {
    columns     = [column.oid]
    ref_columns = [table.t1.column.id]
  }
}
table "t3" {
  schema = sqlschema.s2
  column "id" {
    null = false
    type = int
  }
}
table "s2" "t2" {
  schema  = sqlschema.s2
  comment = "Qualified with s2"
  column "oid" {
    null = false
    type = int
  }
  foreign_key "oid2id" {
    columns     = [column.oid]
    ref_columns = [table.t3.column.id]
  }
}
table "t5" {
  schema = sqlschema.s2
  column "oid" {
    null = false
    type = int
  }
  foreign_key "oid2id1" {
    columns     = [column.oid]
    ref_columns = [table.t1.column.id]
  }
  foreign_key "oid2id2" {
    columns     = [column.oid]
    ref_columns = [table.s1.t2.column.oid]
  }
}
schema "s1" {
}
schema "s2" {
}
`,
		string(got))
}
