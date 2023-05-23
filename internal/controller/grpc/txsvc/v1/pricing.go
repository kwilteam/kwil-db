package txsvc

import (
	"context"
	"fmt"
	"math/big"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/engine2/utils"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
	"github.com/kwilteam/kwil-db/pkg/utils/serialize"
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
	ds, err := serialize.Deserialize[entity.Schema](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize dataset: %w", err)
	}

	return s.executor.PriceDeploy(&entity.DeployDatabase{
		Tx:     tx,
		Schema: &ds,
	})
}

func (s *Service) priceDrop(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	dsIdent, err := serialize.Deserialize[entity.DatasetIdentifier](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize dataset identifier: %w", err)
	}

	return s.executor.PriceDrop(&entity.DropDatabase{
		DBID: utils.GenerateDBID(dsIdent.Name, dsIdent.Owner),
		Tx:   tx,
	})
}

func (s *Service) priceAction(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	executionBody, err := serialize.Deserialize[entity.ActionExecution](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize action execution: %w", err)
	}

	return s.executor.PriceExecute(&entity.ExecuteAction{
		Tx:            tx,
		ExecutionBody: &executionBody,
	})
}
