package txsvc

import (
	"context"
	"fmt"
	"math/big"

	"go.uber.org/zap"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/modules/datasets"
	"github.com/kwilteam/kwil-db/pkg/transactions"
)

func (s *Service) EstimatePrice(ctx context.Context, req *txpb.EstimatePriceRequest) (*txpb.EstimatePriceResponse, error) {
	tx := req.Tx
	s.log.Debug("Estimating price", zap.String("payload_type", tx.Body.PayloadType))
	var price *big.Int
	var err error

	switch tx.Body.PayloadType {
	case transactions.PayloadTypeDeploySchema.String():
		price, err = s.priceDeploy(ctx, tx.Body)
	case transactions.PayloadTypeDropSchema.String():
		price, err = s.priceDrop(ctx, tx.Body)
	case transactions.PayloadTypeExecuteAction.String():
		price, err = s.priceAction(ctx, tx.Body)
	case transactions.PayloadTypeValidatorJoin.String():
		price, err = s.priceValidatorJoin(ctx, tx.Body)
	case transactions.PayloadTypeValidatorApprove.String():
		price, err = s.priceValidatorLeave(ctx, tx.Body)
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

func (s *Service) priceDeploy(ctx context.Context, txBody *txpb.Transaction_Body) (*big.Int, error) {
	schema := &transactions.Schema{}
	err := schema.UnmarshalBinary(txBody.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	convertedSchema, err := datasets.ConvertSchemaToEngine(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to convert schema: %w", err)
	}

	return s.engine.PriceDeploy(ctx, convertedSchema)
}

func (s *Service) priceDrop(ctx context.Context, txBody *txpb.Transaction_Body) (*big.Int, error) {
	dropSchema := &transactions.DropSchema{}
	err := dropSchema.UnmarshalBinary(txBody.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal drop schema: %w", err)
	}

	return s.engine.PriceDrop(ctx, dropSchema.DBID)
}

func (s *Service) priceAction(ctx context.Context, txBody *txpb.Transaction_Body) (*big.Int, error) {
	executionBody := &transactions.ActionExecution{}
	err := executionBody.UnmarshalBinary(txBody.Payload)
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

func (s *Service) priceValidatorJoin(ctx context.Context, txBody *txpb.Transaction_Body) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

func (s *Service) priceValidatorLeave(ctx context.Context, txBody *txpb.Transaction_Body) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}
