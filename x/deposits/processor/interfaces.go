package processor

type Deposit interface {
	Caller() string
	Amount() string
}

type Spend interface {
	Caller() string
	Amount() string
}

type WithdrawalRequest interface {
	Amount() string
	Wallet() string
	Nonce() string
	Expiration() int64
}

type WithdrawalConfirmation interface {
	Nonce() string
}

type FinalizedBlock interface {
	Height() int64
}
