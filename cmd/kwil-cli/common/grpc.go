package common

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type RoundTripper func(context.Context, *grpc.ClientConn) error

func DialGrpc(ctx context.Context, fn RoundTripper) (err error) {
	endpoint := viper.GetString("endpoint")
	if endpoint == "" {
		return errors.New("endpoint not set: use `kwil configure` to set a default endpoint or pass the --endpoint flag")
	}

	apiKey := viper.GetString("api-key")
	timeout := viper.GetDuration("timeout")

	clientContext := ctx
	if apiKey != "" {
		clientContext = metadata.AppendToOutgoingContext(clientContext, "authorization", apiKey)
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		clientContext, cancel = context.WithTimeout(clientContext, timeout)
		defer cancel()
	}

	opts := []grpc.DialOption{grpc.WithBlock(), grpc.WithTransportCredentials(insecure.NewCredentials())}
	cc, err := grpc.DialContext(clientContext, endpoint, opts...)
	if err != nil {
		if err == context.DeadlineExceeded {
			return fmt.Errorf("timeout dialing server: %s", endpoint)
		}
		return err
	}
	defer cc.Close()

	return fn(clientContext, cc)
}
