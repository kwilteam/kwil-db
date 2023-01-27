package validator_test

import (
	"kwil/x/execution/validator"
	"kwil/x/types/databases/mocks"
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
