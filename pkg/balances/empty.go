package balances

import (
	"math/big"

	"github.com/kwilteam/kwil-db/pkg/log"
	"go.uber.org/zap"
)

type EmptyAccountStore struct {
	logger log.Logger
}

func NewEmptyAccountStore(l log.Logger) *EmptyAccountStore {
	return &EmptyAccountStore{
		logger: l,
	}
}

func (e *EmptyAccountStore) BatchCredit(credits []*Credit, cfg *ChainConfig) error {
	for _, credit := range credits {
		e.logger.Info("credit", zap.String("address", credit.AccountAddress), zap.String("amount", credit.Amount.String()))
	}
	return nil
}

func (e *EmptyAccountStore) Close() error {
	return nil
}

var emptyStoreAccountValue = big.NewInt(5000000000000000000)

func (e *EmptyAccountStore) GetAccount(address string) (*Account, error) {
	return &Account{
		Address: address,
		Balance: emptyStoreAccountValue,
		Nonce:   0,
	}, nil
}

func (e *EmptyAccountStore) Spend(spend *Spend) error {
	e.logger.Info("spend", zap.String("address", spend.AccountAddress), zap.String("amount", spend.Amount.String()))
	return nil
}

func (e *EmptyAccountStore) ChainExists(chainCode int32) (bool, error) {
	return true, nil
}

func (e *EmptyAccountStore) CreateChain(chainCode int32, height int64) error {
	return nil
}

func (e *EmptyAccountStore) BatchSpend(spendList []*Spend, chain *ChainConfig) error {
	for _, spend := range spendList {
		e.logger.Info("spend", zap.String("address", spend.AccountAddress), zap.String("amount", spend.Amount.String()))
	}
	return nil
}

func (e *EmptyAccountStore) Credit(credit *Credit) error {
	e.logger.Info("credit", zap.String("address", credit.AccountAddress), zap.String("amount", credit.Amount.String()))
	return nil
}

func (e *EmptyAccountStore) GetHeight(chainCode int32) (int64, error) {
	return 100000, nil
}

func (e *EmptyAccountStore) SetHeight(chainCode int32, height int64) error {
	return nil
}
