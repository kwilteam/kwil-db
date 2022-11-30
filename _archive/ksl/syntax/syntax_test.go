package syntax_test

import (
	"ksl/ast"
	"os"
	"testing"
)

func TestSyntax(t *testing.T) {
	data := `
	@backend("postgres")

	model User {
		id int @id
		name string @unique
		dob date @default("1980-01-01")
		email string @db.varchar(255)
		aliases string[]
		role Role @default(USER)
		posts Post[]
	}

	model Post {
		id int @id
		title string
		author User @ref(fields: author_id, references: id)
		author_id int
	}

	enum Role {
		ADMIN
		USER
	}
`

	sch := ast.ParseString(data, "test.ksl")
	if sch.HasErrors() {
		sch.WriteDiagnostics(os.Stdout, true)
	}
}
