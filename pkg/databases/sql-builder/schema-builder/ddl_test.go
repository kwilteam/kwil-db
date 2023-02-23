package schemabuilder_test

import (
	"fmt"
	"kwil/pkg/databases/clean"
	"kwil/pkg/databases/mocks"
	schemabuilder "kwil/pkg/databases/sql-builder/schema-builder"
	"kwil/pkg/databases/validator"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateDDL(t *testing.T) {
	ddl, err := schemabuilder.GenerateDDL(&mocks.Db1)
	if err != nil {
		t.Errorf("failed to generate ddl: %v", err)
	}

	// validate
	clean.Clean(&mocks.Db1)
	vldtr := validator.Validator{}
	err = vldtr.Validate(&mocks.Db1)
	if err != nil {
		t.Errorf("failed to validate database: %v", err)
	}

	for _, stmt := range mocks.ALL_MOCK_DDL {
		if !assert.Contains(t, ddl, stmt) {
			t.Errorf("missing ddl statement: %v", stmt)
		}
	}

	fmt.Println(ddl)
	panic("")
}
