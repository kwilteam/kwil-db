package plan

import (
	"context"
	"fmt"
	"testing"

	"github.com/kwilteam/kwil-db/internal/engine/types"
	sqlparser "github.com/kwilteam/kwil-db/parse/sql"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/stretchr/testify/assert"
)

// mockSchemaBase is a mock schema for testing.
// It is equivalent to the following raw schema:
// `database basedb;
//
//	table users {
//	   id int primary notnull,
//	   username text max(20),
//	   age int,
//	}
//
//	table posts {
//	   id int primary notnull,
//	   user_id int,
//	   title text,
//	   content text,
//	   fk (user_id) references users(id)
//	}
//
// `
var mockSchemaBase = &types.Schema{
	Name: "basedb",
	Tables: []*types.Table{
		{
			Name: "users",
			Columns: []*types.Column{
				{
					Name: "id",
					Type: types.INT,
					Attributes: []*types.Attribute{
						{
							Type: types.PRIMARY_KEY,
						},
						{
							Type: types.NOT_NULL,
						},
					},
				},
				{
					Name: "username",
					Type: types.TEXT,
					Attributes: []*types.Attribute{
						{
							Type:  types.MAX_LENGTH,
							Value: "20",
						},
					},
				},
				{
					Name: "age",
					Type: types.INT,
				},
			},
		},
		{
			Name: "posts",
			Columns: []*types.Column{
				{
					Name: "id",
					Type: types.INT,
					Attributes: []*types.Attribute{
						{
							Type: types.PRIMARY_KEY,
						},
						{
							Type: types.NOT_NULL,
						},
					},
				},
				{
					Name: "user_id",
					Type: types.INT,
				},
				{
					Name: "title",
					Type: types.TEXT,
				},
				{
					Name: "content",
					Type: types.TEXT,
				},
			},
			ForeignKeys: []*types.ForeignKey{
				{
					ChildKeys: []string{
						"user_id",
					},
					ParentKeys:  []string{"id"},
					ParentTable: "users",
				},
			},
		},
	},
}

// mockSchemaRefer is a mock schema for testing.
// It is equivalent to the following raw schema:
// `database referdb;
//
//	table replies {
//	   id int primary notnull,
//	   post_id int,
//	   content text,
//	}
//
// `
var mockSchemaRefer = &types.Schema{
	Name: "referdb",
	Tables: []*types.Table{
		{
			Name: "replies",
			Columns: []*types.Column{
				{
					Name: "id",
					Type: types.INT,
					Attributes: []*types.Attribute{
						{
							Type: types.PRIMARY_KEY,
						},
						{
							Type: types.NOT_NULL,
						},
					},
				},
				{
					Name: "post_id",
					Type: types.INT,
				},
				{
					Name: "content",
					Type: types.TEXT,
				},
			},
		},
	},
}

type mockCatalog struct {
	schemas map[string]*types.Schema
}

func (m *mockCatalog) GetSchema(ctx context.Context, dbid string) (*types.Schema, error) {
	if s, ok := m.schemas[dbid]; ok {
		return s, nil
	} else {
		return nil, fmt.Errorf("schema %s not found", dbid)
	}
}

func (m *mockCatalog) TableByName(ctx context.Context, schema *types.Schema, tableName string) (*types.Table,
	error) {
	for _, t := range schema.Tables {
		if t.Name == tableName {
			return t, nil
		}
	}

	return nil, fmt.Errorf("table %s not found", tableName)
}

func TestBuilder_build(t *testing.T) {
	mcat := &mockCatalog{
		schemas: map[string]*types.Schema{
			"basedb":  mockSchemaBase,
			"referdb": mockSchemaRefer,
		},
	}
	ctx := NewBuilderContext(mockSchemaBase.Name)

	////////

	tests := []struct {
		name    string
		args    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "simple select",
			args: "select * from users",
		},
		{
			name: "nested select without alias",
			args: "select username, age from (select * from users)",
		},
		{
			name: "nested select with alias",
			args: "select u.username + 2, u.age from (select * from users) as u",
		},
		{
			name: "select limit",
			args: "select * from users limit 1",
		},
		{
			name: "select limit offset",
			args: "select * from users limit 1 offset 2",
		},
		{
			name: "select limit offset 2",
			args: "select * from users limit 2, 1",
		},
		{
			name: "select order by",
			args: "select * from users order by users.username",
		},
		{
			name: "select order by limit",
			args: "select * from users order by users.username limit 1",
		},
		{
			name: "filter where",
			args: "select * from users where users.username = 'bingo'",
		},
		{
			name: "filter where and",
			args: "select * from users where users.username like 'bingo%' and users.age > 18",
		},
		{
			name: "group by",
			args: "select users.age as age, count(users.age) from users group by users.age",
		},
		{
			name: "group by having",
			args: "select users.age as age, count(users.age) from users group by users.age having users.age > 18",
		},
		{
			name: "join select from star", // fix this
			args: "select u.*, p.title from users as u join posts as p on u.id = p.user_id",
		},
		{
			name: "join select from table",
			args: "select u.name, p.title from users as u join posts as p on u.id = p.user_id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			b := NewBuilder(ctx, mcat)

			ast, err := sqlparser.Parse(tt.args)
			assert.NoError(t1, err)

			got := b.build(ast.(*tree.Select))
			//if (err != nil) != tt.wantErr {
			//	t1.Errorf("Transform() error = %v, wantErr %v", err, tt.wantErr)
			//	return
			//}
			//if !reflect.DeepEqual(got, tt.want) {
			//	t1.Errorf("Transform() got = %v, want %v", got, tt.want)
			//}

			fmt.Println(explain(got))
			fmt.Println("===schema")
			for _, field := range got.Schema().fields {
				fmt.Println(field)
			}
		})
	}
}


func TestDataFrame(t *testing.T) {

}