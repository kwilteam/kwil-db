package types

// TODO: we need to get rid of this folder.  There were a ton of import issues so it is here for now

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
