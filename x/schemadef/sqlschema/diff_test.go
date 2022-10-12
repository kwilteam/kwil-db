package sqlschema_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"kwil/x/schemadef/sqlschema"
)

func TestChanges_IndexAddTable(t *testing.T) {
	changes := sqlschema.Changes{
		&sqlschema.AddTable{T: sqlschema.NewTable("users")},
		&sqlschema.DropTable{T: sqlschema.NewTable("posts")},
		&sqlschema.AddTable{T: sqlschema.NewTable("posts")},
		&sqlschema.AddTable{T: sqlschema.NewTable("posts")},
	}
	require.Equal(t, 2, changes.IndexAddTable("posts"))
	require.Equal(t, -1, changes.IndexAddTable("post_tags"))
}

func TestChanges_IndexDropTable(t *testing.T) {
	changes := sqlschema.Changes{
		&sqlschema.DropTable{T: sqlschema.NewTable("users")},
		&sqlschema.AddTable{T: sqlschema.NewTable("posts")},
		&sqlschema.DropTable{T: sqlschema.NewTable("posts")},
	}
	require.Equal(t, 2, changes.IndexDropTable("posts"))
	require.Equal(t, -1, changes.IndexDropTable("post_tags"))
}

func TestChanges_IndexAddColumn(t *testing.T) {
	changes := sqlschema.Changes{
		&sqlschema.AddColumn{C: sqlschema.NewColumn("name")},
		&sqlschema.DropColumn{C: sqlschema.NewColumn("name")},
		&sqlschema.AddColumn{C: sqlschema.NewColumn("name")},
	}
	require.Equal(t, 0, changes.IndexAddColumn("name"))
	require.Equal(t, -1, changes.IndexAddColumn("created_at"))
}

func TestChanges_IndexDropColumn(t *testing.T) {
	changes := sqlschema.Changes{
		&sqlschema.AddColumn{C: sqlschema.NewColumn("name")},
		&sqlschema.DropColumn{C: sqlschema.NewColumn("name")},
		&sqlschema.AddColumn{C: sqlschema.NewColumn("name")},
	}
	require.Equal(t, 1, changes.IndexDropColumn("name"))
	require.Equal(t, -1, changes.IndexDropColumn("created_at"))
}

func TestChanges_IndexAddIndex(t *testing.T) {
	changes := sqlschema.Changes{
		&sqlschema.DropIndex{I: sqlschema.NewIndex("name")},
		&sqlschema.AddIndex{I: sqlschema.NewIndex("created_at")},
		&sqlschema.AddIndex{I: sqlschema.NewIndex("name")},
	}
	require.Equal(t, 2, changes.IndexAddIndex("name"))
	require.Equal(t, -1, changes.IndexAddIndex("age"))
}

func TestChanges_IndexDropIndex(t *testing.T) {
	changes := sqlschema.Changes{
		&sqlschema.AddIndex{I: sqlschema.NewIndex("name")},
		&sqlschema.DropIndex{I: sqlschema.NewIndex("created_at")},
		&sqlschema.DropIndex{I: sqlschema.NewIndex("name")},
	}
	require.Equal(t, 2, changes.IndexDropIndex("name"))
	require.Equal(t, -1, changes.IndexDropIndex("age"))
}

func TestChanges_RemoveIndex(t *testing.T) {
	changes := make(sqlschema.Changes, 0, 5)
	for i := 0; i < 5; i++ {
		changes = append(changes, &sqlschema.AddColumn{C: sqlschema.NewColumn(strconv.Itoa(i))})
	}
	changes.RemoveIndex(0)
	require.Equal(t, 4, len(changes))
	for i := 0; i < 4; i++ {
		require.Equal(t, strconv.Itoa(i+1), changes[i].(*sqlschema.AddColumn).C.Name)
	}
	changes.RemoveIndex(0, 3, 2)
	require.Equal(t, 1, len(changes))
	require.Equal(t, "2", changes[0].(*sqlschema.AddColumn).C.Name)
}

func TestReverseChanges(t *testing.T) {
	tests := []struct {
		input  []sqlschema.SchemaChange
		expect []sqlschema.SchemaChange
	}{
		{
			input: []sqlschema.SchemaChange{
				(*sqlschema.AddColumn)(nil),
			},
			expect: []sqlschema.SchemaChange{
				(*sqlschema.AddColumn)(nil),
			},
		},
		{
			input: []sqlschema.SchemaChange{
				(*sqlschema.AddColumn)(nil),
				(*sqlschema.DropColumn)(nil),
			},
			expect: []sqlschema.SchemaChange{
				(*sqlschema.DropColumn)(nil),
				(*sqlschema.AddColumn)(nil),
			},
		},
		{
			input: []sqlschema.SchemaChange{
				(*sqlschema.AddColumn)(nil),
				(*sqlschema.ModifyColumn)(nil),
				(*sqlschema.DropColumn)(nil),
			},
			expect: []sqlschema.SchemaChange{
				(*sqlschema.DropColumn)(nil),
				(*sqlschema.ModifyColumn)(nil),
				(*sqlschema.AddColumn)(nil),
			},
		},
		{
			input: []sqlschema.SchemaChange{
				(*sqlschema.AddColumn)(nil),
				(*sqlschema.ModifyColumn)(nil),
				(*sqlschema.DropColumn)(nil),
				(*sqlschema.ModifyColumn)(nil),
			},
			expect: []sqlschema.SchemaChange{
				(*sqlschema.ModifyColumn)(nil),
				(*sqlschema.DropColumn)(nil),
				(*sqlschema.ModifyColumn)(nil),
				(*sqlschema.AddColumn)(nil),
			},
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			sqlschema.ReverseChanges(tt.input)
			require.Equal(t, tt.expect, tt.input)
		})
	}
}
