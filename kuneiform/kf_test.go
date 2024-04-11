package kuneiform_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/internal/engine/execution"
	"github.com/kwilteam/kwil-db/internal/engine/procedures"
	"github.com/kwilteam/kwil-db/kuneiform"
	"github.com/stretchr/testify/require"
)

// TODO: we should probably delete this file.
// I am using this to allow me to test something testing in the integration
// test with quicker iteraqtions by not having to run docker.

var schema = `
database users;

// This schema is meant to test the changes introduced in Kwil v0.8.
// This includes:
// - new types such as uuid, bool, and arrays
// - new (optional) syntax for constraints and foreign key actions
// - procedures, and the procedural language

table users {
    id uuid primary key,
    name text maxlen(30) not null unique,
    age int max(150),
    address text not null unique,
    
    // post_count tracks the amount of posts a user has
    // it is kept as a separate table for quick lookup times
    post_count int not null default(0),
    #age_idx index(age)
}

table posts {
    id uuid primary key,
    content text maxlen(300) not null,
    author_id uuid not null,
    // post_num is the index of the post for the author
    // the first post for a user will be 1, then 2, 3, etc.
    // it is used for chronological ordering
    post_num int not null,
    foreign key (author_id) references users(id) on update cascade on delete cascade,
    #author_idx index(author_id)
}

// create_user creates a user in the database.
// It is assigned a unique uuid.
procedure create_user($name text, $age int) public {
    // we will generate a uuid from the txid
    INSERT INTO users (id, name, age, address)
    VALUES (uuid_generate_v5('985b93a4-2045-44d6-bde4-442a4e498bc6'::uuid, @txid),
        $name,
        $age,
        @caller
    );
}

// get_user gets a user id, age, and address from a username
procedure get_user($name text) public view returns (id uuid, age int, address text, post_count int) {
    for $row in SELECT id, age, address, post_count FROM users WHERE name = $name {
        return $row.id, $row.age, $row.address, $row.post_count; // will return on the first iteration
    }

    error(format('user "%s" not found', $name));
}

// increment_post_count increments a user's post count, identified by the owner.
// it returns the user id and the new post count
procedure increment_post_count($address text) private returns (id uuid, post_count int) {
    for $row in UPDATE users SET post_count = post_count+1 WHERE address = $address returning id, post_count {
        return $row.id, $row.post_count;
    }

    error(format('user owned by "%s" not found', $address));
}

// get_users by age gets users of a certain age
procedure get_users_by_age($age int) public view returns table(id uuid, name text, address text, post_count int) {
    return SELECT id, name, address, post_count FROM users WHERE age = $age;
}

// create_post creates a post
procedure create_post($content text) public {
    $user_id uuid;
    $post_count int;
    $user_id, $post_count := increment_post_count(@caller);

    INSERT INTO posts (id, content, author_id, post_num)
    VALUES (
        uuid_generate_v5('985b93a4-2045-44d6-bde4-442a4e498bc6'::uuid, @txid),
        $content,
        $user_id,
        $post_count
    );
}`

func Test_KF(t *testing.T) {
	schema, err := kuneiform.Parse(schema)
	require.NoError(t, err)

	_ = schema

	_, err = procedures.GeneratePLPGSQL(schema, "pgschema", "ctx", execution.PgSessionVars)
	require.NoError(t, err)
}
