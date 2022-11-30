package execution

import (
	"context"
	"ksl/sqlclient"
	"kwil/_archive/metadata"
)

// TODO: replace this with the implementation from kwil/x/schema/database.go
// it was not finished at the time of this writing
type Input struct {
	Name  string
	Value string
}

type Service interface {
	Execute(context.Context, string, string, string, string, []Input) error
	Read(context.Context, string, string, string, []Input) (*Result, error)
}

type executionService struct {
	md        metadata.Service
	connector metadata.Connector
}

func NewService(md metadata.Service, conn metadata.Connector) *executionService {
	return &executionService{
		md:        md,
		connector: conn,
	}
}

func NewTestService() *executionService {
	return &executionService{
		md:        newMockMdService(),
		connector: metadata.ConnectorFunc(LocalConnectionInfo),
	}
}

func LocalConnectionInfo(wallet string) (string, error) {
	return "postgres://localhost:5432/kwil?sslmode=disable&user=postgres&password=postgres", nil
}

/*
There are a lot of things that need to change here
*/
func (s *executionService) Execute(ctx context.Context, owner, name, caller, query string, inputs []Input) error {
	md, err := s.md.GetMetadata(ctx, metadata.RequestMetadata{Wallet: owner, Database: name})
	if err != nil {
		return err
	}

	q, err := getQuery(&md, &query)
	if err != nil {
		return err
	}

	// we need to verify that the caller is authorized to execute the query
	perms, err := hasPermissions(&md, &caller, &query)
	if err != nil {
		return err
	}
	if !perms {
		return ErrUnauthorized
	}

	// next, we need to verify the types
	_, err = validateTypes(q, &query, inputs)
	if err != nil {
		return err
	}

	// get connection info
	url, err := s.connector.GetConnectionInfo(owner)
	if err != nil {
		return err
	}

	client, err := sqlclient.Open(url)
	if err != nil {
		return err
	}
	defer client.Close()
	// execute
	// _, err = client.ExecContext(q.Statement, ins)
	// if err != nil {
	// 	return err
	// }

	return nil
}

func getQuery(md *metadata.Metadata, query *string) (*metadata.Query, error) {
	var q *metadata.Query
	for _, qs := range md.Queries {
		if qs.Name == *query {
			q = &qs
			break
		}
	}

	if q == nil {
		return nil, ErrQueryNotFound
	}

	return q, nil
}

// TODO: There should be a more efficient way for finding roles and queries than having to iterate through all of them
// this will get the callers role and check if they have permissions to execute the query
func hasPermissions(md *metadata.Metadata, caller, query *string) (bool, error) {
	// search for default role
	var rl *metadata.Role
	for _, role := range md.Roles {
		if role.Name == md.DefaultRole {
			rl = &role
			break
		}
	}

	if rl == nil {
		return false, ErrRoleNotFound
	}
	// now search for the query
	var found bool
	for _, qs := range md.Queries {
		if qs.Name == *query {
			found = true
			break
		}
	}

	if !found {
		return false, ErrRoleDoesNotHavePermission
	}

	return true, nil
}

// this will try to transform the inputs into the types specified in the query
// it will return the query with the inputs transformed
// TODO: implement this
func validateTypes(md *metadata.Query, query *string, inputs []Input) ([]any, error) {
	return []any{"0xAfFDC06cF34aFD7D5801A13d48C92AD39609901D"}, nil
}
