package specifications

import (
	"context"
	"testing"
	"time"

	"github.com/kwilteam/kwil-db/pkg/transactions"

	"github.com/stretchr/testify/assert"
)

// DatabaseDeployDsl is dsl for database deployment specification
type NetworkOpsDsl interface {
	ApproveNode(ctx context.Context, joinerPubKey string, approverPrivKey string) error
	ValidatorNodeJoin(ctx context.Context, pubKey string, power int64) error
	ValidatorNodeLeave(ctx context.Context, pubKey string) error
	// ValidatorJoinStatus(ctx context.Context, pubKey []byte) error
	ValidatorSetCount(ctx context.Context) (int, error)
	DeployDatabase(ctx context.Context, db *transactions.Schema) (txHash []byte, err error)
	DropDatabase(ctx context.Context, dbName string) error
}

func NetworkNodeDeploySpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl) {
	netops.DeployDatabase(ctx, SchemaLoader.Load(t, schemaTestDB))
	time.Sleep(15 * time.Second)
}

func NetworkNodeValidatorSetSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl, count int) {
	t.Log("Executing network node validator set specification")
	cnt, err := netops.ValidatorSetCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, count, cnt)
}

func NetworkNodeJoinSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl, joiner string) {
	t.Log("Executing network node join specification")
	err := netops.ValidatorNodeJoin(ctx, joiner, 1)
	assert.NoError(t, err)
}

func NetworkNodeLeaveSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl, joiner string) {
	t.Log("Executing network node leave specification")
	err := netops.ValidatorNodeLeave(ctx, joiner)
	assert.NoError(t, err)
}

func NetworkNodeApproveSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl, joiner string, approver string) {
	t.Log("Executing network node approve specification")
	err := netops.ApproveNode(ctx, joiner, approver)
	assert.NoError(t, err)
}
