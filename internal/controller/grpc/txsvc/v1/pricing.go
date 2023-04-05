package txsvc

import (
	"context"
	"fmt"
	txpb "kwil/api/protobuf/tx/v1"
	"kwil/internal/entity"
	"kwil/pkg/engine/models"
	kTx "kwil/pkg/tx"
	"kwil/pkg/utils/serialize"
	"math/big"
)

func (s *Service) EstimatePrice(ctx context.Context, req *txpb.EstimatePriceRequest) (*txpb.EstimatePriceResponse, error) {
	tx, err := serialize.Convert[txpb.Tx, kTx.Transaction](req.Tx)
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
	ds, err := serialize.Deserialize[models.Dataset](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize dataset: %w", err)
	}

	return s.executor.PriceDeploy(&entity.DeployDatabase{
		Tx:     tx,
		Schema: &ds,
	})
}

func (s *Service) priceDrop(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	dsIdent, err := serialize.Deserialize[models.DatasetIdentifier](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize dataset identifier: %w", err)
	}

	return s.executor.PriceDrop(&entity.DropDatabase{
		DBID: models.GenerateSchemaId(dsIdent.Owner, dsIdent.Name),
		Tx:   tx,
	})
}

func (s *Service) priceAction(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	executionBody, err := serialize.Deserialize[models.ActionExecution](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize action execution: %w", err)
	}

	return s.executor.PriceExecute(&entity.ExecuteAction{
		Tx:     tx,
		Action: executionBody.Action,
	})
}
