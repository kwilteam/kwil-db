package validation_test

import (
	"kwil/pkg/engine/models/mocks"
	"kwil/pkg/engine/models/validation"
	"testing"
)

func Test_Validate(t *testing.T) {
	db := mocks.MOCK_DATASET1
	err := validation.Validate(&db)
	if err != nil {
		t.Errorf("error validating database: %v", err)
	}

}
