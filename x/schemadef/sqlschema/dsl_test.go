package sqlschema_test

import (
	"kwil/x/schemadef/sqlschema"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTable_AddColumns(t *testing.T) {
	users := sqlschema.NewTable("users").
		SetComment("users table").
		AddColumns(
			sqlschema.NewBoolColumn("active", "bool"),
			sqlschema.NewNullStringColumn("name", "varchar", sqlschema.StringSize(255)),
		)
	require.Equal(
		t,
		&sqlschema.Table{
			Name: "users",
			Attrs: []sqlschema.Attr{
				&sqlschema.Comment{Text: "users table"},
			},
			Columns: []*sqlschema.Column{
				{Name: "active", Type: &sqlschema.ColumnType{Type: &sqlschema.BoolType{T: "bool"}}},
				{Name: "name", Type: &sqlschema.ColumnType{Nullable: true, Type: &sqlschema.StringType{T: "varchar", Size: 255}}},
			},
		},
		users,
	)
}

func TestSchema_AddTables(t *testing.T) {
	userColumns := []*sqlschema.Column{
		sqlschema.NewIntColumn("id", "int"),
		sqlschema.NewBoolColumn("active", "boolean"),
		sqlschema.NewNullStringColumn("name", "varchar", sqlschema.StringSize(255)),
	}
	users := sqlschema.NewTable("users").
		AddColumns(userColumns...).
		SetPrimaryKey(sqlschema.NewPrimaryKey(userColumns[0])).
		SetComment("users table").
		AddIndexes(
			sqlschema.NewUniqueIndex("unique_name").
				AddColumns(userColumns[2]).
				SetComment("index comment"),
		)
	postColumns := []*sqlschema.Column{
		sqlschema.NewIntColumn("id", "int"),
		sqlschema.NewStringColumn("text", "longtext"),
		sqlschema.NewNullIntColumn("author_id", "int"),
	}
	posts := sqlschema.NewTable("posts").
		AddColumns(postColumns...).
		SetPrimaryKey(sqlschema.NewPrimaryKey(postColumns[0])).
		SetComment("posts table").
		AddForeignKeys(
			sqlschema.NewForeignKey("author_id").
				AddColumns(postColumns[2]).
				SetRefTable(users).
				AddRefColumns(userColumns[0]).
				SetOnDelete(sqlschema.Cascade).
				SetOnUpdate(sqlschema.SetNull),
		)
	require.Equal(
		t,
		func() *sqlschema.Schema {
			s := &sqlschema.Schema{Name: "public"}
			users := &sqlschema.Table{
				Name:   "users",
				Schema: s,
				Attrs: []sqlschema.Attr{
					&sqlschema.Comment{Text: "users table"},
				},
				Columns: []*sqlschema.Column{
					{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "int"}}},
					{Name: "active", Type: &sqlschema.ColumnType{Type: &sqlschema.BoolType{T: "boolean"}}},
					{Name: "name", Type: &sqlschema.ColumnType{Nullable: true, Type: &sqlschema.StringType{T: "varchar", Size: 255}}},
				},
			}
			s.Tables = append(s.Tables, users)
			users.PrimaryKey = &sqlschema.Index{Unique: true, Parts: []*sqlschema.IndexPart{{Column: users.Columns[0]}}}
			users.PrimaryKey.Table = users
			users.Columns[0].Indexes = append(users.Columns[0].Indexes, users.PrimaryKey)
			users.Indexes = append(users.Indexes, &sqlschema.Index{
				Name:   "unique_name",
				Unique: true,
				Parts:  []*sqlschema.IndexPart{{Column: users.Columns[2]}},
				Attrs:  []sqlschema.Attr{&sqlschema.Comment{Text: "index comment"}},
			})
			users.Indexes[0].Table = users
			users.Columns[2].Indexes = users.Indexes

			posts := &sqlschema.Table{
				Name:   "posts",
				Schema: s,
				Attrs: []sqlschema.Attr{
					&sqlschema.Comment{Text: "posts table"},
				},
				Columns: []*sqlschema.Column{
					{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "int"}}},
					{Name: "text", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "longtext"}}},
					{Name: "author_id", Type: &sqlschema.ColumnType{Nullable: true, Type: &sqlschema.IntegerType{T: "int"}}},
				},
			}
			s.Tables = append(s.Tables, posts)
			posts.PrimaryKey = &sqlschema.Index{Unique: true, Parts: []*sqlschema.IndexPart{{Column: posts.Columns[0]}}}
			posts.PrimaryKey.Table = posts
			posts.Columns[0].Indexes = append(posts.Columns[0].Indexes, posts.PrimaryKey)
			posts.ForeignKeys = append(posts.ForeignKeys, &sqlschema.ForeignKey{
				Name:       "author_id",
				Table:      posts,
				Columns:    posts.Columns[2:],
				RefTable:   users,
				RefColumns: users.Columns[0:1],
				OnDelete:   sqlschema.Cascade,
				OnUpdate:   sqlschema.SetNull,
			})
			posts.Columns[2].ForeignKeys = posts.ForeignKeys
			return s
		}(),
		sqlschema.New("public").AddTables(users, posts),
	)
}

func TestSchema_SetCharset(t *testing.T) {
	s := sqlschema.New("public")
	require.Empty(t, s.Attrs)
	s.SetCharset("utf8mb4")
	require.Len(t, s.Attrs, 1)
	require.Equal(t, &sqlschema.Charset{V: "utf8mb4"}, s.Attrs[0])
	s.SetCharset("latin1")
	require.Len(t, s.Attrs, 1)
	require.Equal(t, &sqlschema.Charset{V: "latin1"}, s.Attrs[0])
	s.UnsetCharset()
	require.Empty(t, s.Attrs)
}

func TestSchema_SetCollation(t *testing.T) {
	s := sqlschema.New("public")
	require.Empty(t, s.Attrs)
	s.SetCollation("utf8mb4_general_ci")
	require.Len(t, s.Attrs, 1)
	require.Equal(t, &sqlschema.Collation{V: "utf8mb4_general_ci"}, s.Attrs[0])
	s.SetCollation("latin1_swedish_ci")
	require.Len(t, s.Attrs, 1)
	require.Equal(t, &sqlschema.Collation{V: "latin1_swedish_ci"}, s.Attrs[0])
	s.UnsetCollation()
	require.Empty(t, s.Attrs)
}

func TestSchema_SetComment(t *testing.T) {
	s := sqlschema.New("public")
	require.Empty(t, s.Attrs)
	s.SetComment("1")
	require.Len(t, s.Attrs, 1)
	require.Equal(t, &sqlschema.Comment{Text: "1"}, s.Attrs[0])
	s.SetComment("2")
	require.Len(t, s.Attrs, 1)
	require.Equal(t, &sqlschema.Comment{Text: "2"}, s.Attrs[0])
}
