package schemabuilder_test

import (
	"kwil/x/execution/clean"
	"kwil/x/execution/mocks"
	schemabuilder "kwil/x/execution/sql-builder/schema-builder"
	"kwil/x/execution/validation"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateDDL(t *testing.T) {
	ddl, err := schemabuilder.GenerateDDL(&mocks.Db1)
	if err != nil {
		t.Errorf("failed to generate ddl: %v", err)
	}

	// validate
	err = clean.CleanDatabase(&mocks.Db1)
	if err != nil {
		t.Errorf("failed to clean database: %v", err)
	}

	err = validation.ValidateDatabase(&mocks.Db1)
	if err != nil {
		t.Errorf("failed to validate database: %v", err)
	}

	for _, stmt := range mocks.ALL_MOCK_DDL {
		if !assert.Contains(t, ddl, stmt) {
			t.Errorf("missing ddl statement: %v", stmt)
		}
	}
}
