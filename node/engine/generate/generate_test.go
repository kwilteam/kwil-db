package generate_test

import (
	"strings"
	"testing"
	"unicode"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/node/engine/generate"
	"github.com/kwilteam/kwil-db/parse/postgres"
	"github.com/stretchr/testify/assert"
)

func TestGenerateDDL(t *testing.T) {
	type args struct {
		table *types.Table
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "table with composite primary key",
			args: args{
				table: &types.Table{
					Name: "test",
					Columns: []*types.Column{
						{
							Name: "id",
							Type: types.IntType,
							Attributes: []*types.Attribute{
								{
									Type: types.NOT_NULL,
								},
							},
						},
						{
							Name: "name",
							Type: types.TextType,
							Attributes: []*types.Attribute{
								{
									Type: types.NOT_NULL,
								},
								{
									Type:  types.DEFAULT,
									Value: "'foo'",
								},
							},
						},
					},
					Indexes: []*types.Index{
						{
							Name:    "test_index",
							Type:    types.UNIQUE_BTREE,
							Columns: []string{"id", "name"},
						},
						{
							Name:    "CompositePrimaryKey",
							Type:    types.PRIMARY,
							Columns: []string{"id", "name"},
						},
					},
				},
			},
			want: []string{
				`CREATE TABLE "dbid"."test" ("id" INT8 NOT NULL, "name" TEXT NOT NULL DEFAULT 'foo', PRIMARY KEY ("id", "name"));`,
				`CREATE UNIQUE INDEX "test_index" ON "dbid"."test" ("id", "name");`,
			},
		},
		{
			name: "table with composite primary key and composite index",
			args: args{
				table: &types.Table{
					Name: "test",
					Columns: []*types.Column{
						{
							Name: "id",
							Type: types.IntType,
							Attributes: []*types.Attribute{
								{
									Type: types.NOT_NULL,
								},
							},
						},
						{
							Name: "name",
							Type: types.TextType,
							Attributes: []*types.Attribute{
								{
									Type: types.NOT_NULL,
								},
								{
									Type:  types.DEFAULT,
									Value: "'foo'",
								},
							},
						},
					},
					Indexes: []*types.Index{
						{
							Name:    "test_index",
							Type:    types.UNIQUE_BTREE,
							Columns: []string{"id", "name"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "table with foreign key on update set cascade",
			args: args{
				table: &types.Table{
					Name: "test",
					Columns: []*types.Column{
						{
							Name: "id",
							Type: types.IntType,
							Attributes: []*types.Attribute{
								{
									Type: types.PRIMARY_KEY,
								},
							},
						},
						{
							Name: "name",
							Type: types.TextType,
							Attributes: []*types.Attribute{
								{
									Type:  types.DEFAULT,
									Value: "'foo'",
								},
							},
						},
					},
					ForeignKeys: []*types.ForeignKey{
						{

							ChildKeys:   []string{"name"},
							ParentKeys:  []string{"username"},
							ParentTable: "users",
							Actions: []*types.ForeignKeyAction{
								{
									On: types.ON_UPDATE,
									Do: types.DO_CASCADE,
								},
							},
						},
					},
				},
			},
			want: []string{`CREATE TABLE "dbid"."test" ("id" INT8, "name" TEXT DEFAULT 'foo', FOREIGN KEY ("name") REFERENCES "dbid"."users"("username") ON UPDATE CASCADE, PRIMARY KEY ("id"));`},
		},
		{
			name: "table with multiple foreign keys and multiple actions per foreign key",
			args: args{
				table: &types.Table{
					Name: "table1",
					Columns: []*types.Column{
						{
							Name: "id",
							Type: types.IntType,
							Attributes: []*types.Attribute{
								{
									Type: types.PRIMARY_KEY,
								},
							},
						},
						{
							Name: "name",
							Type: types.TextType,
							Attributes: []*types.Attribute{
								{
									Type:  types.DEFAULT,
									Value: "'foo'",
								},
							},
						},
					},
					ForeignKeys: []*types.ForeignKey{
						{
							ChildKeys:   []string{"name"},
							ParentKeys:  []string{"username"},
							ParentTable: "users",
							Actions: []*types.ForeignKeyAction{
								{
									On: types.ON_UPDATE,
									Do: types.DO_CASCADE,
								},
								{
									On: types.ON_DELETE,
									Do: types.DO_SET_DEFAULT,
								},
							},
						},
						{
							ChildKeys:   []string{"id", "name"},
							ParentKeys:  []string{"id", "username"},
							ParentTable: "table2",
							Actions: []*types.ForeignKeyAction{
								{
									On: types.ON_UPDATE,
									Do: types.DO_SET_NULL,
								},
								{
									On: types.ON_DELETE,
									Do: types.DO_SET_NULL,
								},
							},
						},
					},
				},
			},
			want: []string{`CREATE TABLE "dbid"."table1" ("id" INT8, "name" TEXT DEFAULT 'foo', FOREIGN KEY ("name") REFERENCES "dbid"."users"("username") ON UPDATE CASCADE ON DELETE SET DEFAULT, FOREIGN KEY ("id", "name") REFERENCES "dbid"."table2"("id", "username") ON UPDATE SET NULL ON DELETE SET NULL, PRIMARY KEY ("id"));`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generate.GenerateDDL("dbid", tt.args.table)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateDDL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("GenerateDDL(): Got and Want have different lengths")
			}

			for i, statement := range got {
				want := tt.want[i]
				if !compareIgnoringWhitespace(statement, want) {
					t.Errorf("GenerateDDL() got = %v, want %v", got, tt.want)
				}

				err = postgres.CheckSyntaxReplaceDollar(statement)
				assert.NoErrorf(t, err, "postgres syntax check failed: %s", err)
			}
		})
	}
}

// there used to be a bug where the DDL generator would edit a table's primary key index,
// if one existed.  It would add an extra '\"' to the beginning and end of each column name.
func Test_PrimaryIndexModification(t *testing.T) {
	testTable := &types.Table{
		Name: "test",
		Columns: []*types.Column{
			{
				Name: "id1",
				Type: types.IntType,
			},
			{
				Name: "id2", // doing this to check composite primary keys
				Type: types.IntType,
			},
		},
		Indexes: []*types.Index{
			{
				Name: "primary",
				Columns: []string{
					"id1",
					"id2",
				},
				Type: types.PRIMARY,
			},
		},
	}

	_, err := generate.GenerateDDL("dbid", testTable)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// check that the primary key index was not modified
	if testTable.Indexes[0].Columns[0] != "id1" {
		t.Errorf("primary key index was modified. Expected 'id1', got '%s'", testTable.Indexes[0].Columns[0])
	}

	if testTable.Indexes[0].Columns[1] != "id2" {
		t.Errorf("primary key index was modified. Expected 'id2', got '%s'", testTable.Indexes[0].Columns[1])
	}
}

func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1 // skip this rune
		}
		return r
	}, s)
}

// compareIgnoringWhitespace compares two strings while ignoring whitespace characters.
func compareIgnoringWhitespace(a, b string) bool {
	aWithoutWhitespace := removeWhitespace(a)
	bWithoutWhitespace := removeWhitespace(b)

	return aWithoutWhitespace == bWithoutWhitespace
}
