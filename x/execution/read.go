package execution

import (
	"context"
	"ksl/sqlclient"
	"kwil/x/schema"
)

func (s *executionService) Read(ctx context.Context, owner, name, query string, inputs []Input) (*Result, error) {
	md, err := s.md.GetMetadata(ctx, schema.RequestMetadata{Wallet: owner, Database: name})
	if err != nil {
		return nil, err
	}
	// first, validate that the query exists in md.Queries
	q, err := getQuery(&md, &query)
	if err != nil {
		return nil, err
	}

	// next, we need to verify the types
	ins, err := validateTypes(q, &query, inputs)
	if err != nil {
		return nil, err
	}

	// get connection info
	url, err := s.connector.GetConnectionInfo(owner)
	if err != nil {
		return nil, err
	}

	client, err := sqlclient.Open(ctx, url)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	// execute
	res, err := client.DB.Query(q.Statement, ins)
	if err != nil {
		return nil, err
	}

	var r Result
	err = r.Load(res)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

//
// ksl type | postgres type
// int			integer
// bigint		bigint
// float		double precision
// decimal		numeric
// string		text
// bool			boolean
// datetime		timestamp
// date			date
// time			time
// bytes		bytea
