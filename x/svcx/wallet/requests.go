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

	// create and serialize SpendEvent from SpendRequest

	panic("implement me")
}

func (s *WithdrawalRequest) AsMessage() *mx.RawMessage {
	// wallet id as key
	// request as value (need to include type as a marker in order to deserialize later during processing)

	// create and serialize WithdrawalEvent from WithdrawalRequest

	panic("implement me")
}

func NewSpendRequest(walletId string /* other params here */) SpendRequest {
	// generate request id (uuid)
	panic("implement me")
}

func NewWithdrawalRequest(walletId string /* other params here */) WithdrawalRequest {
	// generate request id (uuid)
	panic("implement me")
}
