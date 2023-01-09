package executor

import (
	"kwil/x/execution/validation"
	"kwil/x/types/databases"
)

func (s *executor) ValidateDatabase(db *databases.Database) error {
	return validation.ValidateDatabase(db)
}

func (s *executor) databaseExists(name string) bool {
	_, ok := s.databases[name]
	return ok
}
