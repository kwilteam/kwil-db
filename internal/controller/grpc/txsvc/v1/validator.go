package txsvc

import (
	"context"
	"encoding/json"
	"fmt"

	localClient "github.com/cometbft/cometbft/rpc/client/local"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/entity"
	"go.uber.org/zap"
)

func (s *Service) ApproveValidator(ctx context.Context, req *txpb.ValidatorApprovalRequest) (*txpb.ValidatorApprovalResponse, error) {
	validator := req.PubKey
	var log string
	if s.Validators.IsValidator(validator) {
		log = fmt.Sprintf("Validator %s is already approved\n", validator)
		fmt.Println(log)
		return &txpb.ValidatorApprovalResponse{Status: txpb.RequestStatus_OK, Log: log}, nil
	}

	err := s.Validators.AddValidator(validator)
	if err != nil {
		log = fmt.Sprintf("Validator %s couldn't be approved\n", validator)
		fmt.Println(log)
		return &txpb.ValidatorApprovalResponse{Status: txpb.RequestStatus_ERROR, Log: log}, err
	}

	log = fmt.Sprintf("Validator %s has been approved\n", validator)
	fmt.Println(log)
	return &txpb.ValidatorApprovalResponse{Status: txpb.RequestStatus_OK, Log: log}, nil
}

func (s *Service) ValidatorJoin(ctx context.Context, req *txpb.ValidatorJoinRequest) (*txpb.ValidatorJoinResponse, error) {
	tx, err := convertTx(req.Tx)
	if err != nil {
		fmt.Println("failed to convert Tx", err)
		return nil, fmt.Errorf("failed to convert Tx: %w", err)
	}

	bts, err := json.Marshal(tx)
	if err != nil {
		fmt.Println("failed to marshal Tx", err)
		return nil, fmt.Errorf("failed to marshal Tx: %w", err)
	}

	bcClient := localClient.New(s.BcNode)
	res, err := bcClient.BroadcastTxAsync(ctx, bts)
	if err != nil {
		fmt.Println("failed to broadcast Tx", err)
		return nil, fmt.Errorf("failed to broadcast Tx: %w", err)
	}

	s.log.Info("broadcasted transaction ", zap.String("payload_type", tx.PayloadType.String()))
	fmt.Println("Tx Hash:", res.Hash)
	return &txpb.ValidatorJoinResponse{
		Receipt: &txpb.TxReceipt{
			TxHash: tx.Hash,
		},
	}, nil
}

func UnmarshalValidator(payload []byte) (*entity.Validator, error) {
	validator := entity.Validator{}

	err := json.Unmarshal(payload, &validator)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal validator: %w", err)
	}

	return &validator, nil
}
