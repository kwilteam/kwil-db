package txsvc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"

	"github.com/cometbft/cometbft/crypto/tmhash"
	localClient "github.com/cometbft/cometbft/rpc/client/local"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/internal/node"
)

func (s *Service) ApproveValidator(ctx context.Context, req *txpb.ValidatorApprovalRequest) (*txpb.ValidatorApprovalResponse, error) {
	var log string
	key := fmt.Sprintf(`{"type":"tendermint/PubKeyEd25519","value":"%s"}`, req.PubKey)
	fmt.Println("Approve Validator Key:", key)

	var publicKey crypto.PubKey
	err := cmtjson.Unmarshal([]byte(key), &publicKey)
	if err != nil {
		log = fmt.Sprintln("failed to unmarshal validator public key", err)
		return &txpb.ValidatorApprovalResponse{Status: txpb.RequestStatus_ERROR, Log: log}, fmt.Errorf("failed to unmarshal validator public key: %w", err)
	}

	fmt.Println("Approve Validator PubKey:", publicKey, publicKey.Address(), publicKey.Address().String())

	Validators := s.NodeReactor.GetPool().ApprovedVals
	address := publicKey.Address().String()
	if Validators.IsValidator(address) {
		log = fmt.Sprintf("Validator %s is already approved\n", address)
		fmt.Println(log)
		return &txpb.ValidatorApprovalResponse{Status: txpb.RequestStatus_OK, Log: log}, nil
	}

	err = Validators.AddValidator(address)
	if err != nil {
		log = fmt.Sprintf("Validator %s couldn't be approved\n", address)
		fmt.Println(log)
		return &txpb.ValidatorApprovalResponse{Status: txpb.RequestStatus_ERROR, Log: log}, err
	}

	log = fmt.Sprintf("Validator %s has been approved\n", address)
	fmt.Println(log)
	return &txpb.ValidatorApprovalResponse{Status: txpb.RequestStatus_OK, Log: log}, nil
}

func (s *Service) ValidatorJoin(ctx context.Context, req *txpb.ValidatorJoinRequest) (*txpb.ValidatorJoinResponse, error) {
	if !s.NodeReactor.GetPool().IsNodeValidator() {
		fmt.Println("Node is not a validator - Not broadcasting join request")
		return nil, fmt.Errorf("node is not a validator - not broadcasting join request")
	}

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
	txHash := tmhash.Sum(bts)

	validator, err := UnmarshalValidator(tx.Payload)
	if err != nil {
		fmt.Println("failed to unmarshal Validator", err)
		return &txpb.ValidatorJoinResponse{
			Receipt: &txpb.TxReceipt{
				TxHash: txHash,
			},
		}, fmt.Errorf("failed to unmarshal Validator: %w", err)
	}
	fmt.Println("ValidatorJoin Validator:", validator.PubKey, validator)

	pubKey, err := UnmarshalValidatorPublicKey(string(validator.PubKey))
	if err != nil {
		fmt.Println("failed to unmarshal Validator public key", err)
		return &txpb.ValidatorJoinResponse{
			Receipt: &txpb.TxReceipt{
				TxHash: txHash,
			},
		}, fmt.Errorf("failed to unmarshal Validator public key: %w", err)
	}

	// Check if the validator is already approved by the network, if so, broadcast the tx to the blockchain to add the node as a validator to the network
	if s.NodeReactor.GetPool().ApprovedNetworkVals.IsValidator(pubKey.Address().String()) {
		fmt.Println("ValidatorJoin Validator is already approved by the network", pubKey.Address().String())
		bcClient := localClient.New(s.BcNode)
		_, err := bcClient.BroadcastTxAsync(context.Background(), bts)
		if err != nil {
			fmt.Println("failed to broadcast tx", err)
		}
		return &txpb.ValidatorJoinResponse{
			Receipt: &txpb.TxReceipt{
				TxHash: txHash,
			},
		}, err
	}

	// If the validator is not approved by the network, add the request to the network pool for approval
	s.NodeReactor.GetPool().AddRequest(node.JoinRequest{
		PubKey: pubKey,
		Tx:     bts,
	})

	return &txpb.ValidatorJoinResponse{
		Receipt: &txpb.TxReceipt{
			TxHash: txHash,
		},
	}, nil
}

func (s *Service) ValidatorLeave(ctx context.Context, req *txpb.ValidatorLeaveRequest) (*txpb.ValidatorLeaveResponse, error) {
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
	res, err := bcClient.BroadcastTxAsync(context.Background(), bts)
	if err != nil {
		fmt.Println("Error broadcasting tx for removing the node as Validator", "err", err)
		return &txpb.ValidatorLeaveResponse{
			Receipt: &txpb.TxReceipt{
				TxHash: res.Hash,
			},
		}, fmt.Errorf("error broadcasting tx for removing the node as Validator: %w", err)
	}

	return &txpb.ValidatorLeaveResponse{
		Receipt: &txpb.TxReceipt{
			TxHash: res.Hash,
		},
	}, nil
}

func (s *Service) ValidatorJoinStatus(ctx context.Context, req *txpb.ValidatorJoinStatusRequest) (*txpb.ValidatorJoinStatusResponse, error) {
	key := fmt.Sprintf(`{"type":"tendermint/PubKeyEd25519","value":"%s"}`, req.Pubkey)
	fmt.Println("Approve Validator Key:", key)

	var publicKey crypto.PubKey
	err := cmtjson.Unmarshal([]byte(key), &publicKey)
	if err != nil {
		return &txpb.ValidatorJoinStatusResponse{}, fmt.Errorf("failed to unmarshal validator public key: %w", err)
	}

	address := publicKey.Address().String()
	status := s.NodeReactor.GetPool().GetStatus(address)
	var joinStatus txpb.JoinStatus

	switch status.Status {
	case int64(node.Initiated):
		joinStatus = txpb.JoinStatus_INITIATED
	case int64(node.Pending):
		joinStatus = txpb.JoinStatus_PENDING
	case int64(node.Approved):
		joinStatus = txpb.JoinStatus_APPROVED
	case int64(node.Rejected):
		joinStatus = txpb.JoinStatus_REJECTED
	default:
		joinStatus = txpb.JoinStatus_UNKNOWN
	}

	return &txpb.ValidatorJoinStatusResponse{
		Approved:           status.Approved_votes,
		Rejected:           status.Rejected_votes,
		Pending:            status.Required_votes,
		ApprovedValidators: status.Approved_validators,
		RejectedValidators: status.Rejected_validators,
		Status:             joinStatus,
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

func UnmarshalValidatorPublicKey(pubKey string) (crypto.PubKey, error) {
	var publicKey crypto.PubKey
	key := fmt.Sprintf(`{"type":"tendermint/PubKeyEd25519","value":"%s"}`, pubKey)
	fmt.Println("Key:", key)

	err := cmtjson.Unmarshal([]byte(key), &publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal validator public key: %w", err)
	}
	fmt.Println("publicKey: ", publicKey)
	return publicKey, nil
}
