package executor

import (
	"kwil/pkg/databases"
	"kwil/pkg/databases/spec"
	"kwil/pkg/databases/validator"
)

func (s *executor) ValidateDatabase(db *databases.Database[*spec.KwilAny]) error {
	vld := validator.Validator{}
	return vld.Validate(db)
}

func (s *executor) databaseExists(name string) bool {
	_, ok := s.databases[name]
	return ok
}
