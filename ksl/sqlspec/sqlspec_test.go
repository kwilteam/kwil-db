package sqlspec_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"ksl/sqlspec"
)

func TestSqlSpecFile(t *testing.T) {
	realm, decDiags := sqlspec.UnmarshalFile("/Users/bryan/go/src/ksl/data/test.kwil")
	require.Empty(t, decDiags)
	_ = realm
}

func TestSqlSpec(t *testing.T) {
	data := `table users {
		id:    int64     @id
		email: string  @unique @size(1024)
		name:  string?
		dob:   date?
		created_at: datetime
		updated_at: datetime

		@@index("users_email_idx", columns=[name], type=BTREE, unique=true)
		@@foreign_key("users_email_fkey", columns=[email], references=[email])
	}

	table posts {
		id:        int64     @id
		title:     string?
		content:   string?
		published: bool    @default(true)
		author_id: int64     @foreign_key(users.id, on_delete="SET NULL")
	}

	enum test_enum {
		one
		two
		three
	}

	role general [default] {
		allow = [add_users, add_posts]
	}

	role admin extends general {
		allow = [delete_post]
	}

	query add_users {
		statement = "INSERT INTO users (email, name) VALUES ($1, $2)"
	}

	query add_posts {
		statement = <<-SQL
			INSERT INTO posts (title, content, published, author_id)
			VALUES ($1, $2, $3, $4)
		SQL
	}

	query delete_post {
		statement = "DELETE FROM posts WHERE id = $1"
	}`

	realm, decDiags := sqlspec.Unmarshal([]byte(data), "test.kwil")
	require.Empty(t, decDiags)
	_ = realm

	// wr := ksl.NewDiagnosticTextWriter(os.Stderr, map[string]*ksl.File{"test.kwil": {Bytes: []byte(data)}}, 120, false)
	// for _, diag := range diags {
	// 	wr.WriteDiagnostic(diag)
	// }
	// require.Empty(t, diags)
}
