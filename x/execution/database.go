package execution

import (
	"context"
	"kwil/x/schema"
)

// TODO: replace this with the implementatiopn from kwil/x/schema/database.go
// it was not finished at the time of this writing
type Input struct {
	Name  string
	Value string
}

type ExecutionResponse struct {
}

type Service interface {
	Execute(ctx context.Context, owner, name, query string, inputs []Input) (*ExecutionResponse, error)
}

type executionService struct {
	md schema.Service
}

func NewExecutionService() *executionService {
	return &executionService{}
}

/*
There are a lot of things that need to change here
*/
func (s *executionService) Execute(ctx context.Context, owner, name, caller, query string, inputs []Input) (*ExecutionResponse, error) {
	md, err := s.md.GetDatabase(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	var q *schema.Query
	// first, validate that the query exists in md.Queries
	for _, qs := range md.Queries {
		if q.Name == query {
			q = qs
			break
		}
	}

	if q == nil {
		return nil, ErrQueryNotFound
	}

	// we need to verify that the caller is authorized to execute the query
	perms, err := hasPermissions(&md, &caller, &query)
	if err != nil {
		return nil, err
	}
	if !perms {
		return nil, ErrUnauthorized
	}

	// next, we need to verify the types
}

// TODO: implement this
// this will get the callers role and check if they have permissions to execute the query
func hasPermissions(md *schema.Database, caller, query *string) (bool, error) {
	return true, nil
}

// this will try to transform the inputs into the types specified in the query
// it will return the query with the inputs transformed
func validateTypes(md *schema.Query, query *string, inputs []Input) (bool, error) {
	return true, nil
}
