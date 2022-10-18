package pgschema

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"kwil/x/schemadef/sqlschema"

	"github.com/stretchr/testify/require"
)

func TestDiff_TableDiff(t *testing.T) {
	type testcase struct {
		name        string
		from, to    *sqlschema.Table
		wantChanges []sqlschema.SchemaChange
		wantErr     bool
	}
	tests := []testcase{
		{
			name: "no changes",
			from: &sqlschema.Table{Name: "users", Schema: &sqlschema.Schema{Name: "public"}},
			to:   &sqlschema.Table{Name: "users"},
		},
		{
			name: "change primary key columns",
			from: func() *sqlschema.Table {
				t := &sqlschema.Table{Name: "users", Schema: &sqlschema.Schema{Name: "public"}, Columns: []*sqlschema.Column{{Name: "id", Type: &sqlschema.ColumnType{Raw: "int", Type: &sqlschema.IntegerType{T: "int"}}}}}
				t.PrimaryKey = &sqlschema.Index{
					Parts: []*sqlschema.IndexPart{{Column: t.Columns[0]}},
				}
				return t
			}(),
			to:      &sqlschema.Table{Name: "users"},
			wantErr: true,
		},
		func() testcase {
			f := &sqlschema.Table{
				Name: "users",
				Columns: []*sqlschema.Column{
					{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "int"}}},
				},
			}
			f.PrimaryKey = &sqlschema.Index{Parts: []*sqlschema.IndexPart{{Column: f.Columns[0]}}}

			t := &sqlschema.Table{
				Name: "users",
				Columns: []*sqlschema.Column{
					{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "int"}}, Attrs: []sqlschema.Attr{&Identity{Sequence: &Sequence{Start: 1024, Increment: 1}}}},
				},
			}
			t.PrimaryKey = &sqlschema.Index{Parts: []*sqlschema.IndexPart{{Column: t.Columns[0]}}}

			return testcase{
				name: "change identity attributes",
				from: f,
				to:   t,
				wantChanges: []sqlschema.SchemaChange{
					&sqlschema.ModifyColumn{
						From:   &sqlschema.Column{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "int"}}},
						To:     &sqlschema.Column{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "int"}}, Attrs: []sqlschema.Attr{&Identity{Sequence: &Sequence{Start: 1024, Increment: 1}}}},
						Change: sqlschema.ChangeAttr,
					},
				},
			}
		}(),
		{
			name: "drop partition key",
			from: sqlschema.NewTable("logs").
				AddAttrs(&Partition{
					T:     PartitionTypeRange,
					Parts: []*PartitionPart{{Column: "c"}},
				}),
			to:      sqlschema.NewTable("logs"),
			wantErr: true,
		},
		{
			name: "add partition key",
			from: sqlschema.NewTable("logs"),
			to: sqlschema.NewTable("logs").
				AddAttrs(&Partition{
					T:     PartitionTypeRange,
					Parts: []*PartitionPart{{Column: "c"}},
				}),
			wantErr: true,
		},
		{
			name: "change partition key column",
			from: sqlschema.NewTable("logs").
				AddAttrs(&Partition{
					T:     PartitionTypeRange,
					Parts: []*PartitionPart{{Column: "c"}},
				}),
			to: sqlschema.NewTable("logs").
				AddAttrs(&Partition{
					T:     PartitionTypeRange,
					Parts: []*PartitionPart{{Column: "d"}},
				}),
			wantErr: true,
		},
		{
			name: "change partition key type",
			from: sqlschema.NewTable("logs").
				AddAttrs(&Partition{
					T:     PartitionTypeRange,
					Parts: []*PartitionPart{{Column: "c"}},
				}),
			to: sqlschema.NewTable("logs").
				AddAttrs(&Partition{
					T:     PartitionTypeHash,
					Parts: []*PartitionPart{{Column: "c"}},
				}),
			wantErr: true,
		},
		{
			name: "add check",
			from: &sqlschema.Table{Name: "t1", Schema: &sqlschema.Schema{Name: "public"}},
			to:   &sqlschema.Table{Name: "t1", Attrs: []sqlschema.Attr{&sqlschema.Check{Name: "t1_c1_check", Expr: "(c1 > 1)"}}},
			wantChanges: []sqlschema.SchemaChange{
				&sqlschema.AddCheck{
					C: &sqlschema.Check{Name: "t1_c1_check", Expr: "(c1 > 1)"},
				},
			},
		},
		{
			name: "drop check",
			from: &sqlschema.Table{Name: "t1", Attrs: []sqlschema.Attr{&sqlschema.Check{Name: "t1_c1_check", Expr: "(c1 > 1)"}}},
			to:   &sqlschema.Table{Name: "t1"},
			wantChanges: []sqlschema.SchemaChange{
				&sqlschema.DropCheck{
					C: &sqlschema.Check{Name: "t1_c1_check", Expr: "(c1 > 1)"},
				},
			},
		},
		{
			name: "add comment",
			from: &sqlschema.Table{Name: "t1", Schema: &sqlschema.Schema{Name: "public"}},
			to:   &sqlschema.Table{Name: "t1", Attrs: []sqlschema.Attr{&sqlschema.Comment{Text: "t1"}}},
			wantChanges: []sqlschema.SchemaChange{
				&sqlschema.AddAttr{
					A: &sqlschema.Comment{Text: "t1"},
				},
			},
		},
		{
			name: "drop comment",
			from: &sqlschema.Table{Name: "t1", Schema: &sqlschema.Schema{Name: "public"}, Attrs: []sqlschema.Attr{&sqlschema.Comment{Text: "t1"}}},
			to:   &sqlschema.Table{Name: "t1"},
			wantChanges: []sqlschema.SchemaChange{
				&sqlschema.ModifyAttr{
					From: &sqlschema.Comment{Text: "t1"},
					To:   &sqlschema.Comment{Text: ""},
				},
			},
		},
		{
			name: "modify comment",
			from: &sqlschema.Table{Name: "t1", Schema: &sqlschema.Schema{Name: "public"}, Attrs: []sqlschema.Attr{&sqlschema.Comment{Text: "t1"}}},
			to:   &sqlschema.Table{Name: "t1", Attrs: []sqlschema.Attr{&sqlschema.Comment{Text: "t1!"}}},
			wantChanges: []sqlschema.SchemaChange{
				&sqlschema.ModifyAttr{
					From: &sqlschema.Comment{Text: "t1"},
					To:   &sqlschema.Comment{Text: "t1!"},
				},
			},
		},
		func() testcase {
			var (
				s    = sqlschema.New("public")
				from = sqlschema.NewTable("t1").
					SetSchema(s).
					AddColumns(
						sqlschema.NewIntColumn("c1", "int").
							SetGeneratedExpr(&sqlschema.GeneratedExpr{Expr: "1", Type: "STORED"}),
					)
				to = sqlschema.NewTable("t1").
					SetSchema(s).
					AddColumns(
						sqlschema.NewIntColumn("c1", "int"),
					)
			)
			return testcase{
				name: "drop generation expression",
				from: from,
				to:   to,
				wantChanges: []sqlschema.SchemaChange{
					&sqlschema.ModifyColumn{From: from.Columns[0], To: to.Columns[0], Change: sqlschema.ChangeGenerated},
				},
			}
		}(),
		{
			name: "change generation expression",
			from: sqlschema.NewTable("t1").
				SetSchema(sqlschema.New("public")).
				AddColumns(
					sqlschema.NewIntColumn("c1", "int").
						SetGeneratedExpr(&sqlschema.GeneratedExpr{Expr: "1", Type: "STORED"}),
				),
			to: sqlschema.NewTable("t1").
				SetSchema(sqlschema.New("public")).
				AddColumns(
					sqlschema.NewIntColumn("c1", "int").
						SetGeneratedExpr(&sqlschema.GeneratedExpr{Expr: "2", Type: "STORED"}),
				),
			wantErr: true,
		},
		func() testcase {
			var (
				from = &sqlschema.Table{
					Name: "t1",
					Schema: &sqlschema.Schema{
						Name: "public",
					},
					Columns: []*sqlschema.Column{
						{Name: "c1", Type: &sqlschema.ColumnType{Raw: "json", Type: &sqlschema.JSONType{T: "json"}}},
						{Name: "c2", Type: &sqlschema.ColumnType{Raw: "int8", Type: &sqlschema.IntegerType{T: "int8"}}},
					},
				}
				to = &sqlschema.Table{
					Name: "t1",
					Columns: []*sqlschema.Column{
						{
							Name:    "c1",
							Type:    &sqlschema.ColumnType{Raw: "json", Type: &sqlschema.JSONType{T: "json"}, Nullable: true},
							Default: &sqlschema.RawExpr{X: "{}"},
							Attrs:   []sqlschema.Attr{&sqlschema.Comment{Text: "json comment"}},
						},
						{Name: "c3", Type: &sqlschema.ColumnType{Raw: "int", Type: &sqlschema.IntegerType{T: "int"}}},
					},
				}
			)
			return testcase{
				name: "columns",
				from: from,
				to:   to,
				wantChanges: []sqlschema.SchemaChange{
					&sqlschema.ModifyColumn{
						From:   from.Columns[0],
						To:     to.Columns[0],
						Change: sqlschema.ChangeNullability | sqlschema.ChangeComment | sqlschema.ChangeDefault,
					},
					&sqlschema.DropColumn{C: from.Columns[1]},
					&sqlschema.AddColumn{C: to.Columns[1]},
				},
			}
		}(),
		// Modify enum type or values.
		func() testcase {
			var (
				from = sqlschema.NewTable("users").
					SetSchema(sqlschema.New("public")).
					AddColumns(
						sqlschema.NewEnumColumn("state", sqlschema.EnumName("state"), sqlschema.EnumValues("on")),
						sqlschema.NewEnumColumn("enum1", sqlschema.EnumName("enum1"), sqlschema.EnumValues("a")),
						sqlschema.NewEnumColumn("enum3", sqlschema.EnumName("enum3"), sqlschema.EnumValues("a")),
						sqlschema.NewEnumColumn("enum4", sqlschema.EnumName("enum4"), sqlschema.EnumValues("a"), sqlschema.EnumSchema(sqlschema.New("public"))),
					)
				to = sqlschema.NewTable("users").
					SetSchema(sqlschema.New("public")).
					AddColumns(
						// Add value.
						sqlschema.NewEnumColumn("state", sqlschema.EnumName("state"), sqlschema.EnumValues("on", "off")),
						// Change type.
						sqlschema.NewEnumColumn("enum1", sqlschema.EnumName("enum2"), sqlschema.EnumValues("a")),
						// No change as schema is optional.
						sqlschema.NewEnumColumn("enum3", sqlschema.EnumName("enum3"), sqlschema.EnumValues("a"), sqlschema.EnumSchema(sqlschema.New("public"))),
						// Enum type was changed (reside in a different schema).
						sqlschema.NewEnumColumn("enum4", sqlschema.EnumName("enum4"), sqlschema.EnumValues("a"), sqlschema.EnumSchema(sqlschema.New("test"))),
					)
			)
			return testcase{
				name: "enums",
				from: from,
				to:   to,
				wantChanges: []sqlschema.SchemaChange{
					&sqlschema.ModifyColumn{From: from.Columns[0], To: to.Columns[0], Change: sqlschema.ChangeType},
					&sqlschema.ModifyColumn{From: from.Columns[1], To: to.Columns[1], Change: sqlschema.ChangeType},
					&sqlschema.ModifyColumn{From: from.Columns[3], To: to.Columns[3], Change: sqlschema.ChangeType},
				},
			}
		}(),
		// Modify array of type enum.
		func() testcase {
			var (
				from = sqlschema.NewTable("users").
					SetSchema(sqlschema.New("public")).
					AddColumns(
						sqlschema.NewColumn("a1").SetType(&ArrayType{T: "state[]", Type: &sqlschema.EnumType{T: "state", Values: []string{"on"}}}),
						sqlschema.NewColumn("a2").SetType(&ArrayType{T: "state[]", Type: &sqlschema.EnumType{T: "state", Values: []string{"on", "off"}}}),
						sqlschema.NewColumn("a3").SetType(&ArrayType{T: "state[]", Type: &sqlschema.EnumType{T: "state", Values: []string{"on", "off"}}}),
					)
				to = sqlschema.NewTable("users").
					SetSchema(sqlschema.New("public")).
					AddColumns(
						// Add value.
						sqlschema.NewColumn("a1").SetType(&ArrayType{T: "state[]", Type: &sqlschema.EnumType{T: "state", Values: []string{"on", "off"}}}),
						// Drop value.
						sqlschema.NewColumn("a2").SetType(&ArrayType{T: "state[]", Type: &sqlschema.EnumType{T: "state", Values: []string{"on"}}}),
						// Same values.
						sqlschema.NewColumn("a3").SetType(&ArrayType{T: "state[]", Type: &sqlschema.EnumType{T: "state", Values: []string{"on", "off"}}}),
					)
			)
			return testcase{
				name: "enum arrays",
				from: from,
				to:   to,
				wantChanges: []sqlschema.SchemaChange{
					&sqlschema.ModifyColumn{From: from.Columns[0], To: to.Columns[0], Change: sqlschema.ChangeType},
					&sqlschema.ModifyColumn{From: from.Columns[1], To: to.Columns[1], Change: sqlschema.ChangeType},
				},
			}
		}(),
		func() testcase {
			var (
				from = &sqlschema.Table{
					Name: "t1",
					Schema: &sqlschema.Schema{
						Name: "public",
					},
					Columns: []*sqlschema.Column{
						{Name: "c1", Type: &sqlschema.ColumnType{Raw: "json", Type: &sqlschema.JSONType{T: "json"}}, Default: &sqlschema.RawExpr{X: "'{}'"}},
						{Name: "c2", Type: &sqlschema.ColumnType{Raw: "int8", Type: &sqlschema.IntegerType{T: "int8"}}},
						{Name: "c3", Type: &sqlschema.ColumnType{Raw: "int", Type: &sqlschema.IntegerType{T: "int"}}},
					},
				}
				to = &sqlschema.Table{
					Name: "t1",
					Schema: &sqlschema.Schema{
						Name: "public",
					},
					Columns: []*sqlschema.Column{
						{Name: "c1", Type: &sqlschema.ColumnType{Raw: "json", Type: &sqlschema.JSONType{T: "json"}}, Default: &sqlschema.RawExpr{X: "'{}'::json"}},
						{Name: "c2", Type: &sqlschema.ColumnType{Raw: "int8", Type: &sqlschema.IntegerType{T: "int8"}}},
						{Name: "c3", Type: &sqlschema.ColumnType{Raw: "int", Type: &sqlschema.IntegerType{T: "int"}}},
					},
				}
			)
			from.Indexes = []*sqlschema.Index{
				{Name: "c1_index", Unique: true, Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[0]}}},
				{Name: "c2_unique", Unique: true, Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}},
				{Name: "c3_predicate", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}},
				{Name: "c4_predicate", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}, Attrs: []sqlschema.Attr{&IndexPredicate{Predicate: "(c4 <> NULL)"}}},
				{Name: "c4_storage_params", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}, Attrs: []sqlschema.Attr{&IndexStorageParams{PagesPerRange: 4}}},
				{Name: "c5_include_no_change", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}, Attrs: []sqlschema.Attr{&IndexInclude{Columns: columnNames(from.Columns[:1])}}},
				{Name: "c5_include_added", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}},
				{Name: "c5_include_dropped", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}, Attrs: []sqlschema.Attr{&IndexInclude{Columns: columnNames(from.Columns[:1])}}},
			}
			to.Indexes = []*sqlschema.Index{
				{Name: "c1_index", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[0]}}},
				{Name: "c3_unique", Unique: true, Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: to.Columns[1]}}},
				{Name: "c3_predicate", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}, Attrs: []sqlschema.Attr{&IndexPredicate{Predicate: "c3 <> NULL"}}},
				{Name: "c4_predicate", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}, Attrs: []sqlschema.Attr{&IndexPredicate{Predicate: "c4 <> NULL"}}},
				{Name: "c4_storage_params", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}, Attrs: []sqlschema.Attr{&IndexStorageParams{PagesPerRange: 2}}},
				{Name: "c5_include_no_change", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}, Attrs: []sqlschema.Attr{&IndexInclude{Columns: columnNames(from.Columns[:1])}}},
				{Name: "c5_include_added", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}, Attrs: []sqlschema.Attr{&IndexInclude{Columns: columnNames(from.Columns[:1])}}},
				{Name: "c5_include_dropped", Table: from, Parts: []*sqlschema.IndexPart{{Seq: 1, Column: from.Columns[1]}}},
			}
			return testcase{
				name: "indexes",
				from: from,
				to:   to,
				wantChanges: []sqlschema.SchemaChange{
					&sqlschema.ModifyIndex{From: from.Indexes[0], To: to.Indexes[0], Change: sqlschema.ChangeUnique},
					&sqlschema.DropIndex{I: from.Indexes[1]},
					&sqlschema.ModifyIndex{From: from.Indexes[2], To: to.Indexes[2], Change: sqlschema.ChangeAttr},
					&sqlschema.ModifyIndex{From: from.Indexes[4], To: to.Indexes[4], Change: sqlschema.ChangeAttr},
					&sqlschema.ModifyIndex{From: from.Indexes[6], To: to.Indexes[6], Change: sqlschema.ChangeAttr},
					&sqlschema.ModifyIndex{From: from.Indexes[7], To: to.Indexes[7], Change: sqlschema.ChangeAttr},
					&sqlschema.AddIndex{I: to.Indexes[1]},
				},
			}
		}(),
		func() testcase {
			var (
				ref = &sqlschema.Table{
					Name: "t2",
					Schema: &sqlschema.Schema{
						Name: "public",
					},
					Columns: []*sqlschema.Column{
						{Name: "id", Type: &sqlschema.ColumnType{Raw: "int", Type: &sqlschema.IntegerType{T: "int"}}},
						{Name: "ref_id", Type: &sqlschema.ColumnType{Raw: "int", Type: &sqlschema.IntegerType{T: "int"}}},
					},
				}
				from = &sqlschema.Table{
					Name: "t1",
					Schema: &sqlschema.Schema{
						Name: "public",
					},
					Columns: []*sqlschema.Column{
						{Name: "t2_id", Type: &sqlschema.ColumnType{Raw: "int", Type: &sqlschema.IntegerType{T: "int"}}},
					},
				}
				to = &sqlschema.Table{
					Name: "t1",
					Schema: &sqlschema.Schema{
						Name: "public",
					},
					Columns: []*sqlschema.Column{
						{Name: "t2_id", Type: &sqlschema.ColumnType{Raw: "int", Type: &sqlschema.IntegerType{T: "int"}}},
					},
				}
			)
			from.ForeignKeys = []*sqlschema.ForeignKey{
				{Table: from, Columns: from.Columns, RefTable: ref, RefColumns: ref.Columns[:1]},
			}
			to.ForeignKeys = []*sqlschema.ForeignKey{
				{Table: to, Columns: to.Columns, RefTable: ref, RefColumns: ref.Columns[1:]},
			}
			return testcase{
				name: "foreign-keys",
				from: from,
				to:   to,
				wantChanges: []sqlschema.SchemaChange{
					&sqlschema.ModifyForeignKey{
						From:   from.ForeignKeys[0],
						To:     to.ForeignKeys[0],
						Change: sqlschema.ChangeRefColumn,
					},
				},
			}
		}(),
	}
	for _, tt := range tests {
		db, m, err := sqlmock.New()
		require.NoError(t, err)
		mock{m}.version("130000")
		drv, err := Open(db)
		require.NoError(t, err)
		t.Run(tt.name, func(t *testing.T) {
			changes, err := drv.TableDiff(tt.from, tt.to)
			require.Equalf(t, tt.wantErr, err != nil, "got: %v", err)
			require.EqualValues(t, tt.wantChanges, changes)
		})
	}
}

