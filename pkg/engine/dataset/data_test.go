package dataset_test

import (
	"context"
	"fmt"

	"github.com/cstockton/go-conv"
	"github.com/kwilteam/kwil-db/pkg/engine/dataset"
	"github.com/kwilteam/kwil-db/pkg/engine/types"
	"github.com/kwilteam/kwil-db/pkg/engine/types/testdata"
)

const datasetName = "test"

var (
	test_tables = []*types.Table{
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
					},
				},
				{
					Name: "username",
					Type: types.TEXT,
					Attributes: []*types.Attribute{
						{
							Type: types.NOT_NULL,
						},
						{
							Type: types.UNIQUE,
						},
						{
							Type:  types.MIN_LENGTH,
							Value: "5",
						},
						{
							Type:  types.MAX_LENGTH,
							Value: "32",
						},
					},
				},
				{
					Name: "age",
					Type: types.INT,
					Attributes: []*types.Attribute{
						{
							Type: types.NOT_NULL,
						},
						{
							Type:  types.MIN,
							Value: "13",
						},
						{
							Type:  types.MAX,
							Value: "200",
						},
					},
				},
				{
					Name: "address",
					Type: types.BLOB,
					Attributes: []*types.Attribute{
						{
							Type: types.NOT_NULL,
						},
						{
							Type: types.UNIQUE,
						},
					},
				},
			},
			Indexes: []*types.Index{
				{
					Name:    "username",
					Columns: []string{"username"},
					Type:    types.BTREE,
				},
			},
		},
		&testdata.Table_posts,
	}

	test_procedures = []*types.Procedure{
		procedure_create_user,
		procedure_create_post,
		procedure_get_time,
		procedure_create_post_and_user,
	}

	procedure_create_user = &types.Procedure{
		Name:   "create_user",
		Args:   []string{"$id", "$username", "$age"},
		Public: true,
		Statements: []string{
			"INSERT INTO users (id, username, age, address) VALUES ($id, $username, $age, @caller);",
		},
	}

	procedure_get_time = &types.Procedure{
		Name:   "get_time",
		Args:   []string{},
		Public: true,
		Statements: []string{
			"$time = date_ext.time('current');", // test passing in a string
		},
	}

	procedure_create_post = &types.Procedure{
		Name:   "create_post",
		Args:   []string{"$id", "$title", "$content", "$author_id"},
		Public: true,
		Statements: []string{
			"$post_date = date_ext.time(100);", // test passing in a number
			"INSERT INTO posts (id, title, content, author_id, post_date) VALUES ($id, $title, $content, $author_id, $post_date);",
		},
	}

	procedure_create_post_and_user = &types.Procedure{
		Name:   "create_post_and_user",
		Args:   []string{"$id", "$title", "$content", "$author_id", "$username", "$age"},
		Public: true,
		Statements: []string{
			"create_user($author_id, $username, $age);",
			"create_post($id, $title, $content, $author_id);",
			"SELECT users.username as username, users.address as wallet_address, posts.title as title, posts.content as content FROM users LEFT JOIN posts ON users.id = posts.author_id WHERE users.id = $author_id;",
		},
	}
)

// test extensions

var (
	testAvailableExtensions = []*testExt{
		{
			name: "time",
			mustConfig: map[string]string{
				"format": "unix",
			},
			methodFunc: func(ctx context.Context, method string, args ...any) ([]any, error) {
				if method != "time" {
					return nil, fmt.Errorf("invalid method %s", method)
				}

				if len(args) != 1 {
					return nil, fmt.Errorf("invalid number of args %d", len(args))
				}

				intVal, err := conv.Int(args[0])
				if err == nil {
					if intVal != 100 {
						return nil, fmt.Errorf("invalid arg value %d", intVal)
					}
				} else {
					strVal, ok := args[0].(string)
					if !ok {
						return nil, fmt.Errorf("invalid arg type %T", args[0])
					}

					if strVal != "current" {
						return nil, fmt.Errorf("invalid arg value %s", strVal)
					}
				}

				return []any{uint64(100)}, nil
			},
		},
	}
	testExtensions = []*types.Extension{
		{
			Name: "time",
			Initialization: map[string]string{
				"format": "unix",
			},
			Alias: "date_ext",
		},
	}
)

type methodFunc func(ctx context.Context, method string, args ...any) ([]any, error)

func (f methodFunc) Execute(ctx context.Context, method string, args ...any) ([]any, error) {
	return f(ctx, method, args...)
}

type testExt struct {
	name       string
	mustConfig map[string]string
	methodFunc methodFunc
}

func (t *testExt) Initialize(context.Context, map[string]string) (dataset.InitializedExtension, error) {
	for key, value := range t.mustConfig {
		configedValue, ok := t.mustConfig[key]
		if !ok {
			return nil, fmt.Errorf("missing config key %s", key)
		}

		if configedValue != value {
			return nil, fmt.Errorf("invalid config value for key %s", key)
		}
	}

	return t, nil
}

func (t *testExt) Execute(ctx context.Context, method string, args ...any) ([]any, error) {
	return t.methodFunc(ctx, method, args...)
}
