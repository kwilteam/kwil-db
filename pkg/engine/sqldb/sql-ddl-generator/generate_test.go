package sqlddlgenerator_test

import (
	"strings"
	"testing"
	"unicode"

	"github.com/kwilteam/kwil-db/pkg/engine/dto"
	sqlddlgenerator "github.com/kwilteam/kwil-db/pkg/engine/sqldb/sql-ddl-generator"
)

func TestGenerateDDL(t *testing.T) {
	type args struct {
		table *dto.Table
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "table with primary key attribute",
			args: args{
				table: &dto.Table{
					Name: "test",
					Columns: []*dto.Column{
						{
							Name: "id",
							Type: dto.INT,
							Attributes: []*dto.Attribute{
								{
									Type: "priMAry_key", // testing string case insensitivity
								},
								{
									Type: dto.NOT_NULL,
								},
							},
						},
						{
							Name: "name",
							Type: dto.TEXT,
							Attributes: []*dto.Attribute{
								{
									Type: dto.NOT_NULL,
								},
								{
									Type:  dto.DEFAULT,
									Value: "foo",
								},
							},
						},
					},
					Indexes: []*dto.Index{
						{
							Name:    "test_index",
							Type:    dto.UNIQUE_BTREE,
							Columns: []string{"id", "name"},
						},
					},
				},
			},
			want: []string{
				`CREATE TABLE "test" ("id" INTEGER NOT NULL, "name" TEXT NOT NULL DEFAULT 'foo', PRIMARY KEY ("id")) WITHOUT ROWID, STRICT;`,
				`CREATE UNIQUE INDEX "test_index" ON "test" ("id", "name");`,
			},
		},
		{
			name: "table with composite primary key",
			args: args{
				table: &dto.Table{
					Name: "test",
					Columns: []*dto.Column{
						{
							Name: "id",
							Type: dto.INT,
							Attributes: []*dto.Attribute{
								{
									Type: dto.NOT_NULL,
								},
							},
						},
						{
							Name: "name",
							Type: dto.TEXT,
							Attributes: []*dto.Attribute{
								{
									Type: dto.NOT_NULL,
								},
								{
									Type:  dto.DEFAULT,
									Value: "foo",
								},
							},
						},
					},
					Indexes: []*dto.Index{
						{
							Name:    "test_index",
							Type:    dto.UNIQUE_BTREE,
							Columns: []string{"id", "name"},
						},
						{
							Name:    "CompositePrimaryKey",
							Type:    dto.PRIMARY,
							Columns: []string{"id", "name"},
						},
					},
				},
			},
			want: []string{
				`CREATE TABLE "test" ("id" INTEGER NOT NULL, "name" TEXT NOT NULL DEFAULT 'foo', PRIMARY KEY ("id", "name")) WITHOUT ROWID, STRICT;`,
				`CREATE UNIQUE INDEX "test_index" ON "test" ("id", "name");`,
			},
		},
		{
			name: "table with composite primary key and attribute primary key",
			args: args{
				table: &dto.Table{
					Name: "test",
					Columns: []*dto.Column{
						{
							Name: "id",
							Type: dto.INT,
							Attributes: []*dto.Attribute{
								{
									Type: "primary_key",
								},
								{
									Type: dto.NOT_NULL,
								},
							},
						},
						{
							Name: "name",
							Type: dto.TEXT,
							Attributes: []*dto.Attribute{
								{
									Type: dto.NOT_NULL,
								},
								{
									Type:  dto.DEFAULT,
									Value: "foo",
								},
							},
						},
					},
					Indexes: []*dto.Index{
						{
							Name:    "test_index",
							Type:    dto.UNIQUE_BTREE,
							Columns: []string{"id", "name"},
						},
						{
							Name:    "CompositePrimaryKey",
							Type:    dto.PRIMARY,
							Columns: []string{"id", "name"},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "table with composite primary key and composite index",
			args: args{
				table: &dto.Table{
					Name: "test",
					Columns: []*dto.Column{
						{
							Name: "id",
							Type: dto.INT,
							Attributes: []*dto.Attribute{
								{
									Type: dto.NOT_NULL,
								},
							},
						},
						{
							Name: "name",
							Type: dto.TEXT,
							Attributes: []*dto.Attribute{
								{
									Type: dto.NOT_NULL,
								},
								{
									Type:  dto.DEFAULT,
									Value: "foo",
								},
							},
						},
					},
					Indexes: []*dto.Index{
						{
							Name:    "test_index",
							Type:    dto.UNIQUE_BTREE,
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
				table: &dto.Table{
					Name: "test",
					Columns: []*dto.Column{
						{
							Name: "id",
							Type: dto.INT,
							Attributes: []*dto.Attribute{
								{
									Type: dto.PRIMARY_KEY,
								},
							},
						},
						{
							Name: "name",
							Type: dto.TEXT,
							Attributes: []*dto.Attribute{
								{
									Type:  dto.DEFAULT,
									Value: "foo",
								},
							},
						},
					},
					ForeignKeys: []*dto.ForeignKey{
						{

							ChildKeys:   []string{"name"},
							ParentKeys:  []string{"username"},
							ParentTable: "users",
							Actions: []*dto.ForeignKeyAction{
								{
									On: dto.ON_UPDATE,
									Do: dto.DO_CASCADE,
								},
							},
						},
					},
				},
			},
			want: []string{`CREATE TABLE "test" ("id" INTEGER, "name" TEXT DEFAULT 'foo', FOREIGN KEY ("name") REFERENCES "users"("username") ON UPDATE CASCADE, PRIMARY KEY ("id")) WITHOUT ROWID, STRICT;`},
		},
		{
			name: "table with multiple foreign keys and multiple actions per foreign key",
			args: args{
				table: &dto.Table{
					Name: "table1",
					Columns: []*dto.Column{
						{
							Name: "id",
							Type: dto.INT,
							Attributes: []*dto.Attribute{
								{
									Type: dto.PRIMARY_KEY,
								},
							},
						},
						{
							Name: "name",
							Type: dto.TEXT,
							Attributes: []*dto.Attribute{
								{
									Type:  dto.DEFAULT,
									Value: "foo",
								},
							},
						},
					},
					ForeignKeys: []*dto.ForeignKey{
						{
							ChildKeys:   []string{"name"},
							ParentKeys:  []string{"username"},
							ParentTable: "users",
							Actions: []*dto.ForeignKeyAction{
								{
									On: dto.ON_UPDATE,
									Do: dto.DO_CASCADE,
								},
								{
									On: dto.ON_DELETE,
									Do: dto.DO_SET_DEFAULT,
								},
							},
						},
						{
							ChildKeys:   []string{"id", "name"},
							ParentKeys:  []string{"id", "username"},
							ParentTable: "table2",
							Actions: []*dto.ForeignKeyAction{
								{
									On: dto.ON_UPDATE,
									Do: dto.DO_SET_NULL,
								},
								{
									On: dto.ON_DELETE,
									Do: dto.DO_SET_NULL,
								},
							},
						},
					},
				},
			},
			want: []string{`CREATE TABLE "table1" ("id" INTEGER, "name" TEXT DEFAULT 'foo', FOREIGN KEY ("name") REFERENCES "users"("username") ON UPDATE CASCADE ON DELETE SET DEFAULT, FOREIGN KEY ("id", "name") REFERENCES "table2"("id", "username") ON UPDATE SET NULL ON DELETE SET NULL, PRIMARY KEY ("id")) WITHOUT ROWID, STRICT;`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sqlddlgenerator.GenerateDDL(tt.args.table)
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
			}
		})
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
