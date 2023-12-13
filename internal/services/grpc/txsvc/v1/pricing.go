package txsvc

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/core/rpc/conversion"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"go.uber.org/zap"
)

func (s *Service) EstimatePrice(ctx context.Context, req *txpb.EstimatePriceRequest) (*txpb.EstimatePriceResponse, error) {
	s.log.Debug("Estimating price", zap.String("payload_type", req.Tx.Body.PayloadType))

	tx, err := conversion.ConvertFromPBTx(req.Tx)
	if err != nil {
		return nil, fmt.Errorf("failed to convert transaction: %w", err)
	}
	price, err := s.nodeApp.Price(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate price: %w", err)
	}

	return &txpb.EstimatePriceResponse{
		Price: price.String(),
	}, nil
}
