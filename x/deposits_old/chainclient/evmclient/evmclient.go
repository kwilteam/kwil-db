package evmclient

import (
	"context"
	"errors"
	"math/big"

	esc "kwil/x/deposits_old/chainclient/evmclient/contracts"
	ct "kwil/x/deposits_old/types"
	"kwil/x/logx"

	"github.com/ethereum/go-ethereum/core/types"
	ethc "github.com/ethereum/go-ethereum/ethclient"
)

type ethClient struct {
	client *ethc.Client
	log    logx.SugaredLogger
	cid    *big.Int
}

func New(l logx.Logger, endpoint, chainCode string) (*ethClient, error) {

	client, err := ethc.Dial(endpoint)
	log := l.Sugar().With("chain", chainCode)
	if err != nil {
		log.Errorf("failed to connect to ethereum client: %v", err)
		return nil, err
	}

	cid, err := determineChainID(chainCode)
	if err != nil {
		log.Errorf("failed to determine chain id: %v", err)
		return nil, err
	}

	return &ethClient{
		client: client,
		log:    log,
		cid:    cid,
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
	return esc.New(ec.client, addr, ec.cid)
}

func (ec *ethClient) GetLatestBlock(ctx context.Context) (int64, error) {
	h, err := ec.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}

	return h.Number.Int64(), nil
}

func determineChainID(c string) (*big.Int, error) {
	switch c {
	default:
		return big.NewInt(0), ErrInvalidChain
	case "eth-mainnet":
		return big.NewInt(1), nil
	case "eth-goerli":
		return big.NewInt(5), nil
	}
}

var ErrInvalidChain = errors.New("invalid chain")
