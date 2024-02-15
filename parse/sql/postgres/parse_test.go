package postgres_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/parse/sql/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCheckSyntax(t *testing.T) {
	//run `go test -tags enablecgo ./...` to test against checkSyntaxCgo
	assert.NoError(t, postgres.CheckSyntax("select 1;"))
}