func TestDiff_SchemaDiff(t *testing.T) {
	db, m, err := sqlmock.New()
	require.NoError(t, err)
	mock{m}.version("130000")
	drv, err := Open(db)
	require.NoError(t, err)
	from := &sqlschema.Schema{
		Tables: []*sqlschema.Table{
			{Name: "users"},
			{Name: "pets"},
		},
	}
	to := &sqlschema.Schema{
		Tables: []*sqlschema.Table{
			{
				Name: "users",
				Columns: []*sqlschema.Column{
					{Name: "t2_id", Type: &sqlschema.ColumnType{Raw: "int", Type: &sqlschema.IntegerType{T: "int"}}},
				},
			},
			{Name: "groups"},
		},
	}
	from.Tables[0].Schema = from
	from.Tables[1].Schema = from
	changes, err := drv.SchemaDiff(from, to)
	require.NoError(t, err)
	require.EqualValues(t, []sqlschema.SchemaChange{
		&sqlschema.ModifyTable{T: to.Tables[0], Changes: []sqlschema.SchemaChange{&sqlschema.AddColumn{C: to.Tables[0].Columns[0]}}},
		&sqlschema.DropTable{T: from.Tables[1]},
		&sqlschema.AddTable{T: to.Tables[1]},
	}, changes)
}
