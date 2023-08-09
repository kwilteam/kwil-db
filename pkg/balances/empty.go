package balances

// type EmptyAccountStore struct {
// 	logger log.Logger
// }

// func NewEmptyAccountStore(l log.Logger) *EmptyAccountStore {
// 	return &EmptyAccountStore{
// 		logger: l,
// 	}
// }

// func (e *EmptyAccountStore) Close() error {
// 	return nil
// }

// var emptyStoreAccountValue = big.NewInt(5000000000000000000)

// func (e *EmptyAccountStore) GetAccount(address string) (*Account, error) {
// 	return &Account{
// 		Address: address,
// 		Balance: emptyStoreAccountValue,
// 		Nonce:   0,
// 	}, nil
// }

// func (e *EmptyAccountStore) Spend(spend *Spend) error {
// 	e.logger.Info("spend", zap.String("address", spend.AccountAddress), zap.String("amount", spend.Amount.String()))
// 	return nil
// }
// func (a *EmptyAccountStore) GasEnabled() bool {
// 	return false
// }

// func (e *EmptyAccountStore) ApplyChangeset(changeset io.Reader) error {
// 	return nil
// }

// func (e *EmptyAccountStore) CreateSession() (Session, error) {
// 	return nil, nil
// }

// func (e *EmptyAccountStore) Savepoint() (Savepoint, error) {
// 	return nil, nil
// }
