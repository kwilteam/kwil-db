package specifications

import (
	"context"
)

// TxQueryDsl is dsl for tx query specification
type TxQueryDsl interface {
	TxSuccess(ctx context.Context, txHash []byte) error
}
