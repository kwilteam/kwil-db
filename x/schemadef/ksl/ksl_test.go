package ksl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	data := `@syntax "postgres"
datasource db {
	provider = "postgres"
	input = $test
}

/// A collection of users
collection users {
	id: int @serial
	name: string
	age: int @default(21)
	array_field: []int
	enum_field: test_enum

	test = "value"
	@@index([name, age], name="idx_name_age", type="btree")

	index idx_name_age {
		columns = [name, age]
		type = "btree"
	}

	@@foreign_key([name, age], references=[users.id], on_delete="cascade")
}

table posts {
	id: int @serial
	title: string
	user_id: int @references(users.id)
}

role general [default] {
	allow = [get_users]
}

enum test_enum {
	one
	two
	three
}

role admin extends general {
	allow = [add_user, delete_user]
}

query get_users{
	statement = <<SQL
		SELECT * FROM users
	SQL
}

query add_user {
	statement = <<SQL
		INSERT INTO users (name, age) VALUES ($1, $2)
	SQL
}

query delete_user {
	statement = "DELETE FROM users WHERE id = $1"
}
`

	schema, err := Parse("test.kdl", strings.NewReader(data))
	require.Nil(t, err)

	t.Log(Format(schema))
}
