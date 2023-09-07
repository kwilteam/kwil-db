package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/client/types"
	"os"

	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	engineUtils "github.com/kwilteam/kwil-db/pkg/engine/utils"
	grpcClient "github.com/kwilteam/kwil-db/pkg/grpc/client/v1"
	"github.com/kwilteam/kwil-db/pkg/log"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"github.com/kwilteam/kwil-db/pkg/validators"

	"github.com/cstockton/go-conv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	grpcCodes "google.golang.org/grpc/codes"
	grpcCreds "google.golang.org/grpc/credentials"
	grpcInsecure "google.golang.org/grpc/credentials/insecure"
	grpcStatus "google.golang.org/grpc/status"
)

var (
	ErrNotFound = errors.New("not found")
)

type Client struct {
	client   *grpcClient.Client
	datasets map[string]*transactions.Schema
	Signer   crypto.Signer
	logger   log.Logger
	certFile string // the TLS certificate for the grpc Client
}

func newTLSConfig(certFile string) (*tls.Config, error) {
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}
	// NOTE: we're testing a special case of "-" meaning use TLS, but just use
	// system CAs without appending a known server certificate. We may change
	// this or formally document it.
	if certFile != "-" {
		b, err := os.ReadFile(certFile)
		if err != nil {
			return nil, err
		}
		if !rootCAs.AppendCertsFromPEM(b) {
			return nil, fmt.Errorf("credentials: failed to append certificates")
		}
	}
	return &tls.Config{
		// For proper verification of the server-provided certificate chain
		// during the TLS handshake, the root CAs, which may contain a custom
		// certificate we appended above, are used by the client tls.Conn. If we
		// disable verification with InsecureSkipVerify, the connection is still
		// encrypted, but we cannot ensure the server is who they claim to be.
		RootCAs:    rootCAs,
		MinVersion: tls.VersionTLS12,
	}, nil
}

// New creates a new client
func New(host string, opts ...ClientOpt) (c *Client, err error) {
	c = &Client{
		datasets: make(map[string]*transactions.Schema),
		logger:   log.NewNoOp(), // by default we do not want to force client to log anything
	}

	for _, opt := range opts {
		opt(c)
	}

	var transOpt grpcCreds.TransportCredentials
	if c.certFile == "" {
		transOpt = grpcInsecure.NewCredentials()
	} else {
		tlsConfig, err := newTLSConfig(c.certFile)
		if err != nil {
			return nil, err
		}
		transOpt = grpcCreds.NewTLS(tlsConfig)
	}

	c.client, err = grpcClient.New(host, grpc.WithTransportCredentials(transOpt))
	if err != nil {
		return nil, err
	}

	zapFields := []zapcore.Field{
		zap.String("host", host),
	}
	if c.Signer != nil {
		zapFields = append(zapFields, zap.String("from", c.Signer.PubKey().Address().String()))
	}

	c.logger = *c.logger.Named("client").With(zapFields...)

	return c, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

// GetSchema returns the entity of a database
func (c *Client) GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error) {
	ds, ok := c.datasets[dbid]
	if ok {
		return ds, nil
	}

	ds, err := c.client.GetSchema(ctx, dbid)
	if err != nil {
		return nil, err
	}

	c.datasets[dbid] = ds
	return ds, nil
}

// DeployDatabase deploys a schema
func (c *Client) DeployDatabase(ctx context.Context, payload *transactions.Schema) (transactions.TxHash, error) {
	tx, err := c.newTx(ctx, payload)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("deploying database",
		zap.String("signature_type", tx.Signature.Type.String()),
		zap.String("signature", base64.StdEncoding.EncodeToString(tx.Signature.Signature)))

	return c.client.Broadcast(ctx, tx)
}

// DropDatabase drops a database
func (c *Client) DropDatabase(ctx context.Context, name string) (transactions.TxHash, error) {
	pub, err := c.getPublicKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get address from private key: %w", err)
	}

	identifier := &transactions.DropSchema{
		DBID: engineUtils.GenerateDBID(name, pub.Bytes()),
	}

	tx, err := c.newTx(ctx, identifier)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("deploying database",
		zap.String("signature_type", tx.Signature.Type.String()),
		zap.String("signature", base64.StdEncoding.EncodeToString(tx.Signature.Signature)))

	res, err := c.client.Broadcast(ctx, tx)
	if err != nil {
		return nil, err
	}

	delete(c.datasets, identifier.DBID)

	return res, nil
}

// ExecuteAction executes an action.
// It returns the receipt, as well as outputs which is the decoded body of the receipt.
// It can take any number of inputs, and if multiple tuples of inputs are passed, it will execute them transactionally.
func (c *Client) ExecuteAction(ctx context.Context, dbid string, action string, tuples ...[]any) (transactions.TxHash, error) {
	stringTuples, err := convertTuples(tuples)
	if err != nil {
		return nil, err
	}

	executionBody := &transactions.ActionExecution{
		Action:    action,
		DBID:      dbid,
		Arguments: stringTuples,
	}

	tx, err := c.newTx(ctx, executionBody)
	if err != nil {
		return nil, err
	}

	c.logger.Debug("deploying database",
		zap.String("signature_type", tx.Signature.Type.String()),
		zap.String("signature", base64.StdEncoding.EncodeToString(tx.Signature.Signature)))

	return c.client.Broadcast(ctx, tx)
}

