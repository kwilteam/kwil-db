package txsvc

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/internal/entity"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	kTx "github.com/kwilteam/kwil-db/pkg/tx"
)

func (s *Service) deploy(ctx context.Context, tx *kTx.Transaction) (*kTx.Receipt, error) {
	ds, err := unmarshalSchema(tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize dataset: %w", err)
	}

	if ds.Owner != tx.Sender {
		return nil, fmt.Errorf("database owner is not the same as the tx sender")
	}

	return s.executor.Deploy(ctx, &entity.DeployDatabase{
		Schema: ds,
		Tx:     tx,
	})
}

func (s *Service) drop(ctx context.Context, tx *kTx.Transaction) (*kTx.Receipt, error) {
	dsIdent, err := unmarshalDatasetIdentifier(tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize dataset identifier: %w", err)
	}

	if dsIdent.Owner != tx.Sender {
		return nil, fmt.Errorf("database owner is not the same as the tx sender")
	}

	return s.executor.Drop(ctx, &entity.DropDatabase{
		DBID: utils.GenerateDBID(dsIdent.Name, dsIdent.Owner),
		Tx:   tx,
	})
}

func (s *Service) executeAction(ctx context.Context, tx *kTx.Transaction) (*kTx.Receipt, error) {
	executionBody, err := unmarshalActionExecution(tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize action execution: %w", err)
	}

	return s.executor.Execute(&entity.ExecuteAction{
		Tx:            tx,
		ExecutionBody: executionBody,
	})
}
