package data

import "github.com/kwilteam/kwil-db/internal/entity"

var (
	TableUsers = &entity.Table{
		Name: "users",
		Columns: []*entity.Column{
			{
				Name: "id",
				Type: "INT",
				Attributes: []*entity.Attribute{
					{
						Type: "PRIMARY_KEY",
					},
					{
						Type: "NOT_NULL",
					},
				},
			},
			{
				Name: "name",
				Type: "TEXT",
				Attributes: []*entity.Attribute{
					{
						Type: "NOT_NULL",
					},
				},
			},
			{
				Name: "age",
				Type: "INT",
				Attributes: []*entity.Attribute{
					{
						Type: "NOT_NULL",
					},
				},
			},
		},
		Indexes: []*entity.Index{
			{
				Name:    "name_idx",
				Columns: []string{"name"},
				Type:    "UNIQUE_BTREE",
			},
		},
	}

	ActionInsertUser = &entity.Action{
		Name:        "insert_user",
		Inputs:      []string{"$id", "$name", "$age"},
		Statements:  []string{"INSERT INTO users (id, name, age) VALUES ($id, $name, $age);"},
		Public:      true,
		Mutability:  entity.MutabilityView.String(),
		Auxiliaries: []string{entity.AuxiliaryTypeMustSign.String()},
	}
)
