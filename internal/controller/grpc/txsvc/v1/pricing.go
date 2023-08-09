package txsvc

import (
	"context"
	"fmt"
	"math/big"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/serialize"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
)

func (s *Service) EstimatePrice(ctx context.Context, req *txpb.EstimatePriceRequest) (*txpb.EstimatePriceResponse, error) {
	tx, err := convertTx(req.Tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}

	var price *big.Int

	switch tx.PayloadType {
	case kTx.DEPLOY_DATABASE:
		price, err = s.priceDeploy(ctx, tx)
	case kTx.DROP_DATABASE:
		price, err = s.priceDrop(ctx, tx)
	case kTx.EXECUTE_ACTION:
		price, err = s.priceAction(ctx, tx)
	case kTx.VALIDATOR_JOIN:
		price, err = s.priceValidatorJoin(ctx, tx)
	case kTx.VALIDATOR_LEAVE:
		price, err = s.priceValidatorLeave(ctx, tx)
	default:
		price, err = nil, fmt.Errorf("invalid payload type")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	return &txpb.EstimatePriceResponse{
		Price: price.String(),
	}, nil
}

func (s *Service) priceDeploy(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	ds, err := serialize.DeserializeSchema(tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize schema: %w", err)
	}

	return s.engine.PriceDeploy(ctx, ds)
}

func (s *Service) priceDrop(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	dbid, err := serialize.DeserializeDBID(tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize DBID: %w", err)
	}

	return s.engine.PriceDrop(ctx, dbid)
}

func (s *Service) priceAction(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	executionBody, err := serialize.DeserializeActionPaload(tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize action execution: %w", err)
	}

	return s.engine.PriceExecute(ctx, executionBody.DBID, executionBody.Action, executionBody.Params)
}

func (s *Service) priceValidatorJoin(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}

func (s *Service) priceValidatorLeave(ctx context.Context, tx *kTx.Transaction) (*big.Int, error) {
	return big.NewInt(10000000000000), nil
}
