package util

import (
	"context"
	"errors"
	"fmt"

	v0 "github.com/kwilteam/kwil-db/internal/api/proto/v0"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type RoundTripper func(context.Context, v0.KwilServiceClient) error

func ConnectKwil(ctx context.Context, v *viper.Viper, fn RoundTripper) (err error) {
	endpoint := v.GetString("endpoint")
	if endpoint == "" {
		return errors.New("endpoint not set: use `kwil configure` to set a default endpoint or pass the --endpoint flag")
	}

	apiKey := v.GetString("api-key")
	timeout := v.GetDuration("timeout")

	clientContext := ctx
	if apiKey != "" {
		clientContext = metadata.AppendToOutgoingContext(clientContext, "authorization", apiKey)
	}

	opts := []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure(), grpc.WithTimeout(timeout)}
	cc, err := grpc.DialContext(ctx, endpoint, opts...)
	if err != nil {
		if err == context.DeadlineExceeded {
			return fmt.Errorf("timeout dialing server: %s", endpoint)
		}
		return err
	}
	defer cc.Close()

	return fn(clientContext, v0.NewKwilServiceClient(cc))
}
