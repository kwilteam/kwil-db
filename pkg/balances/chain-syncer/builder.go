package chainsyncer

import (
	"fmt"
	"kwil/pkg/chain/contracts/escrow"
	"kwil/pkg/log"
	"kwil/pkg/utils/retry"
	"time"
)

// trying a new builder pattern type here
// reference can be found at https://devcharmander.medium.com/design-patterns-in-golang-the-builder-dac468a71194

// post-mortem: I like it because it is pretty readable and gives a nice API, however unlike the options pattern,
// it doesn't show what is required until you call build, which is a bit annoying.  However, given the amount of
// configuration that is required (and the fact that the consumer will probably want to use most optional configs in ChainSyncer),
// I think it is a good tradeoff.

func Builder() *ChainSyncBuilder {
	return &ChainSyncBuilder{
		syncer: &ChainSyncer{
			log:             log.NewNoOp(),
			height:          default_start_height,
			chunkSize:       default_chunk_size,
			receiverAddress: "",
			escrowAddress:   "",
		},
	}
}

type ChainSyncBuilder struct {
	syncer *ChainSyncer
}

type ChainSyncAccountRepoBuilder struct {
	*ChainSyncBuilder
}

type ChainSyncChainClientBuilder struct {
	*ChainSyncBuilder
}

// WithLogger sets the logger for the chain syncer
func (c *ChainSyncBuilder) WithLogger(logger log.Logger) *ChainSyncBuilder {
	c.syncer.log = logger
	return c
}

// WritesTo specifies the account repository that the chain syncer will write to
func (c *ChainSyncBuilder) WritesTo(repo accountRepository) *ChainSyncAccountRepoBuilder {
	c.syncer.accountRepository = repo
	return &ChainSyncAccountRepoBuilder{c}
}

// ListensTo specifies the address of the escrow contract that the chain syncer will listen to
func (c *ChainSyncBuilder) ListensTo(address string) *ChainSyncChainClientBuilder {
	c.syncer.escrowAddress = address
	return &ChainSyncChainClientBuilder{c}
}

// WithChainClient sets the chain client that the chain syncer will use
func (c *ChainSyncChainClientBuilder) WithChainClient(client chainClient) *ChainSyncChainClientBuilder {
	c.syncer.chainClient = client
	return c
}

// WithStartingHeight sets the starting height of the chain syncer
func (c *ChainSyncChainClientBuilder) WithStartingHeight(height int64) *ChainSyncChainClientBuilder {
	c.syncer.height = height
	return c
}

// WithChunkSize sets the chunk size of the chain syncer
func (c *ChainSyncChainClientBuilder) WithChunkSize(size int64) *ChainSyncChainClientBuilder {
	c.syncer.chunkSize = size
	return c
}

// WithReceiverAddress sets the receiver address of the chain syncer
func (c *ChainSyncChainClientBuilder) WithReceiverAddress(address string) *ChainSyncChainClientBuilder {
	c.syncer.receiverAddress = address
	return c
}

// Build builds the chain syncer
func (c *ChainSyncBuilder) Build() (*ChainSyncer, error) {
	if c.syncer.accountRepository == nil {
		return nil, fmt.Errorf("account repository not set")
	}
	if c.syncer.chainClient == nil {
		return nil, fmt.Errorf("chain client not set")
	}
	if c.syncer.escrowAddress == "" {
		return nil, fmt.Errorf("escrow address not set")
	}
	if c.syncer.receiverAddress == "" {
		return nil, fmt.Errorf("deposit receiver address not set")
	}

	escrowCtr, err := c.syncer.chainClient.Contracts().Escrow(c.syncer.escrowAddress)
	if err != nil {
		return nil, err
	}

	c.syncer.escrowContract = escrowCtr
	c.syncer.chainCode = c.syncer.chainClient.ChainCode()

	err = c.ensureChainExistsInDB(c.syncer.chainCode.Int32(), c.syncer.height)
	if err != nil {
		return nil, err
	}

	c.syncer.retrier = retry.New(escrowCtr,
		retry.WithLogger[escrow.EscrowContract](c.syncer.log),
		retry.WithFactor[escrow.EscrowContract](2),
		retry.WithMin[escrow.EscrowContract](time.Second*1),
		retry.WithMax[escrow.EscrowContract](time.Second*10),
	)

	return c.syncer, nil
}

// ensureChainExistsInDB ensures that the chain exists in the database
func (c *ChainSyncBuilder) ensureChainExistsInDB(chainCode int32, height int64) error {
	exists, err := c.syncer.accountRepository.ChainExists(chainCode)
	if err != nil {
		return err
	}

	if !exists {
		err = c.syncer.accountRepository.CreateChain(chainCode, height)
		if err != nil {
			return err
		}
	}

	return nil
}
