package manager_test

import (
	"kwil/x/sqlx/manager"
	"testing"
)

func Test_ExportedDB(t *testing.T) {
	db := MockDatabase

	// most issues with this package get caught by compiler, mostly just checking for panics from nil pointers
	db.AsProtobuf()
}

var (
	MockDatabase = manager.ExportedDB{
		Name:        "mydb",
		Owner:       "kwil",
		DefaultRole: "goblin_mode",
		Tables:      MockTables,
		Queries:     MockQueries,
		Roles:       MockRoles,
		Indexes:     MockIndexes,
	}

	MockTables = []*manager.ExportedTable{
		{
			Name: "users",
			Columns: []*manager.ExportedColumn{
				{
					Name: "user_id",
					Type: "int",
					Attributes: []*manager.ExportedAttribute{
						{
							Name:  "primary key",
							Value: "",
						},
					},
				},
				{
					Name: "first_name",
					Type: "varchar",
					Attributes: []*manager.ExportedAttribute{
						{
							Name:  "not null",
							Value: "",
						},
					},
				},
				{
					Name: "last_name",
					Type: "varchar",
					Attributes: []*manager.ExportedAttribute{
						{
							Name:  "not null",
							Value: "",
						},
					},
				},
			},
		},
		{
			Name: "posts",
			Columns: []*manager.ExportedColumn{
				{
					Name: "post_id",
					Type: "int",
					Attributes: []*manager.ExportedAttribute{
						{
							Name:  "primary key",
							Value: "",
						},
					},
				},
				{
					Name: "user_id",
					Type: "int",
					Attributes: []*manager.ExportedAttribute{
						{
							Name:  "not null",
							Value: "",
						},
					},
				},
				{
					Name: "title",
					Type: "varchar",
					Attributes: []*manager.ExportedAttribute{
						{
							Name:  "not null",
							Value: "",
						},
						{
							Name:  "default",
							Value: "Untitled",
						},
					},
				},
				{
					Name: "body",
					Type: "text",
					Attributes: []*manager.ExportedAttribute{
						{
							Name:  "not null",
							Value: "",
						},
					},
				},
			},
		},
	}

	MockQueries = []*manager.ExportedQuery{
		{
			Name:      "new_user",
			Statement: "INSERT INTO users (user_id, first_name, last_name, species) VALUES ($1, $2, $3, $4);",
			Inputs: []*manager.ExportedInput{
				{
					Name:    "user_id",
					Type:    "int",
					Ordinal: 0,
				},
				{
					Name:    "first_name",
					Type:    "varchar",
					Ordinal: 1,
				},
				{
					Name:    "last_name",
					Type:    "varchar",
					Ordinal: 2,
				},
			},
			Defaults: []*manager.ExportedDefault{
				{
					Name:    "species",
					Type:    "varchar",
					Value:   "human",
					Ordinal: 4,
				},
			},
		},
		{
			Name:      "get_post",
			Statement: "SELECT * FROM posts WHERE post_id = $1 AND user_id = $2;",
			Inputs: []*manager.ExportedInput{
				{
					Name:    "post_id",
					Type:    "int",
					Ordinal: 0,
				},
				{
					Name:    "user_id",
					Type:    "int",
					Ordinal: 1,
				},
			},
		},
	}

	MockRoles = []*manager.ExportedRole{
		{
			Name: "goblin_mode",
			Queries: []string{
				"new_user",
				"get_post",
			},
		},
		{
			Name: "bennan_mode",
			Queries: []string{
				"get_post",
			},
		},
	}

	MockIndexes = []*manager.ExportedIndex{
		{
			Name:  "users_first_name_idx",
			Table: "users",
			Columns: []string{
				"first_name",
				"last_name",
			},
			Using: "btree",
		},
		{
			Name:    "post_user_id_idx",
			Table:   "posts",
			Columns: []string{"user_id"},
			Using:   "btree",
		},
	}
)
