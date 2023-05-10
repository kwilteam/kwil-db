package chainsyncer

import (
	"context"
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/balances"
	"github.com/kwilteam/kwil-db/pkg/chain/contracts"
	"github.com/kwilteam/kwil-db/pkg/chain/contracts/escrow"
	provider "github.com/kwilteam/kwil-db/pkg/chain/provider/dto"
	chainCodes "github.com/kwilteam/kwil-db/pkg/chain/types"
	"github.com/kwilteam/kwil-db/pkg/log"
	"math/big"
	"time"

	"go.uber.org/zap"
)

const (
	default_start_height = 0
	default_chunk_size   = 100000
)

type ChainSyncer struct {
	log log.Logger

	// accountRepository is the interface that is used to interact with the account store
	accountRepository accountRepository

	// chainClient is the client that is used to interact with the chain
	chainClient chainClient

	// height is the height of the last block that has been synced
	height int64

	// chunkSize is the number of blocks that are synced in a single batch
	chunkSize int64

	// chainCode is the chain code of the chain that is being synced
	chainCode chainCodes.ChainCode

	// escrowContract is the escrow contract interface
	escrowContract escrow.EscrowContract

	// escrowAddress is the address of the escrow contract
	escrowAddress string

	// receiverAddress is the address of the deposit receiver
	// this will almost always be the address of the node's wallet
	receiverAddress string
}

type accountRepository interface {
	BatchCredit([]*balances.Credit, *balances.ChainConfig) error
	GetHeight(int32) (int64, error)
	CreateChain(chainCode int32, height int64) error
	ChainExists(chainCode int32) (bool, error)
}

type chainClient interface {
	ChainCode() chainCodes.ChainCode
	GetLatestBlock(ctx context.Context) (*provider.Header, error)
	Contracts() contracts.Contracter
	Listen(ctx context.Context, blocks chan<- int64) error
}

// Start starts the chain syncer
func (cs *ChainSyncer) Start(ctx context.Context) error {
	var err error
	cs.height, err = cs.getStartingHeight()
	if err != nil {
		return err
	}

	latestBlock, err := cs.chainClient.GetLatestBlock(ctx)
	if err != nil {
		return err
	}

	cs.log.Info("Starting sync from height", zap.Int64("starting_height", cs.height), zap.Int64("latest_height", latestBlock.Height))
	chunks := splitBlocks(cs.height, latestBlock.Height, cs.chunkSize)

	for _, chunk := range chunks {
		err = cs.syncChunk(ctx, chunk)
		if err != nil {
			return err
		}

		time.Sleep(100 * time.Millisecond) // alchemy doesn't like it when we make too many requests in a short period of time
	}

	return cs.listen(ctx)
}

// getLastHeight retrieves the last synced height from the account repository
func (cs *ChainSyncer) getLastHeight() (int64, error) {
	return cs.accountRepository.GetHeight(cs.chainCode.Int32())
}

// getStartingHeight will return either the last synced height or the configured start height,
// depending on which is greater
func (cs *ChainSyncer) getStartingHeight() (int64, error) {
	lastHeight, err := cs.getLastHeight()
	if err != nil {
		return 0, err
	}

	if lastHeight > cs.height {
		return lastHeight, nil
	}

	return cs.height, nil
}

type chunkRange [2]int64

func splitBlocks(start, end, chunkSize int64) []chunkRange {
	if start == end {
		return []chunkRange{{start, start}}
	}
	var chunks []chunkRange
	for i := start; i < end; i += chunkSize {
		chunkEnd := i + chunkSize
		if chunkEnd > end {
			chunkEnd = end
		}
		chunks = append(chunks, chunkRange{i, chunkEnd - 1})
	}

	if chunks[len(chunks)-1][1] != end {
		chunks[len(chunks)-1][1] = end
	}
	return chunks
}

// getCreditsForRange retrieves all deposits for a given range of blocks, and returns them as credits
func (cs *ChainSyncer) getCreditsForRange(ctx context.Context, start, end int64) ([]*balances.Credit, error) {
	deposits, err := cs.escrowContract.GetDeposits(ctx, start, end, cs.receiverAddress)
	if err != nil {
		return nil, err
	}

	credits := make([]*balances.Credit, len(deposits))
	for i, deposit := range deposits {
		bigAmount, ok := new(big.Int).SetString(deposit.Amount, 10)
		if !ok {
			return nil, fmt.Errorf("failed to parse amount %s", deposit.Amount)
		}

		credits[i] = &balances.Credit{
			AccountAddress: deposit.Caller,
			Amount:         bigAmount,
		}
	}

	return credits, nil
}

// syncChunk syncs a chunk of blocks
func (cs *ChainSyncer) syncChunk(ctx context.Context, chunk chunkRange) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	cs.log.Debug("Syncing chunk", zap.Int64("start", chunk[0]), zap.Int64("end", chunk[1]))

	credits, err := cs.getCreditsForRange(ctx, chunk[0], chunk[1])
	if err != nil {
		return err
	}

	err = cs.accountRepository.BatchCredit(credits, &balances.ChainConfig{
		ChainCode: cs.chainCode.Int32(),
		Height:    chunk[1],
	})
	if err != nil {
		return err
	}

	cs.log.Debug("Synced chunk", zap.Int64("start", chunk[0]), zap.Int64("end", chunk[1]), zap.Int64("deposits in chunk", int64(len(credits))))
	return nil
}

// listen listens for new blocks on the chain and syncs them as they come in
func (cs *ChainSyncer) listen(ctx context.Context) error {
	blockChan := make(chan int64)
	err := cs.chainClient.Listen(ctx, blockChan)
	if err != nil {
		return err
	}

	go func(blockChan <-chan int64) {
		defer func() {
			if err := recover(); err != nil {
				cs.log.Error("Chain syncer panic", zap.Any("error", err))
			}
			cs.log.Info("Chain syncer stopped")
		}()
		for {
			select {
			case <-ctx.Done():
				cs.log.Info("stop Chain syncer")
				return
			case block := <-blockChan:
				cs.log.Debug("Received block", zap.Int64("block", block))

				credits, err := cs.getCreditsForRange(ctx, block, block)
				if err != nil {
					cs.log.Error("Failed to get credits for block", zap.Int64("block", block), zap.Error(err))
					return
				}

				cs.log.Debug("Syncing deposits from block", zap.Int64("block", block), zap.Int64("number of deposits:", int64(len(credits))))

				err = cs.accountRepository.BatchCredit(credits, &balances.ChainConfig{
					ChainCode: cs.chainCode.Int32(),
					Height:    block,
				})
				if err != nil {
					cs.log.Error("Failed to credit block", zap.Int64("block", block), zap.Error(err))
					return
				}
			}
		}
	}(blockChan)

	return nil
}
