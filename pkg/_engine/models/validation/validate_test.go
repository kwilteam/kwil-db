package validation_test

import (
	"github.com/kwilteam/kwil-db/pkg/engine/models/mocks"
	"github.com/kwilteam/kwil-db/pkg/engine/models/validation"
	"testing"
)

func Test_Validate(t *testing.T) {
	db := mocks.MOCK_DATASET1
	err := validation.Validate(&db)
	if err != nil {
		t.Errorf("error validating database: %v", err)
	}

}
