package sql

// var (
// 	ErrNoTransaction = errors.New("no transaction")
// 	ErrNoRows        = errors.New("no rows in result set")
// )

// // TxCloser terminates a transaction by committing or rolling it back. A method
// // that returns this alone would keep the tx under the hood of the parent type,
// // directing queries internally through the scope of a transaction/session
// // started with BeginTx.
// type TxCloser interface {
// 	Rollback(ctx context.Context) error
// 	Commit(ctx context.Context) error
// }

// // TxPrecommitter is the special kind of transaction that can prepare a
// // transaction for commit.
// // It is only available on the outermost transaction.
// type TxPrecommitter interface {
// 	Precommit(ctx context.Context) ([]byte, error)
// }

// type TxBeginner interface {
// 	Begin(ctx context.Context) (TxCloser, error)
// }

// // OuterTxMaker is the special kind of transaction beginner that can make nested
// // transactions, and that explicitly scopes Query/Execute to the tx.
// type OuterTxMaker interface {
// 	BeginTx(ctx context.Context) (OuterTx, error)
// }

// // ReadTxMaker can make read-only transactions.
// // Many read-only transactions can be made at once.
// type ReadTxMaker interface {
// 	BeginReadTx(ctx context.Context) (common.Tx, error)
// }

// // TxMaker is the special kind of transaction beginner that can make nested
// // transactions, and that explicitly scopes Query/Execute to the tx.
// type TxMaker interface {
// 	BeginTx(ctx context.Context) (common.Tx, error)
// }

// // OuterTx is a database transaction. It is the outermost transaction type.
// // "nested transactions" are called savepoints, and can be created with
// // BeginSavepoint. Savepoints can be nested, and are rolled back to the
// // innermost savepoint on Rollback.
// //
// // Anything using implicit tx/session management should use TxCloser.
// type OuterTx interface {
// 	common.Tx
// 	TxPrecommitter
// }
