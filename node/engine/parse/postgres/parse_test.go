package postgres_test

import (
	"testing"

	"github.com/kwilteam/kwil-db/node/engine/parse/postgres"
	"github.com/stretchr/testify/assert"
)

func TestCheckSyntax(t *testing.T) {
	assert.NoError(t, postgres.CheckSyntax("select 'printme';"))
}
