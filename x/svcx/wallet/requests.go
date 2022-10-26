package wallet

import "kwil/x/svcx/messaging/mx"

type SpendRequest struct {
	request_id string
	WalletId   string
	// ...
}

type WithdrawalRequest struct {
	request_id string
	WalletId   string
	// ...
}

func (s *SpendRequest) AsMessage() *mx.RawMessage {
	// wallet id as key
	// request as value (need to include type as a marker in order to deserialize later during processing)
	panic("implement me")
}

func (s *WithdrawalRequest) AsMessage() *mx.RawMessage {
	// wallet id as key
	// request as value (need to include type as a marker in order to deserialize later during processing)
	panic("implement me")
}

func NewSpendRequest(walletId string /* ... */) SpendRequest {
	panic("implement me")
}

func NewWithdrawalRequest(walletId string /* ... */) WithdrawalRequest {
	panic("implement me")
}
