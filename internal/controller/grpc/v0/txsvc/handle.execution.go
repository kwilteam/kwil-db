package txsvc

import (
	"context"
	"fmt"
	txpb "kwil/api/protobuf/tx/v0/gen/go"
	accountTypes "kwil/x/types/accounts"
	"kwil/x/types/databases/clean"
	"kwil/x/types/execution"
	"kwil/x/types/execution/convert"
	"kwil/x/types/transactions"
	"kwil/x/utils/serialize"
)

func (s *Service) handleExecution(ctx context.Context, tx *transactions.Transaction) (*txpb.BroadcastResponse, error) {
	// get the fee
	price, err := s.pricing.GetPrice(tx)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	ok, err := checkFee(tx.Fee, price)
	if err != nil {
		return nil, fmt.Errorf("failed to check fee: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("fee is not enough")
	}

	// try to spend the fee
	err = s.dao.Spend(ctx, &accountTypes.Spend{
		Address: tx.Sender,
		Amount:  price,
		Nonce:   tx.Nonce,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to spend fee: %w", err)
	}

	// get executionBody
	executionBody, err := serialize.Deserialize[*execution.ExecutionBody[[]byte]](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload of type ExecutionBody: %w", err)
	}

	clean.Clean(&executionBody)

	convExecutionBody, err := convert.Bytes.BodyToKwilAny(executionBody)
	if err != nil {
		return nil, fmt.Errorf("failed to convert execution body to kwil any: %w", err)
	}

	// execute
	err = s.executor.ExecuteQuery(ctx, convExecutionBody, tx.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return &txpb.BroadcastResponse{
		Hash: tx.Hash,
		Fee:  price,
	}, nil
}
