package deposits

type Deposit struct {
	Wallet string `json:"wallet"`
	Amount string `json:"amount"`
	TxHash string `json:"tx_hash"`
	Height int64  `json:"height"`
}
