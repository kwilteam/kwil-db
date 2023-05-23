package data

import "github.com/kwilteam/kwil-db/internal/entity"

var (
	TableUsers = &entity.Table{
		Name: "users",
		Columns: []*entity.Column{
			{
				Name: "id",
				Type: "int",
				Attributes: []*entity.Attribute{
					{
						Type: "primary_key",
					},
					{
						Type: "not_null",
					},
				},
			},
			{
				Name: "name",
				Type: "text",
				Attributes: []*entity.Attribute{
					{
						Type: "not_null",
					},
				},
			},
			{
				Name: "age",
				Type: "int",
				Attributes: []*entity.Attribute{
					{
						Type: "not_null",
					},
				},
			},
		},
		Indexes: []*entity.Index{
			{
				Name:    "name_idx",
				Columns: []string{"name"},
				Type:    "unique_btree",
			},
		},
	}

	ActionInsertUser = &entity.Action{
		Name:       "insert_user",
		Inputs:     []string{"$id", "$name", "$age"},
		Statements: []string{"INSERT INTO users (id, name, age) VALUES ($id, $name, $age);"},
		Public:     true,
	}
)
