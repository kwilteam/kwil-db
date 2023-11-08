package types

import "math/big"

type ApproveResponse struct {
	TxHash string `json:"txHash"`
}

// type BalanceParams struct {
// 	Address string
// }

// type BalanceResponse struct {
// 	Balance *big.Int
// }

// type TransferResponse struct {
// 	TxHash string `json:"txHash"`
// }

// type DepositBalanceParams struct {
// 	Validator string
// 	Address   string
// }

// type DepositBalanceResponse struct {
// 	Balance *big.Int
// }

// type DepositEvent struct {
// 	Caller string
// 	Target string
// 	Amount string
// 	Height int64
// 	TxHash string
// }

type DepositParams struct {
	Validator string
	Amount    *big.Int
}

type DepositResponse struct {
	TxHash string
}

// type ReturnFundsParams struct {
// 	Recipient     string
// 	CorrelationId string
// 	Amount        *big.Int
// 	Fee           *big.Int
// }

// type ReturnFundsResponse struct {
// 	TxHash string
// }

// type WithdrawalConfirmationEvent struct {
// 	Caller   string // the node that confirmed the withdrawal
// 	Receiver string // the user that requested the withdrawal
// 	Amount   string
// 	Fee      string
// 	Cid      string
// 	Height   int64
// 	TxHash   string
// }
