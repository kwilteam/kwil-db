package data

import "github.com/kwilteam/kwil-db/pkg/engine/dto"

var (
	TableUsers = &dto.Table{
		Name: "users",
		Columns: []*dto.Column{
			{
				Name: "id",
				Type: dto.INT,
				Attributes: []*dto.Attribute{
					{
						Type: dto.PRIMARY_KEY,
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
				},
			},
			{
				Name: "age",
				Type: dto.INT,
				Attributes: []*dto.Attribute{
					{
						Type: dto.NOT_NULL,
					},
				},
			},
		},
	}

	ActionInsertUser = &dto.Action{
		Name:       "insert_user",
		Inputs:     []string{"$id", "$name", "$age"},
		Statements: []string{"INSERT INTO users (id, name, age) VALUES ($id, $name, $age);"},
		Public:     true,
	}

	ExtensionTest = &dto.ExtensionInitialization{
		Name: "test",
		Metadata: map[string]string{
			"foo": "bar",
		},
	}
)