// CallAction call an action, if auxiliary `mustsign` is set, need to sign the action payload. It returns the records.
func (c *Client) CallAction(ctx context.Context, dbid string, action string, inputs []any, opts ...CallOpt) (*Records, error) {
	callOpts := &callOptions{}

	for _, opt := range opts {
		opt(callOpts)
	}

	stringInputs, err := convertTuple(inputs)
	if err != nil {
		return nil, err
	}

	payload := &transactions.ActionCall{
		DBID:      dbid,
		Action:    action,
		Arguments: stringInputs,
	}

	shouldSign, err := shouldAuthenticate(c.Signer, callOpts.forceAuthenticated)
	if err != nil {
		return nil, err
	}

	msg, err := transactions.CreateSignedMessage(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create signed message: %w", err)
	}

	if shouldSign {
		err = msg.Sign(c.Signer)

		if err != nil {
			return nil, fmt.Errorf("failed to create signed message: %w", err)
		}
	}

	res, err := c.client.Call(ctx, msg)
	if err != nil {
		return nil, err
	}

	return NewRecordsFromMaps(res), nil
}

// shouldAuthenticate decides whether the client should authenticate or not
// if enforced is not nil, it will be used instead of the default value
// otherwise, if the private key is not nil, it will authenticate
func shouldAuthenticate(signer crypto.Signer, enforced *bool) (bool, error) {
	if enforced != nil {
		if !*enforced {
			return false, nil
		}

		if signer == nil {
			return false, fmt.Errorf("private key is nil, but authentication is enforced")
		}

		return true, nil
	}

	return signer != nil, nil
}

func DecodeOutputs(bts []byte) ([]map[string]any, error) {
	if len(bts) == 0 {
		return []map[string]any{}, nil
	}

	var outputs []map[string]any
	err := json.Unmarshal(bts, &outputs)
	if err != nil {
		return nil, err
	}

	return outputs, nil
}

// Query executes a query
func (c *Client) Query(ctx context.Context, dbid string, query string) (*Records, error) {
	res, err := c.client.Query(ctx, dbid, query)
	if err != nil {
		return nil, err
	}

	return NewRecordsFromMaps(res), nil
}

func (c *Client) ListDatabases(ctx context.Context, ownerPubKey []byte) ([]string, error) {
	return c.client.ListDatabases(ctx, ownerPubKey)
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	return c.client.Ping(ctx)
}

func (c *Client) GetAccount(ctx context.Context, pubKey []byte) (*balances.Account, error) {
	return c.client.GetAccount(ctx, pubKey)
}

func (c *Client) ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*validators.JoinRequest, error) {
	res, err := c.client.ValidatorJoinStatus(ctx, pubKey)
	if err != nil {
		if stat, ok := grpcStatus.FromError(err); ok {
			if stat.Code() == grpcCodes.NotFound {
				return nil, ErrNotFound
			}
		}
		return nil, err
	}
	return res, nil
}

func (c *Client) CurrentValidators(ctx context.Context) ([]*validators.Validator, error) {
	return c.client.CurrentValidators(ctx)
}

func (c *Client) ApproveValidator(ctx context.Context, joiner []byte) ([]byte, error) {
	_, err := crypto.Ed25519PublicKeyFromBytes(joiner)
	if err != nil {
		return nil, fmt.Errorf("invalid candidate validator public key: %w", err)
	}
	payload := &transactions.ValidatorApprove{
		Candidate: joiner,
	}
	tx, err := c.newTx(ctx, payload)
	if err != nil {
		return nil, err
	}

	hash, err := c.client.Broadcast(ctx, tx)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

func (c *Client) ValidatorJoin(ctx context.Context) ([]byte, error) {
	const power = 1
	return c.validatorUpdate(ctx, power)
}

func (c *Client) ValidatorLeave(ctx context.Context) ([]byte, error) {
	return c.validatorUpdate(ctx, 0)
}

func (c *Client) validatorUpdate(ctx context.Context, power int64) ([]byte, error) {
	pubKey := c.Signer.PubKey().Bytes()

	var payload transactions.Payload
	if power <= 0 {
		payload = &transactions.ValidatorLeave{
			Validator: pubKey,
		}
	} else {
		payload = &transactions.ValidatorJoin{
			Candidate: pubKey,
			Power:     uint64(power),
		}
	}

	tx, err := c.newTx(ctx, payload)
	if err != nil {
		return nil, err
	}

	hash, err := c.client.Broadcast(ctx, tx)
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// convertTuples converts user passed tuples to strings.
// this is necessary for RLP encoding
func convertTuples(tuples [][]any) ([][]string, error) {
	ins := [][]string{}
	for _, tuple := range tuples {
		stringTuple, err := convertTuple(tuple)
		if err != nil {
			return nil, err
		}
		ins = append(ins, stringTuple)
	}

	return ins, nil
}

// convertTuple converts user passed tuple to strings.
func convertTuple(tuple []any) ([]string, error) {
	stringTuple := []string{}
	for _, val := range tuple {

		stringVal, err := conv.String(val)
		if err != nil {
			return nil, err
		}

		stringTuple = append(stringTuple, stringVal)
	}

	return stringTuple, nil
}

// TxQuery get transaction by hash
func (c *Client) TxQuery(ctx context.Context, txHash []byte) (*types.TxQueryResponse, error) {
	res, err := c.client.TxQuery(ctx, txHash)
	if err != nil {
		return nil, err
	}

	return res, nil
}
