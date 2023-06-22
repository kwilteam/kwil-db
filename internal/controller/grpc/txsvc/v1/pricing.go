package txsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
)

func (s *Service) EstimatePrice(ctx context.Context, req *txpb.EstimatePriceRequest) (*txpb.EstimatePriceResponse, error) {
	tx, err := convertTx(req.Tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	switch tx.PayloadType {
	case kTx.DEPLOY_DATABASE:
		return handlePricing(s.priceDeploy(ctx, tx))
	case kTx.DROP_DATABASE:
		return handlePricing(s.priceDrop(ctx, tx))
	case kTx.EXECUTE_ACTION:
		return handlePricing(s.priceAction(ctx, tx))
	case kTx.VALIDATOR_JOIN:
		return handlePricing(s.priceValidatorJoin(ctx, tx))
	case kTx.VALIDATOR_LEAVE:
		return handlePricing(s.priceValidatorLeave(ctx, tx))
	default:
		return nil, fmt.Errorf("invalid payload type")
	}
}

func handlePricing(price *big.Int, err error) (*txpb.EstimatePriceResponse, error) {
	if err != nil {
		return nil, err
	}

	return &txpb.EstimatePriceResponse{
		Price: price.String(),
	}, nil
}

func (s *Service) priceDeploy(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	ds, err := UnmarshalSchema(tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize dataset: %w", err)
	}

	return s.executor.PriceDeploy(&entity.DeployDatabase{
		Tx:     tx,
		Schema: ds,
	})
}

func (s *Service) priceDrop(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	dsIdent, err := UnmarshalDatasetIdentifier(tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize dataset identifier: %w", err)
	}

	return s.executor.PriceDrop(&entity.DropDatabase{
		DBID: utils.GenerateDBID(dsIdent.Name, dsIdent.Owner),
		Tx:   tx,
	})
}

func (s *Service) priceAction(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	executionBody, err := UnmarshalActionExecution(tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize action execution: %w", err)
	}

	return s.executor.PriceExecute(&entity.ExecuteAction{
		Tx:            tx,
		ExecutionBody: executionBody,
	})
}

func (s *Service) priceValidatorJoin(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

func (s *Service) priceValidatorLeave(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

func UnmarshalActionExecution(payload []byte) (*entity.ActionExecution, error) {
	exec := entity.ActionExecution{}

	err := json.Unmarshal(payload, &exec)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal action execution: %w", err)
	}

	return &exec, nil
}

func UnmarshalSchema(payload []byte) (*entity.Schema, error) {
	schema := entity.Schema{}

	err := json.Unmarshal(payload, &schema)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %w", err)
	}

	return &schema, nil
}

func UnmarshalDatasetIdentifier(payload []byte) (*entity.DatasetIdentifier, error) {
	dsIdent := entity.DatasetIdentifier{}

	err := json.Unmarshal(payload, &dsIdent)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal dataset identifier: %w", err)
	}

	return &dsIdent, nil
}
