package testdata

import types "github.com/kwilteam/kwil-db/common"

var (
	ProcedureCreateUser = &types.Procedure{
		Name:   "create_user",
		Args:   []string{"$id", "$username", "$age"},
		Public: true,
		Statements: []string{
			"INSERT INTO users (id, username, age, address) VALUES ($id, $username, $age, @caller);",
		},
	}

	ProcedureGetUserByAddress = &types.Procedure{
		Name:   "get_user_by_address",
		Args:   []string{"$address"},
		Public: true,
		Modifiers: []types.Modifier{
			types.ModifierView,
		},
		Statements: []string{
			"SELECT id, username, age FROM users WHERE address = $address;",
		},
	}

	ProcedureCreatePost = &types.Procedure{
		Name:   "create_post",
		Args:   []string{"$id", "$title", "$content", "$post_date"},
		Public: true,
		Statements: []string{
			`INSERT INTO posts (id, title, content, author_id, post_date) VALUES (
				$id, $title, $content, 
				(SELECT id FROM users WHERE address = @caller LIMIT 1),
				$post_date);`,
		},
	}

	ProcedureGetPosts = &types.Procedure{
		Name:   "get_posts",
		Args:   []string{"$username"},
		Public: true,
		Modifiers: []types.Modifier{
			types.ModifierView,
		},
		Statements: []string{
			`SELECT p.id as id, p.title as title, p.content as content, p.post_date as post_date, u.username as author FROM posts AS p
				INNER JOIN users AS u ON p.author_id = u.id
				WHERE u.username = $username;`,
		},
	}

	// ProcedureAdminDeleteUser is a procedure that can only be called by the owner of the schema
	ProcedureAdminDeleteUser = &types.Procedure{
		Name:   "admin_delete_user",
		Args:   []string{"$id"},
		Public: true,
		Modifiers: []types.Modifier{
			types.ModifierOwner,
		},
		Statements: []string{
			"DELETE FROM users WHERE id = $id;",
		},
	}

	// ProcedureCallsPrivate is a procedure that calls a private procedure
	ProcedureCallsPrivate = &types.Procedure{
		Name:   "calls_private",
		Args:   []string{},
		Public: true,
		Statements: []string{
			"private_procedure();",
		},
	}

	// ProcedurePrivate is a private procedure
	ProcedurePrivate = &types.Procedure{
		Name:   "private_procedure",
		Args:   []string{},
		Public: false,
		Statements: []string{
			"SELECT * FROM users;",
		},
	}

	// ProcedureRecursive is a recursive procedure that should hit a max stack
	// depth error before using the system's max stack memory, which is fatal.
	ProcedureRecursive = &types.Procedure{
		Name:   "recursive_procedure",
		Args:   []string{"$id", "$a", "$b"},
		Public: true,
		Statements: []string{
			"recursive_procedure($id, $a, $b);",
		},
	}

	// ProcedureRecursiveSneakyA is procedure that calls
	// ProcedureRecursiveSneakyB, which calls ProcedureRecursiveSneakyA, which
	// calls ProcedureRecursiveSneakyB, which calls...
	ProcedureRecursiveSneakyA = &types.Procedure{
		Name:   "recursive_procedure_a",
		Args:   []string{},
		Public: true,
		Statements: []string{
			"recursive_procedure_b();",
		},
	}

	// ProcedureRecursiveSneakyB is procedure that calls ProcedureRecursiveSneakyA.
	ProcedureRecursiveSneakyB = &types.Procedure{
		Name:   "recursive_procedure_b",
		Args:   []string{},
		Public: true,
		Statements: []string{
			"recursive_procedure_a();",
		},
	}
)
