package specifications

import (
	"context"
	"testing"
	"time"

	schema "github.com/kwilteam/kwil-db/internal/entity"
	"github.com/stretchr/testify/assert"
)

// DatabaseDeployDsl is dsl for database deployment specification
type NetworkOpsDsl interface {
	ApproveNode(ctx context.Context, pubKey []byte) error
	ValidatorNodeJoin(ctx context.Context, pubKey []byte, power int64) error
	ValidatorNodeLeave(ctx context.Context, pubKey []byte) error
	ValidatorJoinStatus(ctx context.Context, pubKey []byte) error
	ValidatorSetCount(ctx context.Context) (int, error)
	DeployDatabase(ctx context.Context, db *schema.Schema) error
	DropDatabase(ctx context.Context, dbName string) error
}

func NetworkNodeApproveSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl, pubkey []byte) {
	t.Log("Executing network node approve specification")

	err := netops.ApproveNode(ctx, pubkey)
	assert.NoError(t, err)
}

func NetworkNodeJoinSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl, pubkey []byte) {
	t.Log("Executing network node join specification")

	pre_cnt, err := netops.ValidatorSetCount(ctx)
	assert.NoError(t, err)

	err = netops.ValidatorNodeJoin(ctx, pubkey, 1)
	assert.NoError(t, err)
	_ = netops.ValidatorJoinStatus(ctx, pubkey)

	netops.DeployDatabase(ctx, SchemaLoader.Load(t, schema_testdb))
	time.Sleep(15 * time.Second)

	post_cnt, err := netops.ValidatorSetCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, pre_cnt+1, post_cnt)
}

func NetworkNodeJoinFailureSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl, pubkey []byte) {
	t.Log("Executing network node join failure specification")
	pre_cnt, err := netops.ValidatorSetCount(ctx)
	assert.NoError(t, err)

	err = netops.ValidatorNodeJoin(ctx, pubkey, 1)
	assert.NoError(t, err)

	time.Sleep(15 * time.Second)
	_ = netops.ValidatorJoinStatus(ctx, pubkey)
	netops.DeployDatabase(ctx, SchemaLoader.Load(t, schema_testdb))

	post_cnt, err := netops.ValidatorSetCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, pre_cnt, post_cnt)
}

func NetworkNodeLeaveSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl, pubkey []byte) {
	t.Log("Executing network node leave success specification")

	pre_cnt, err := netops.ValidatorSetCount(ctx)
	assert.NoError(t, err)

	err = netops.ValidatorNodeLeave(ctx, pubkey)
	assert.NoError(t, err)

	netops.DropDatabase(ctx, SchemaLoader.Load(t, schema_testdb).Name)
	time.Sleep(15 * time.Second)

	post_cnt, err := netops.ValidatorSetCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, pre_cnt, post_cnt+1)
}

func NetworkNodeValidatorSetSpecification(ctx context.Context, t *testing.T, netops NetworkOpsDsl, count int) {
	t.Log("Executing network node validator set specification")
	cnt, err := netops.ValidatorSetCount(ctx)
	assert.NoError(t, err)
	assert.Equal(t, count, cnt)
}
