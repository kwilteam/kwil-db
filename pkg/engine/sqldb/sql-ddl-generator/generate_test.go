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
