package service

import (
	"kwil/x/execution/dto"
	"kwil/x/execution/validation"
)

func (s *executionService) ValidateDatabase(db *dto.Database) error {
	return validation.ValidateDatabase(db)
}

func (s *executionService) databaseExists(name string) bool {
	_, ok := s.databases[name]
	return ok
}
