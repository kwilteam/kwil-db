package schemabuilder_test

import (
	"kwil/pkg/execution/sql-builder/schema-builder"
	"kwil/pkg/execution/validator"
	"kwil/pkg/types/databases/clean"
	mocks2 "kwil/pkg/types/databases/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateDDL(t *testing.T) {
	ddl, err := schemabuilder.GenerateDDL(&mocks2.Db1)
	if err != nil {
		t.Errorf("failed to generate ddl: %v", err)
	}

	// validate
	clean.Clean(&mocks2.Db1)
	vldtr := validator.Validator{}
	err = vldtr.Validate(&mocks2.Db1)
	if err != nil {
		t.Errorf("failed to validate database: %v", err)
	}

	for _, stmt := range mocks2.ALL_MOCK_DDL {
		if !assert.Contains(t, ddl, stmt) {
			t.Errorf("missing ddl statement: %v", stmt)
		}
	}
}
