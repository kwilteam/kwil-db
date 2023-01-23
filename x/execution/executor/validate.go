package executor

import (
	"kwil/x/execution/validator"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
)

func (s *executor) ValidateDatabase(db *databases.Database[anytype.KwilAny]) error {
	vld := validator.Validator{}
	return vld.Validate(db)
}

func (s *executor) databaseExists(name string) bool {
	_, ok := s.databases[name]
	return ok
}
