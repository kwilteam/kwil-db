package schema_test

import (
	"strconv"
	"testing"

	"github.com/kwilteam/kwil-db/internal/schemadef/schema"
	"github.com/stretchr/testify/require"
)

func TestChanges_IndexAddTable(t *testing.T) {
	changes := schema.Changes{
		&schema.AddTable{T: schema.NewTable("users")},
		&schema.DropTable{T: schema.NewTable("posts")},
		&schema.AddTable{T: schema.NewTable("posts")},
		&schema.AddTable{T: schema.NewTable("posts")},
	}
	require.Equal(t, 2, changes.IndexAddTable("posts"))
	require.Equal(t, -1, changes.IndexAddTable("post_tags"))
}

func TestChanges_IndexDropTable(t *testing.T) {
	changes := schema.Changes{
		&schema.DropTable{T: schema.NewTable("users")},
		&schema.AddTable{T: schema.NewTable("posts")},
		&schema.DropTable{T: schema.NewTable("posts")},
	}
	require.Equal(t, 2, changes.IndexDropTable("posts"))
	require.Equal(t, -1, changes.IndexDropTable("post_tags"))
}

func TestChanges_IndexAddColumn(t *testing.T) {
	changes := schema.Changes{
		&schema.AddColumn{C: schema.NewColumn("name")},
		&schema.DropColumn{C: schema.NewColumn("name")},
		&schema.AddColumn{C: schema.NewColumn("name")},
	}
	require.Equal(t, 0, changes.IndexAddColumn("name"))
	require.Equal(t, -1, changes.IndexAddColumn("created_at"))
}

func TestChanges_IndexDropColumn(t *testing.T) {
	changes := schema.Changes{
		&schema.AddColumn{C: schema.NewColumn("name")},
		&schema.DropColumn{C: schema.NewColumn("name")},
		&schema.AddColumn{C: schema.NewColumn("name")},
	}
	require.Equal(t, 1, changes.IndexDropColumn("name"))
	require.Equal(t, -1, changes.IndexDropColumn("created_at"))
}

func TestChanges_IndexAddIndex(t *testing.T) {
	changes := schema.Changes{
		&schema.DropIndex{I: schema.NewIndex("name")},
		&schema.AddIndex{I: schema.NewIndex("created_at")},
		&schema.AddIndex{I: schema.NewIndex("name")},
	}
	require.Equal(t, 2, changes.IndexAddIndex("name"))
	require.Equal(t, -1, changes.IndexAddIndex("age"))
}

func TestChanges_IndexDropIndex(t *testing.T) {
	changes := schema.Changes{
		&schema.AddIndex{I: schema.NewIndex("name")},
		&schema.DropIndex{I: schema.NewIndex("created_at")},
		&schema.DropIndex{I: schema.NewIndex("name")},
	}
	require.Equal(t, 2, changes.IndexDropIndex("name"))
	require.Equal(t, -1, changes.IndexDropIndex("age"))
}

func TestChanges_RemoveIndex(t *testing.T) {
	changes := make(schema.Changes, 0, 5)
	for i := 0; i < 5; i++ {
		changes = append(changes, &schema.AddColumn{C: schema.NewColumn(strconv.Itoa(i))})
	}
	changes.RemoveIndex(0)
	require.Equal(t, 4, len(changes))
	for i := 0; i < 4; i++ {
		require.Equal(t, strconv.Itoa(i+1), changes[i].(*schema.AddColumn).C.Name)
	}
	changes.RemoveIndex(0, 3, 2)
	require.Equal(t, 1, len(changes))
	require.Equal(t, "2", changes[0].(*schema.AddColumn).C.Name)
}

func TestReverseChanges(t *testing.T) {
	tests := []struct {
		input  []schema.SchemaChange
		expect []schema.SchemaChange
	}{
		{
			input: []schema.SchemaChange{
				(*schema.AddColumn)(nil),
			},
			expect: []schema.SchemaChange{
				(*schema.AddColumn)(nil),
			},
		},
		{
			input: []schema.SchemaChange{
				(*schema.AddColumn)(nil),
				(*schema.DropColumn)(nil),
			},
			expect: []schema.SchemaChange{
				(*schema.DropColumn)(nil),
				(*schema.AddColumn)(nil),
			},
		},
		{
			input: []schema.SchemaChange{
				(*schema.AddColumn)(nil),
				(*schema.ModifyColumn)(nil),
				(*schema.DropColumn)(nil),
			},
			expect: []schema.SchemaChange{
				(*schema.DropColumn)(nil),
				(*schema.ModifyColumn)(nil),
				(*schema.AddColumn)(nil),
			},
		},
		{
			input: []schema.SchemaChange{
				(*schema.AddColumn)(nil),
				(*schema.ModifyColumn)(nil),
				(*schema.DropColumn)(nil),
				(*schema.ModifyColumn)(nil),
			},
			expect: []schema.SchemaChange{
				(*schema.ModifyColumn)(nil),
				(*schema.DropColumn)(nil),
				(*schema.ModifyColumn)(nil),
				(*schema.AddColumn)(nil),
			},
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			schema.ReverseChanges(tt.input)
			require.Equal(t, tt.expect, tt.input)
		})
	}
}
