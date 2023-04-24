package txsvc

import (
	"context"
	"fmt"
	"kwil/internal/entity"
	"kwil/pkg/engine/models"
	kTx "kwil/pkg/tx"
	"kwil/pkg/utils/serialize"
)

func (s *Service) deploy(ctx context.Context, tx *kTx.Transaction) (*kTx.Receipt, error) {
	ds, err := serialize.Deserialize[models.Dataset](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize dataset: %w", err)
	}

	if ds.Owner != tx.Sender {
		return nil, fmt.Errorf("database owner is not the same as the tx sender")
	}

	return s.executor.Deploy(&entity.DeployDatabase{
		Schema: &ds,
		Tx:     tx,
	})
}

func (s *Service) drop(ctx context.Context, tx *kTx.Transaction) (*kTx.Receipt, error) {
	dsIdent, err := serialize.Deserialize[models.DatasetIdentifier](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize dataset identifier: %w", err)
	}

	if dsIdent.Owner != tx.Sender {
		return nil, fmt.Errorf("database owner is not the same as the tx sender")
	}

	return s.executor.Drop(&entity.DropDatabase{
		DBID: models.GenerateSchemaId(dsIdent.Owner, dsIdent.Name),
		Tx:   tx,
	})
}

func (s *Service) executeAction(ctx context.Context, tx *kTx.Transaction) (*kTx.Receipt, error) {
	executionBody, err := serialize.Deserialize[models.ActionExecution](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize action execution: %w", err)
	}

	return s.executor.Execute(&entity.ExecuteAction{
		Tx:            tx,
		ExecutionBody: &executionBody,
	})
}
