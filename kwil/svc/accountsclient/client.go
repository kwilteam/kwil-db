package accountsclient

import (
	"context"
	pb "kwil/api/protobuf/account/v0/gen/go"
	"kwil/x/types/accounts"

	"google.golang.org/grpc"
)

type AccountsClient interface {
	GetAccount(ctx context.Context, address string) (accounts.Account, error)
}

type client struct {
	accounts pb.AccountServiceClient
}

func New(cc *grpc.ClientConn) AccountsClient {
	return &client{
		accounts: pb.NewAccountServiceClient(cc),
	}
}
