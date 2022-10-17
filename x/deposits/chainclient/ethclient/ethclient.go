package ethclient

import (
	"context"

	esc "kwil/x/deposits/chainclient/ethclient/contracts"
	ct "kwil/x/deposits/chainclient/types"

	"github.com/ethereum/go-ethereum/core/types"
	ethc "github.com/ethereum/go-ethereum/ethclient"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type ethClient struct {
	client *ethc.Client
	log    *zerolog.Logger
}

func New(endpoint, chainCode string) (ct.Client, error) {
	logger := log.With().Str("module", "ethclient").Str("chain_code", chainCode).Logger()

	client, err := ethc.Dial(endpoint)
	if err != nil {
		logger.Warn().Err(err).Msg("failed to connect to client chain")
		return nil, err
	}

	return &ethClient{
		client: client,
		log:    &logger,
	}, nil
}

// SubscribeBlocks subscribes to new block heights on the chain
func (ec *ethClient) SubscribeBlocks(ctx context.Context, channel chan<- int64) (ct.BlockSubscription, error) {
	headerChan := make(chan *types.Header)
	sub, err := ec.client.SubscribeNewHead(ctx, headerChan)
	if err != nil {
		ec.log.Warn().Err(err).Msg("failed to subscribe to new blocks")
		return sub, err
	}

	// goroutine to convert the header channel to a block height channel
	go func() {
		for {
			select {
			case header := <-headerChan:
				channel <- header.Number.Int64()
			case <-ctx.Done():
				return
			}
		}
	}()

	return sub, nil
}

func (ec *ethClient) GetContract(addr string) (ct.Contract, error) {
	return esc.New(ec.client, addr)
}
