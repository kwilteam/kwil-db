package chain

type DepositEvent struct {
	Sender    string
	Receiver  string
	Amount    string
	Height    int64
	TxHash    string
	BlockHash string
}
