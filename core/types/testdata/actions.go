package testdata

import "github.com/kwilteam/kwil-db/core/types"

var (
	ActionCreateUser = &types.Action{
		Name:       "create_user",
		Parameters: []string{"$id", "$username", "$age"},
		Public:     true,
		Body:       "INSERT INTO users (id, username, age, address) VALUES ($id, $username, $age, @caller);",
	}

	ActionGetUserByAddress = &types.Action{
		Name:       "get_user_by_address",
		Parameters: []string{"$address"},
		Public:     true,
		Modifiers: []types.Modifier{
			types.ModifierView,
		},
		Body: "SELECT id, username, age FROM users WHERE address = $address;",
	}

	ActionCreatePost = &types.Action{
		Name:       "create_post",
		Parameters: []string{"$id", "$title", "$content", "$post_date"},
		Public:     true,
		Body: `INSERT INTO posts (id, title, content, author_id, post_date) VALUES (
				$id, $title, $content, 
				(SELECT id FROM users WHERE address = @caller LIMIT 1),
				$post_date);`,
	}

	ActionGetPosts = &types.Action{
		Name:       "get_posts",
		Parameters: []string{"$username"},
		Public:     true,
		Modifiers: []types.Modifier{
			types.ModifierView,
		},
		Body: `SELECT p.id as id, p.title as title, p.content as content, p.post_date as post_date, u.username as author FROM posts AS p
				INNER JOIN users AS u ON p.author_id = u.id
				WHERE u.username = $username;`,
	}

	// ActionAdminDeleteUser is a procedure that can only be called by the owner of the schema
	ActionAdminDeleteUser = &types.Action{
		Name:       "admin_delete_user",
		Parameters: []string{"$id"},
		Public:     true,
		Modifiers: []types.Modifier{
			types.ModifierOwner,
		},
		Body: "DELETE FROM users WHERE id = $id;",
	}

	// ActionCallsPrivate is a procedure that calls a private procedure
	ActionCallsPrivate = &types.Action{
		Name:       "calls_private",
		Parameters: []string{},
		Public:     true,
		Body:       "private_procedure();",
	}

	// ActionPrivate is a private procedure
	ActionPrivate = &types.Action{
		Name:       "private_procedure",
		Parameters: []string{},
		Public:     false,
		Body:       "SELECT * FROM users;",
	}

	// ActionRecursive is a recursive procedure that should hit a max stack
	// depth error before using the system's max stack memory, which is fatal.
	ActionRecursive = &types.Action{
		Name:       "recursive_procedure",
		Parameters: []string{"$id", "$a", "$b"},
		Public:     true,
		Body:       "recursive_procedure($id, $a, $b);",
	}

	// ActionRecursiveSneakyA is procedure that calls
	// ProcedureRecursiveSneakyB, which calls ActionRecursiveSneakyA, which
	// calls ProcedureRecursiveSneakyB, which calls...
	ActionRecursiveSneakyA = &types.Action{
		Name:       "recursive_procedure_a",
		Parameters: []string{},
		Public:     true,
		Body:       "recursive_procedure_b();",
	}

	// ActionRecursiveSneakyB is procedure that calls ProcedureRecursiveSneakyA.
	ActionRecursiveSneakyB = &types.Action{
		Name:       "recursive_procedure_b",
		Parameters: []string{},
		Public:     true,
		Body:       "recursive_procedure_a();",
	}
)
