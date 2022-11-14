package sql

import "errors"

var ErrConstraintPrimaryKey = errors.New("constraint primary key")
var ErrFailedToParseBigInt = errors.New("failed to parse big int")
var ErrInsufficientFunds = errors.New("insufficient funds")
var ErrTxRollback = errors.New("transaction rollback")
