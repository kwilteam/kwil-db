package accounts

func EmptyAccount(address string) *Account {
	return &Account{
		Address: address,
		Nonce:   0,
		Balance: "0",
		Spent:   "0",
	}
}

type Account struct {
	Address string
	Nonce   int64
	Balance string
	Spent   string
}
