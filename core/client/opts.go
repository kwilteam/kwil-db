package client

import (
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/log"
)

// ClientOptions are options that can be set for the client
type ClientOptions struct {
	// Logger is the logger to use for the client.
	Logger log.Logger

	// Signer will be used to sign transactions and set the Sender field on call messages.
	Signer auth.Signer

	// The chain ID will be used in all transactions, which helps prevent replay attacks on
	// different chains. On the initial connection, the remote node's chain ID is
	// checked against ours to ensure were are on the right network. If the chain ID
	// is empty, we will create and sign transactions for whatever network the
	// remote node claims, which should only be done for testing or when in secure
	// communication with a trusted node (using TLS or Unix sockets).
	ChainID string

	// Silence silences warnings logged from the client.
	Silence bool
}

// Apply applies the passed options to the receiver.
func (c *ClientOptions) Apply(opts *ClientOptions) {
	if opts == nil {
		return
	}

	if opts.Logger.L != nil {
		c.Logger = opts.Logger
	}

	if opts.Signer != nil {
		c.Signer = opts.Signer
	}

	if opts.ChainID != "" {
		c.ChainID = opts.ChainID
	}

	c.Silence = opts.Silence
}

// DefaultOptions returns the default options for the client.
func DefaultOptions() *ClientOptions {
	return &ClientOptions{
		Logger: log.NewNoOp(),
	}
}

type Option func(*ClientOptions)

func WithLogger(logger log.Logger) Option {
	return func(c *ClientOptions) {
		c.Logger = logger
	}
}

// WithSigner sets a signer to use when authoring transactions.
func WithSigner(signer auth.Signer) Option {
	return func(c *ClientOptions) {
		c.Signer = signer
	}
}

// WithChainID sets the chain ID to use when authoring transactions. The chain ID
// will be used in all transactions, which helps prevent replay attacks on
// different chains. On the initial connection, the remote node's chain ID is
// checked against ours to ensure were are on the right network. If the chain ID
// is empty, we will create and sign transactions for whatever network the
// remote node claims, which should only be done for testing or when in secure
// communication with a trusted node (using TLS or Unix sockets).
func WithChainID(chainID string) Option {
	return func(c *ClientOptions) {
		c.ChainID = chainID
	}
}

// SilenceWarnings silences warnings from the client.
func SilenceWarnings() Option {
	return func(c *ClientOptions) {
		c.Silence = true
	}
}

type txOptions struct {
	nonce int64
	fee   *big.Int

	syncBcast bool // wait for mining on broadcast
}

func getTxOpts(opts []TxOpt) *txOptions {
	txOpts := &txOptions{}
	for _, opt := range opts {
		opt(txOpts)
	}
	return txOpts
}

// TxOpt sets an option used when making and broadcasting a transaction.
type TxOpt func(*txOptions)

// WithNonce sets the nonce to use for the transaction.
func WithNonce(nonce int64) TxOpt {
	return func(o *txOptions) {
		o.nonce = nonce
	}
}

// WithFee sets the fee to use on the transaction, otherwise an EstimateCode RPC
// will be performed for the action.
func WithFee(fee *big.Int) TxOpt {
	return func(o *txOptions) {
		o.fee = fee
	}
}

// WithSyncBroadcast indicates that broadcast should wait for the transaction to
// be included in a block, not merely accepted into mempool.
func WithSyncBroadcast() TxOpt {
	return func(o *txOptions) {
		o.syncBcast = true
	}
}
