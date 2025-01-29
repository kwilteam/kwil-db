// package evmsync its a tool to automate synchronizing state from and EVM chain into Kwil.
package evmsync

import (
	"context"
	_ "embed"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/kwilteam/kwil-db/common"
	"github.com/kwilteam/kwil-db/extensions/listeners"
	"github.com/kwilteam/kwil-db/extensions/resolutions"
	"github.com/kwilteam/kwil-db/node/exts/evm-sync/chains"
)

var (
	//go:embed schema.sql
	schema []byte
)

func init() {
	listeners.RegisterListener()

	resolutions.RegisterResolution()
}

// EVMListener listens for events from an EVM chain.
// It allows you to configure which contract addresses and event signatures to listen to,
// as well as how to resolve the event data post-consensus.
// It will guarantee the order of events and ensure that the resolve func is called
// only once per event, in the order that the events were emitted.
type EVMListener struct {
	// ContractAddresses is a list of contract addresses to listen to events from.
	ContractAddresses []string
	// EventSignatures is a list of event signatures to listen to.
	// All events from any contract configured matching any of these signatures will be emitted.
	// It is optional and defaults to all events.
	EventSignatures []string
	// StartHeight is the block height to start syncing from when the node starts.
	// It can be used to skip syncing parts of the chain that are not relevant (e.g.
	// if you are tracking a contract that was deployed at block height 1000000, you
	// can set this to 1000000 to skip syncing the entire chain up to that point).
	// It is optional and defaults to 0.
	StartHeight uint64
	// ExtensionName is the unique name of the extension.
	// It is optional and defaults to "eth_listener".
	ExtensionName string
	// ConfigName is the name of the configuration for the listener.
	// It is optional and defaults to "eth_listener".
	ConfigName string
	// Chain is the chain that the listener is listening to.
	Chain chains.Chain
	// RefundThreshold is the required vote percentage threshold for
	// all voters on a resolution to be refunded the gas costs
	// associated with voting. This allows for resolutions that have
	// not received enough votes to pass to refund gas to the voters
	// that have voted on the resolution. For a 1/3rds threshold,
	// >=1/3rds of the voters must vote for the resolution for
	// refunds to occur. If this threshold is not met, voters will
	// not be refunded when the resolution expires. The number must
	// be a fraction between 0 and 1. If this field is nil, it will
	// default to only refunding voters when the resolution is confirmed.
	RefundThreshold *big.Rat
	// ConfirmationThreshold is the required vote percentage
	// threshold for whether a resolution is confirmed. In a 2/3rds
	// threshold, >=2/3rds of the voters must vote for the resolution
	// for it to be confirmed. Voters will also be refunded if this
	// threshold is met, regardless of the refund threshold. The
	// number must be a fraction between 0 and 1. If this field is
	// nil, it will default to 2/3.
	ConfirmationThreshold *big.Rat
	// ExpirationPeriod is the amount of blocks that the resolution
	// will be valid for before it expires. It is applied additively
	// to the current block height when the resolution is proposed;
	// if the current block height is 10 and the expiration height is
	// 5, the resolution will expire at block 15. If this field is
	// <1, it will default to 14400, which is approximately 1 day
	// assuming 6 second blocks.
	ExpirationPeriod int64
	// Resolve is a function that resolves the event data post-consensus.
	// It can be used to update the state of the application based on the
	// event data. It is required.
	Resolve func(ctx context.Context, app *common.App, log types.Log, block *common.BlockContext) error
}
