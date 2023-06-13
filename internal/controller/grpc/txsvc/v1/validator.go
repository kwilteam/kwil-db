package txsvc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	localClient "github.com/cometbft/cometbft/rpc/client/local"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/internal/entity"
	"go.uber.org/zap"
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

	address := publicKey.Address().String()
	if s.Validators.IsValidator(address) {
		log = fmt.Sprintf("Validator %s is already approved\n", address)
		fmt.Println(log)
		return &txpb.ValidatorApprovalResponse{Status: txpb.RequestStatus_OK, Log: log}, nil
	}

	err = s.Validators.AddValidator(address)
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
	tx, err := convertTx(req.Tx)
	if err != nil {
		fmt.Println("failed to convert Tx", err)
		return nil, fmt.Errorf("failed to convert Tx: %w", err)
	}

	validator, err := UnmarshalValidator(tx.Payload)
	if err != nil {
		fmt.Println("failed to unmarshal Validator", err)
		return &txpb.ValidatorJoinResponse{
			Receipt: &txpb.TxReceipt{
				TxHash: tx.Hash,
			},
		}, fmt.Errorf("failed to unmarshal Validator: %w", err)
	}
	fmt.Println("ValidatorJoin Validator:", validator.PubKey, validator)

	pubKey, err := UnmarshalValidatorPublicKey(string(validator.PubKey))
	if err != nil {
		fmt.Println("failed to unmarshal Validator public key", err)
		return &txpb.ValidatorJoinResponse{
			Receipt: &txpb.TxReceipt{
				TxHash: tx.Hash,
			},
		}, fmt.Errorf("failed to unmarshal Validator public key: %w", err)
	}

	if s.Validators.IsValidator(pubKey.Address().String()) {
		fmt.Println("Validator is already approved - can join")
	} else {
		fmt.Println("Validator is not approved - can't join")
		return &txpb.ValidatorJoinResponse{
			Receipt: &txpb.TxReceipt{
				TxHash: tx.Hash,
			},
		}, fmt.Errorf("validator is not approved - can't join")
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
