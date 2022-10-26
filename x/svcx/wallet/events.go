package wallet

type WithdrawnEvent struct {
	request_id string
	WalletId   string
	// ...
}

type SpentEvent struct {
	request_id string
	WalletId   string
	// ...
}

type DepositedEvent struct {
	// ...
}
