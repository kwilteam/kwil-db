package txsvc

import (
	"context"
	"fmt"

	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
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
