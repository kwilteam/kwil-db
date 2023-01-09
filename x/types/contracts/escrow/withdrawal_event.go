package escrow

type WithdrawalConfirmationEvent struct {
	Caller   string // the node that confirmed the withdrawal
	Receiver string // the user that requested the withdrawal
	Amount   string
	Fee      string
	Cid      string
	Height   int64
	TxHash   string
}
