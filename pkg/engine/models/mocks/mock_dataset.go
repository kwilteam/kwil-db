package mocks

import (
	"kwil/pkg/engine/models"
	"kwil/pkg/engine/types"
)

var (
	MOCK_DATASET1 = models.Dataset{
		Owner: "Owner",
		Name:  "name",
		Tables: []*models.Table{
			&MOCK_TABLE1,
			&MOCK_TABLE2,
		},
		Actions: []*models.Action{
			&ACTION_CREATE_POST,
			&ACTION_CREATE_USER,
			&ACTION_GET_USER,
			&ACTION_GET_POSTS_BY_USER,
			&ACTION_GET_POSTS_BY_AGE,
			&ACTION_GET_ALL_USERS,
		},
	}

	MOCK_TABLE1 = models.Table{
		Name: "users",
		Columns: []*models.Column{
			{
				Name: "id",
				Type: types.INT,
				Attributes: []*models.Attribute{
					{
						Type: types.PRIMARY_KEY,
					},
				},
			},
			{
				Name: "username",
				Type: types.TEXT,
				Attributes: []*models.Attribute{
					{
						Type: types.NOT_NULL,
					},
					{
						Type:  types.MAX_LENGTH,
						Value: types.NewMust(100).Bytes(),
					},
				},
			},
			{
				Name: "age",
				Type: types.INT,
				Attributes: []*models.Attribute{
					{
						Type: types.NOT_NULL,
					},
					{
						Type:  types.MIN,
						Value: types.NewMust(13).Bytes(),
					},
				},
			},
			{
				Name: "address",
				Type: types.TEXT,
				Attributes: []*models.Attribute{
					{
						Type: types.NOT_NULL,
					},
				},
			},
		},
		Indexes: []*models.Index{
			{
				Name:    "users_age_index",
				Columns: []string{"age"},
				Type:    types.BTREE,
			},
		},
	}

	MOCK_TABLE2 = models.Table{
		Name: "posts",
		Columns: []*models.Column{
			{
				Name: "id",
				Type: types.INT,
				Attributes: []*models.Attribute{
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
				Attributes: []*models.Attribute{
					{
						Type: types.NOT_NULL,
					},
				},
			},
			{
				Name: "title",
				Type: types.TEXT,
				Attributes: []*models.Attribute{
					{
						Type: types.NOT_NULL,
					},
					{
						Type:  types.MAX_LENGTH,
						Value: types.NewMust(100).Bytes(),
					},
				},
			},
			{
				Name: "body",
				Type: types.TEXT,
				Attributes: []*models.Attribute{
					{
						Type: types.NOT_NULL,
					},
					{
						Type:  types.MAX_LENGTH,
						Value: types.NewMust(1000).Bytes(),
					},
				},
			},
		},
		Indexes: []*models.Index{
			{
				Name:    "posts_user_id_index",
				Columns: []string{"user_id", "title"},
				Type:    types.UNIQUE_BTREE,
			},
		},
	}

	ACTION_CREATE_USER = models.Action{
		Name:   "create_user",
		Public: true,
		Inputs: []string{"$name", "$age"},
		Statements: []string{
			"INSERT INTO users (username, age, address) VALUES ($name, $age, @caller)",
		},
	}

	ACTION_CREATE_POST = models.Action{
		Name:   "create_post",
		Public: true,
		Inputs: []string{"$user_id", "$title", "$body"},
		Statements: []string{
			"INSERT INTO posts (user_id, title, body) VALUES ($user_id, $title, $body)",
		},
	}

	ACTION_GET_USER = models.Action{
		Name:   "get_user",
		Public: true,
		Inputs: []string{"$id"},
		Statements: []string{
			"SELECT * FROM users WHERE id = $id",
		},
	}

	ACTION_GET_ALL_USERS = models.Action{
		Name:   "get_all_users",
		Public: true,
		Inputs: []string{},
		Statements: []string{
			"SELECT * FROM users",
		},
	}

	ACTION_GET_POSTS_BY_USER = models.Action{
		Name:   "get_posts",
		Public: true,
		Inputs: []string{"$user"},
		Statements: []string{
			"SELECT * FROM posts WHERE user_id = (SELECT id FROM users WHERE username = $user)",
		},
	}

	ACTION_GET_POSTS_BY_AGE = models.Action{
		Name:   "get_posts_by_age",
		Public: true,
		Inputs: []string{"$age"},
		Statements: []string{
			"SELECT * FROM posts WHERE user_id IN (SELECT id FROM users WHERE age = $age)",
		},
	}
)
