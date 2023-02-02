package common

import (
	"context"
	"errors"
	"kwil/internal/pkg/transport"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type RoundTripper func(context.Context, *grpc.ClientConn) error

func DialGrpc(ctx context.Context, fn RoundTripper) (err error) {
	endpoint := viper.GetString("endpoint")
	if endpoint == "" {
		return errors.New("endpoint not set: use `kwil configure` to set a default endpoint or pass the --endpoint flag")
	}

	conn, err := transport.Dial(ctx, endpoint)
	if err != nil {
		return err
	}
	defer conn.Close()

	return fn(ctx, conn)
}
