package deposits

const CorrelationIdCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type WithdrawalRequest struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

type StartWithdrawal struct {
	CorrelationId string `json:"correlationId"`
	Address       string `json:"address"`
	Amount        string `json:"amount"`
	Fee           string `json:"fee"`
	Expiration    int64  `json:"expiration"`
}
