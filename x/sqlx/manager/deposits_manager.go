package manager

import (
	"context"
	"kwil/x/cfgx"
	deposit_store "kwil/x/deposits_old/store"
	"kwil/x/sqlx/sqlclient"
)

type DepositsManager interface {
	Spend(ctx context.Context, wallet string, amount string) error
}

type depositManagerConfig struct {
	spend        bool
	read_balance bool
}

type depositsManager struct {
	client       *sqlclient.DB
	depositStore deposit_store.DepositStore
	config       depositManagerConfig
}

func NewDepositsManager(client *sqlclient.DB, cfg cfgx.Config) (*depositsManager, error) {
	spend, err := cfg.GetBool("deposit-settings.spend-balance", true)
	if err != nil {
		return nil, err
	}
	read_balance, err := cfg.GetBool("deposit-settings.read-balance", true)
	if err != nil {
		return nil, err
	}

	return &depositsManager{
		client:       client,
		depositStore: deposit_store.New(client),
		config: depositManagerConfig{
			spend:        spend,
			read_balance: read_balance,
		},
	}, nil
}

func (m *depositsManager) Spend(ctx context.Context, wallet string, amount string) error {
	if !m.config.spend {
		return nil
	}
	return m.depositStore.Spend(ctx, wallet, amount)
}

func (m *depositsManager) GetBalanceAndSpent(ctx context.Context, wallet string) (string, string, error) {
	if !m.config.read_balance {
		return "", "", nil
	}
	return m.depositStore.GetBalanceAndSpent(ctx, wallet)
}
