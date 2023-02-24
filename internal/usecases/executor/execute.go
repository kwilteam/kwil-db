package executor

import (
	"context"
	"fmt"
	"kwil/pkg/databases"
	"kwil/pkg/databases/executables"
	"kwil/pkg/databases/spec"
	"kwil/pkg/pricing"
	"strings"
)

func (s *executor) ExecuteQuery(ctx context.Context, body *executables.ExecutionBody, caller string) error {
	caller = strings.ToLower(caller)

	db, ok := s.databases[body.Database]
	if !ok {
		return fmt.Errorf("database %s not found", body.Database)
	}

	// check if user can execute
	if !db.CanExecute(caller, body.Query) {
		return fmt.Errorf("user %s cannot execute %s", caller, body.Query)
	}

	// prepare query
	stmt, args, err := db.Prepare(body.Query, caller, body.Inputs)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}

	// execute query
	_, err = s.db.ExecContext(ctx, stmt, args...)
	if err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	return nil
}

func (s *executor) GetQueryCostEstimationInfo(ctx context.Context, body *executables.ExecutionBody, caller string) (*pricing.Params, error) {
	caller = strings.ToLower(caller)
	pricingParams := &pricing.Params{}

	db, ok := s.databases[body.Database]
	if !ok {
		return nil, fmt.Errorf("database %s not found", body.Database)
	}

	// check if user can execute
	if !db.CanExecute(caller, body.Query) {
		return nil, fmt.Errorf("user %s cannot execute %s", caller, body.Query)
	}

	p, err := db.GetPreparer(body.Query, caller, body.Inputs)
	if err != nil {
		return nil, err
	}
	qi, err := db.GetQueryInfo(body.Query, caller, body.Inputs)
	if err != nil {
		return nil, err
	}

	// prepare query for gettign the row count
	tstmt, targs, err := p.PrepareCountAll() //db.PrepareCountAll(body.Query, caller)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statements for query analysis: %w", err)
	}

	// execute query
	rows, err := s.db.QueryContext(ctx, tstmt, targs...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	//var tCount, uCount int64
	for rows.Next() {
		if err := rows.Scan(&pricingParams.T); err != nil {
			return nil, fmt.Errorf("failed to scan query: %w", err)
		}
		fmt.Println(pricingParams.T)
	}
	if pricingParams.T == 0 {
		pricingParams.T = 1
	}

	if qi.QueryType == spec.INSERT {
		pricingParams.U = 1
	} else {
		ustmt, uargs, err := p.PrepareCountUpdated()
		if err != nil {
			return nil, fmt.Errorf("failed to prepare statements for query analysis: %w", err)
		}

		rows, err = s.db.QueryContext(ctx, ustmt, uargs...)
		if err != nil {
			return nil, fmt.Errorf("failed to execute query: %w", err)
		}
		for rows.Next() {
			if err := rows.Scan(&pricingParams.U); err != nil {
				return nil, fmt.Errorf("failed to scan query: %w", err)
			}
			fmt.Println(pricingParams.U)
		}
	}

	dbid, _ := s.GetDBIdentifier(body.Database)
	tableSize, err := s.dao.GetTableSize(ctx, dbid.GetSchemaName(), qi.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get row size: %w", err)
	}

	pricingParams.S = tableSize / pricingParams.T

	pricingParams.I, err = s.dao.GetIndexedColumnCount(ctx, dbid.GetSchemaName(), qi.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get indexed column count: %w", err)
	}

	pricingParams.Q = qi.QueryType
	pricingParams.W = p.GetPredicateLengths()

	return pricingParams, nil
}

func (s *executor) GetQueries(id string) ([]*executables.QuerySignature, error) {
	execInterface, ok := s.databases[id]
	if !ok {
		return nil, fmt.Errorf("database id %s not found", id)
	}
	return execInterface.ListQueries()
}

func (s *executor) GetDBIdentifier(id string) (*databases.DatabaseIdentifier, error) {
	db, ok := s.databases[id]
	if !ok {
		return nil, fmt.Errorf("database id %s not found", id)
	}
	return &databases.DatabaseIdentifier{
		Owner: db.Owner,
		Name:  db.Name,
	}, nil
}
