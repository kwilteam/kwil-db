package app

import (
	"kwil/x/proto/txpb"
	"kwil/x/transactions/service"
)

type Service struct {
	txpb.UnimplementedTxServiceServer

	service service.TransactionService
}
