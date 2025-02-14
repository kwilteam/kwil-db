package signersvc

import (
	"context"

	"github.com/kwilteam/kwil-db/core/client"
	clientTypes "github.com/kwilteam/kwil-db/core/client/types"
	"github.com/kwilteam/kwil-db/core/types"
)

type RewardInstanceInfo struct {
	Chain       string
	Escrow      string
	EpochPeriod string
	Erc20       string
	Decimals    int64
	Balance     string
	Synced      bool
	SyncedAt    int64
	Enabled     bool
}

type Epoch struct {
	ID             types.UUID
	StartHeight    int64
	StartTimestamp int64
	EndHeight      int64
	RewardRoot     []byte
	RewardAmount   string
	EndBlockHash   []byte
	Confirmed      bool
	Voters         []string
	VoteAmounts    []string
	VoteNonces     []int64
	VoteSignatures [][]byte
}

type FinalizedReward struct {
	ID         types.UUID
	Voters     []string
	Signatures [][]byte
	EpochID    types.UUID
	CreatedAt  int64
	//
	StartHeight  int64
	EndHeight    int64
	TotalRewards types.Decimal
	RewardRoot   []byte
	SafeNonce    int64
	SignHash     []byte
	ContractID   types.UUID
	BlockHash    []byte
}

type EpochReward struct {
	Recipient string
	Amount    string
}

// erc20ExtAPI defines the ERC20 reward extension API used by SignerSvc.
type erc20ExtAPI interface {
	GetTarget() string
	SetTarget(ns string)
	InstanceInfo(tx context.Context) (*RewardInstanceInfo, error)
	GetActiveEpochs(ctx context.Context) ([]*Epoch, error)
	ListUnconfirmedEpochs(ctx context.Context, afterHeight int64, limit int) ([]*Epoch, error)
	GetEpochRewards(ctx context.Context, epochID types.UUID) ([]*EpochReward, error)
	VoteEpoch(ctx context.Context, epochID types.UUID, amount *types.Decimal, safeNonce int64, signature []byte) (string, error)
}

type erc20rwExtApi struct {
	clt        *client.Client
	namespace  string
	instanceID string
}

var _ erc20ExtAPI = (*erc20rwExtApi)(nil)

func newERC20RWExtAPI(clt *client.Client, ns string) *erc20rwExtApi {
	return &erc20rwExtApi{
		clt:       clt,
		namespace: ns,
	}
}

func (k *erc20rwExtApi) GetTarget() string {
	return k.namespace
}

func (k *erc20rwExtApi) SetTarget(ns string) {
	k.namespace = ns
}

func (k *erc20rwExtApi) InstanceInfo(ctx context.Context) (*RewardInstanceInfo, error) {
	procedure := "info"
	input := []any{}

	res, err := k.clt.Call(ctx, k.namespace, procedure, input)
	if err != nil {
		return nil, err
	}

	if len(res.QueryResult.Values) == 0 {
		return nil, nil
	}

	er := &RewardInstanceInfo{}
	err = types.ScanTo(res.QueryResult.Values[0],
		&er.Chain, &er.Escrow, &er.EpochPeriod, &er.Erc20, &er.Decimals, &er.Balance, &er.Synced, &er.SyncedAt, &er.Enabled)
	if err != nil {
		return nil, err
	}

	return er, nil
}

func (k *erc20rwExtApi) GetActiveEpochs(ctx context.Context) ([]*Epoch, error) {
	procedure := "get_active_epochs"
	input := []any{}

	res, err := k.clt.Call(ctx, k.namespace, procedure, input)
	if err != nil {
		return nil, err
	}

	if len(res.QueryResult.Values) == 0 {
		return nil, nil
	}

	ers := make([]*Epoch, len(res.QueryResult.Values))
	for i, v := range res.QueryResult.Values {
		er := &Epoch{}
		err = types.ScanTo(v, &er.ID, &er.StartHeight, &er.StartTimestamp, &er.EndHeight,
			&er.RewardRoot, &er.RewardAmount, &er.EndBlockHash, &er.Confirmed, &er.Voters, &er.VoteAmounts, &er.VoteNonces, &er.VoteSignatures)
		if err != nil {
			return nil, err
		}
		ers[i] = er
	}

	return ers, nil
}

func (k *erc20rwExtApi) ListUnconfirmedEpochs(ctx context.Context, afterHeight int64, limit int) ([]*Epoch, error) {
	procedure := "list_epochs"

	input := []any{afterHeight, limit, false}

	res, err := k.clt.Call(ctx, k.namespace, procedure, input)
	if err != nil {
		return nil, err
	}

	if len(res.QueryResult.Values) == 0 {
		return nil, nil
	}

	ers := make([]*Epoch, len(res.QueryResult.Values))
	for i, v := range res.QueryResult.Values {
		er := &Epoch{}
		err = types.ScanTo(v, &er.ID, &er.StartHeight, &er.StartTimestamp, &er.EndHeight,
			&er.RewardRoot, &er.RewardAmount, &er.EndBlockHash, &er.Confirmed, &er.Voters, &er.VoteAmounts, &er.VoteNonces, &er.VoteSignatures)
		if err != nil {
			return nil, err
		}
		ers[i] = er
	}

	return ers, nil
}

func (k *erc20rwExtApi) GetEpochRewards(ctx context.Context, epochID types.UUID) ([]*EpochReward, error) {
	procedure := "get_epoch_rewards"
	input := []any{epochID}

	res, err := k.clt.Call(ctx, k.namespace, procedure, input)
	if err != nil {
		return nil, err
	}

	if len(res.QueryResult.Values) == 0 {
		return nil, nil
	}

	ers := make([]*EpochReward, len(res.QueryResult.Values))
	for i, v := range res.QueryResult.Values {
		er := &EpochReward{}
		err = types.ScanTo(v, &er.Recipient, &er.Amount)
		if err != nil {
			return nil, err
		}
		ers[i] = er
	}

	return ers, nil
}

func (k *erc20rwExtApi) VoteEpoch(ctx context.Context, epochID types.UUID, amount *types.Decimal, safeNonce int64, signature []byte) (string, error) {
	procedure := "vote_epoch"
	input := [][]any{{epochID, amount, safeNonce, signature}}

	res, err := k.clt.Execute(ctx, k.namespace, procedure, input, clientTypes.WithSyncBroadcast(true))
	if err != nil {
		return "", err
	}

	return res.String(), nil
}
