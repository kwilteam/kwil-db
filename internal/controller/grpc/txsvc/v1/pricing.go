package txsvc

import (
	"context"
	"fmt"
	"math/big"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/modules/datasets"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

func (s *Service) EstimatePrice(ctx context.Context, req *txpb.EstimatePriceRequest) (*txpb.EstimatePriceResponse, error) {
	tx, err := convertTransaction(req.Tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	var price *big.Int

	switch tx.Body.PayloadType {
	case transactions.PayloadTypeDeploySchema:
		price, err = s.priceDeploy(ctx, tx)
	case transactions.PayloadTypeDropSchema:
		price, err = s.priceDrop(ctx, tx)
	case transactions.PayloadTypeExecuteAction:
		price, err = s.priceAction(ctx, tx)
	case transactions.PayloadTypeValidatorJoin:
		price, err = s.priceValidatorJoin(ctx, tx)
	case transactions.PayloadTypeValidatorApprove:
		price, err = s.priceValidatorLeave(ctx, tx)
	default:
		price, err = nil, fmt.Errorf("invalid transaction payload type %s", tx.Body.PayloadType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	return &txpb.EstimatePriceResponse{
		Price: price.String(),
	}, nil
}

func (s *Service) priceDeploy(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	schema := &transactions.Schema{}
	err := schema.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	convertedSchema, err := datasets.ConvertSchemaToEngine(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to convert schema: %w", err)
	}

	return s.engine.PriceDeploy(ctx, convertedSchema)
}

func (s *Service) priceDrop(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	dropSchema := &transactions.DropSchema{}
	err := dropSchema.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal drop schema: %w", err)
	}

	return s.engine.PriceDrop(ctx, dropSchema.DBID)
}

func (s *Service) priceAction(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	executionBody := &transactions.ActionExecution{}
	err := executionBody.UnmarshalBinary(tx.Body.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal action execution: %w", err)
	}

	var tuples [][]any
	for _, tuple := range executionBody.Arguments {
		newTuple := make([]any, len(tuple))
		for i, arg := range tuple {
			newTuple[i] = arg
		}

		tuples = append(tuples, newTuple)
	}

	return s.engine.PriceExecute(ctx, executionBody.DBID, executionBody.Action, tuples)
}

func (s *Service) priceValidatorJoin(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

func (s *Service) priceValidatorLeave(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}
