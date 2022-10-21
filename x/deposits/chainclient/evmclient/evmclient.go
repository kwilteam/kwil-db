package evmclient

import (
	"context"

	esc "kwil/x/deposits/chainclient/evmclient/contracts"
	ct "kwil/x/deposits/chainclient/types"
	"kwil/x/logx"

	"github.com/ethereum/go-ethereum/core/types"
	ethc "github.com/ethereum/go-ethereum/ethclient"
)

type ethClient struct {
	client *ethc.Client
	log    logx.SugaredLogger
}

func New(l logx.Logger, endpoint, chainCode string) (*ethClient, error) {

	client, err := ethc.Dial(endpoint)
	log := l.Sugar().With("chain", chainCode)
	if err != nil {
		log.Errorf("failed to connect to ethereum client: %v", err)
		return nil, err
	}

	return &ethClient{
		client: client,
		log:    log,
	}, nil
}

// SubscribeBlocks subscribes to new block heights on the chain
func (ec *ethClient) SubscribeBlocks(ctx context.Context, channel chan<- int64) (ct.BlockSubscription, error) {
	headerChan := make(chan *types.Header)
	sub, err := ec.client.SubscribeNewHead(ctx, headerChan)
	if err != nil {
		ec.log.Errorf("failed to subscribe to new block headers: %v", err)
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

func (ec *ethClient) GetLatestBlock(ctx context.Context) (int64, error) {
	h, err := ec.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}

	return h.Number.Int64(), nil
}
