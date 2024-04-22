package testdata

import (
	"github.com/kwilteam/kwil-db/core/types"
)

var (
	TableUsers = &types.Table{
		Name: "users",
		Columns: []*types.Column{
			{
				Name: "id",
				Type: types.IntType,
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
				Type: types.TextType,
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
				Type: types.IntType,
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
						Value: "420",
					},
				},
			},
			{
				Name: "address",
				Type: types.TextType,
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
				Name: "age_idx",
				Columns: []string{
					"age",
				},
				Type: types.BTREE,
			},
		},
	}

	TablePosts = &types.Table{
		Name: "posts",
		Columns: []*types.Column{
			{
				Name: "id",
				Type: types.IntType,
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
				Name: "title",
				Type: types.TextType,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
					{
						Type:  types.MAX_LENGTH,
						Value: "300",
					},
				},
			},
			{
				Name: "content",
				Type: types.TextType,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
					{
						Type:  types.MAX_LENGTH,
						Value: "10000",
					},
				},
			},
			{
				Name: "author_id",
				Type: types.IntType,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
				},
			},
			{
				Name: "post_date",
				Type: types.TextType,
				Attributes: []*types.Attribute{
					{
						Type: types.NOT_NULL,
					},
				},
			},
		},
		Indexes: []*types.Index{
			{
				Name: "author_idx",
				Columns: []string{
					"author_id",
				},
				Type: types.BTREE,
			},
			{ // author post names must be unique
				Name: "author_title_idx",
				Columns: []string{
					"author_id",
					"title",
				},
				Type: types.UNIQUE_BTREE,
			},
		},
		ForeignKeys: []*types.ForeignKey{
			{
				ChildKeys: []string{
					"author_id",
				},
				ParentKeys: []string{
					"id",
				},
				ParentTable: "users",
				Actions: []*types.ForeignKeyAction{
					{
						On: types.ON_UPDATE,
						Do: types.DO_CASCADE,
					},
					{
						On: types.ON_DELETE,
						Do: types.DO_CASCADE,
					},
				},
			},
		},
	}
)
