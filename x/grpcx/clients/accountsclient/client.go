package accountsclient

import (
	"context"
	"kwil/x/proto/accountspb"
	"kwil/x/types/accounts"

	"google.golang.org/grpc"
)

type AccountsClient interface {
	GetAccount(ctx context.Context, address string) (accounts.Account, error)
}

type client struct {
	accounts accountspb.AccountServiceClient
}

func New(cc *grpc.ClientConn) AccountsClient {
	return &client{
		accounts: accountspb.NewAccountServiceClient(cc),
	}
}
