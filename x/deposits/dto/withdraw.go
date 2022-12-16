package dto

const CidCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type StartWithdrawal struct {
	Wallet string `json:"wallet"`
	Amount string `json:"amount"`
}
