package executor

import (
	"kwil/pkg/databases"
	"kwil/pkg/databases/validator"
	"kwil/pkg/types/data_types/any_type"
)

func (s *executor) ValidateDatabase(db *databases.Database[anytype.KwilAny]) error {
	vld := validator.Validator{}
	return vld.Validate(db)
}

func (s *executor) databaseExists(name string) bool {
	_, ok := s.databases[name]
	return ok
}
