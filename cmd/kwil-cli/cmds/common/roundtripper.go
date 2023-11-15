package common

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/kwilteam/kwil-db/cmd/common/display"
	"github.com/kwilteam/kwil-db/cmd/kwil-cli/config"
	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/gatewayclient"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const (
	// WithoutPrivateKey is a flag that can be passed to DialClient to indicate that the client does not need to use the private key.
	WithoutPrivateKey uint8 = 1 << iota

	// UsingGateway is a flag that can be passed to DialClient to indicate that the client is talking to a gateway.
	// Since very few commands use the gateway, we bind this to specific commands instead of making it a global flag.
	UsingGateway
)

type RoundTripper func(ctx context.Context, client Client, conf *config.KwilCliConfig) error

// DialClient dials a kwil node and calls the passed function with the client.
// It includes the command that is being run, so that it can read global flags.
func DialClient(ctx context.Context, cmd *cobra.Command, flags uint8, fn RoundTripper) error {
	conf, err := config.LoadCliConfig()
	if err != nil {
		return err
	}

	if conf.GrpcURL == "" {
		return fmt.Errorf("kwil provider url is required")
	}

	clientConfig := client.ClientOptions{}
	if conf.PrivateKey != nil {
		clientConfig.Signer = &auth.EthPersonalSigner{Key: *conf.PrivateKey}
		clientConfig.ChainID = conf.ChainID
	} else if flags&WithoutPrivateKey == 0 {
		return fmt.Errorf("private key not provided")
	}

	// if not using the gateway, then we can simply create a regular client and return
	if flags&UsingGateway == 0 {
		client, err := client.Dial(ctx, conf.GrpcURL, &clientConfig)
		if err != nil {
			return err
		}

		return fn(ctx, client, conf)
	}

	// if we reach here, we are talking to a gateway

	client, err := gatewayclient.NewGatewayClient(ctx, conf.GrpcURL, &gatewayclient.GatewayOptions{
		ClientOptions: clientConfig,
		AuthSignFunc: func(message string, signer auth.Signer) (*auth.Signature, error) {
			assumeYes, err := GetAssumeYesFlag(cmd)
			if err != nil {
				return nil, err
			}

			if !assumeYes {
				err := promptMessage(message)
				if err != nil {
					return nil, err
				}
			}

			sig, err := signer.Sign([]byte(message))
			if err != nil {
				return nil, err
			}

			return sig, nil
		},
	})
	if err != nil {
		return err
	}

	if clientConfig.Signer == nil {
		return fn(ctx, client, conf)
	}

	cookie, err := LoadPersistedCookie(KGWAuthTokenFilePath(), clientConfig.Signer.Identity())
	if err == nil && cookie != nil {
		// if setting fails, then don't do fail usage- failure likely means that the client has
		// switched providers, and the cookie is no longer valid.  The gatewayclient will re-authenticate.
		// delete the cookie if it is invalid
		err = client.SetAuthCookie(cookie)
		if err != nil {
			err2 := DeleteCookie(KGWAuthTokenFilePath(), clientConfig.Signer.Identity())
			if err2 != nil {
				return fmt.Errorf("failed to delete cookie: %w", err2)
			}

			display.PrintCmd(cmd, display.RespString(fmt.Sprintf("deleted invalid persisted cookie for user %s", hex.EncodeToString(clientConfig.Signer.Identity()))))

		}
	}

	return fn(ctx, client, conf)
}

// Client is a client that can be used to talk to a kwil node
type Client interface {
	CallAction(ctx context.Context, dbid string, action string, inputs []any) (*client.Records, error)
	ChainID() string
	ChainInfo(ctx context.Context) (*types.ChainInfo, error)
	DeployDatabase(ctx context.Context, payload *transactions.Schema, opts ...client.TxOpt) (transactions.TxHash, error)
	DropDatabase(ctx context.Context, name string, opts ...client.TxOpt) (transactions.TxHash, error)
	DropDatabaseID(ctx context.Context, dbid string, opts ...client.TxOpt) (transactions.TxHash, error)
	ExecuteAction(ctx context.Context, dbid string, action string, tuples [][]any, opts ...client.TxOpt) (transactions.TxHash, error)
	GetAccount(ctx context.Context, pubKey []byte, status types.AccountStatus) (*types.Account, error)
	GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error)
	ListDatabases(ctx context.Context, owner []byte) ([]string, error)
	Ping(ctx context.Context) (string, error)
	Query(ctx context.Context, dbid string, query string) (*client.Records, error)
	TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error)
	WaitTx(ctx context.Context, txHash []byte, interval time.Duration) (*transactions.TcTxQueryResponse, error)
}

// promptMessage prompts the user to sign a message. Return an error if user
// declines to sign.
func promptMessage(msg string) error {
	// display the message to user
	fmt.Println(msg)

	prompt := promptui.Prompt{
		Label:     "Do you want to sign this message?",
		IsConfirm: true,
	}

	_, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("you declined to sign")
	}

	return nil
}
