package entity

type Account struct {
	Address string `json:"address"`
	Balance string `json:"balance"`
	Nonce   int64  `json:"nonce"`
}
