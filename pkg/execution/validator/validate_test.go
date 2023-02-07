package validator_test

import (
	"kwil/pkg/execution/validator"
	"kwil/pkg/types/databases/mocks"
	"testing"
)

func Test_Validate(t *testing.T) {
	db := mocks.Db1

	v := validator.Validator{}
	err := v.Validate(&db)
	if err != nil {
		t.Errorf("error validating database: %v", err)
	}
}
