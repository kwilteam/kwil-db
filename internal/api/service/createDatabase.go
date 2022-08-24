package service

import (
	"context"
	"github.com/kwilteam/kwil-db/pkg/types"
)

// Service Function for CreateDatabase
func (s *Service) CreateDatabase(ctx context.Context, db *types.CreateDatabase) error {

	/*
		Service Function for CreateDatabase

		First, we need to check the cost associated with creating a database.
		Right now, this will be a static cost set in the config file. config.cost.database.create
	*/

	return nil
}
