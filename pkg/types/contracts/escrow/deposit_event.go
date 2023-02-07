package escrow

type DepositEvent struct {
	Caller string
	Target string
	Amount string
	Height int64
	TxHash string
}
