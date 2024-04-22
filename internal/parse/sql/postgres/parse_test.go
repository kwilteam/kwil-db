package postgres_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/internal/parse/sql/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCheckSyntax(t *testing.T) {
	//run `CGO_ENABLED=1 go test ./parse/sql/postgres -v` to test against checkSyntaxCgo
	assert.NoError(t, postgres.CheckSyntax("select 'printme';"))
}
