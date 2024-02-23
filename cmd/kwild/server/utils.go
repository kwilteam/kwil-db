package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math"
	"strings"

	types "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/extensions/actions"
	"github.com/kwilteam/kwil-db/internal/abci"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft/privval"
	"github.com/kwilteam/kwil-db/internal/extensions"
	"github.com/kwilteam/kwil-db/internal/kv"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/p2p"
	cmtlocal "github.com/cometbft/cometbft/rpc/client/local"
	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
)

// getExtensions returns both the local and remote extensions. Remote extensions are identified by
// connecting to the specified extension URLs.
func getExtensions(ctx context.Context, urls []string) (map[string]actions.ExtensionInitializer, error) {
	exts := make(map[string]actions.ExtensionInitializer)

	for name, ext := range actions.RegisteredExtensions() {
		_, ok := exts[name]
		if ok {
			return nil, fmt.Errorf("duplicate extension name: %s", name)
		}
		exts[name] = ext
	}

	for _, url := range urls {
		ext := extensions.New(url)
		err := ext.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to connect extension '%s': %w", ext.Name(), err)
		}

		_, ok := exts[ext.Name()]
		if ok {
			return nil, fmt.Errorf("duplicate extension name: %s", ext.Name())
		}

		exts[ext.Name()] = extensions.AdaptLegacyExtension(ext)
	}
	return exts, nil
}

// wrappedCometBFTClient satisfies the generic txsvc.BlockchainBroadcaster and
// admsvc.Node interfaces, hiding the details of cometBFT.
type wrappedCometBFTClient struct {
	cl *cmtlocal.Local
}

func convertNodeInfo(ni *p2p.DefaultNodeInfo) *types.NodeInfo {
	return &types.NodeInfo{
		ChainID:         ni.Network,
		Name:            ni.Moniker,
		NodeID:          string(ni.ID()),
		ProtocolVersion: ni.ProtocolVersion.P2P,
		AppVersion:      ni.ProtocolVersion.App,
		BlockVersion:    ni.ProtocolVersion.Block,
		ListenAddr:      ni.ListenAddr,
		RPCAddr:         ni.Other.RPCAddress,
	}
}

func (wc *wrappedCometBFTClient) Peers(ctx context.Context) ([]*types.PeerInfo, error) {
	cmtNetInfo, err := wc.cl.NetInfo(ctx)
	if err != nil {
		return nil, err
	}

	peers := make([]*types.PeerInfo, len(cmtNetInfo.Peers))
	for i, p := range cmtNetInfo.Peers {
		peers[i] = &types.PeerInfo{
			NodeInfo:   convertNodeInfo(&p.NodeInfo),
			Inbound:    !p.IsOutbound,
			RemoteAddr: p.RemoteIP,
		}
	}
	return peers, nil
}

func (wc *wrappedCometBFTClient) Status(ctx context.Context) (*types.Status, error) {
	cmtStatus, err := wc.cl.Status(ctx)
	if err != nil {
		return nil, err
	}
	ni, si, vi := &cmtStatus.NodeInfo, &cmtStatus.SyncInfo, &cmtStatus.ValidatorInfo
	return &types.Status{
		Node: convertNodeInfo(ni),
		Sync: &types.SyncInfo{
			AppHash:         strings.ToLower(si.LatestAppHash.String()),
			BestBlockHash:   strings.ToLower(si.LatestBlockHash.String()),
			BestBlockHeight: si.LatestBlockHeight,
			BestBlockTime:   si.LatestBlockTime.UTC(),
		},
		Validator: &types.ValidatorInfo{
			PubKey: vi.PubKey.Bytes(),
			Power:  vi.VotingPower,
		},
	}, nil
}

func (wc *wrappedCometBFTClient) BroadcastTx(ctx context.Context, tx []byte, sync uint8) (*cmtCoreTypes.ResultBroadcastTx, error) {
	var bcastFun func(ctx context.Context, tx cmttypes.Tx) (*cmtCoreTypes.ResultBroadcastTx, error)
	switch sync {
	case 0:
		bcastFun = wc.cl.BroadcastTxAsync
	case 1:
		bcastFun = wc.cl.BroadcastTxSync
	case 2:
		bcastFun = func(ctx context.Context, tx cmttypes.Tx) (*cmtCoreTypes.ResultBroadcastTx, error) {
			res, err := wc.cl.BroadcastTxCommit(ctx, tx)
			if err != nil {
				return nil, err
			}
			if res.CheckTx.Code != abciTypes.CodeTypeOK {
				return &cmtCoreTypes.ResultBroadcastTx{
					Code:      res.CheckTx.Code,
					Data:      res.CheckTx.Data,
					Log:       res.CheckTx.Log,
					Codespace: res.CheckTx.Codespace,
					Hash:      res.Hash,
				}, nil
			}
			return &cmtCoreTypes.ResultBroadcastTx{
				Code:      res.TxResult.Code,
				Data:      res.TxResult.Data,
				Log:       res.TxResult.Log,
				Codespace: res.TxResult.Codespace,
				Hash:      res.Hash,
			}, nil
		}
	}

	return bcastFun(ctx, cmttypes.Tx(tx))
}

// TxQuery locates a transaction in the node's blockchain or mempool. If the
// transaction could not be located, and error of type internal/abci.ErrTxNotFound is
// returned.
func (wc *wrappedCometBFTClient) TxQuery(ctx context.Context, hash []byte, prove bool) (*cmtCoreTypes.ResultTx, error) {
	// First check confirmed transactions. The Tx method of the cometbft client
	// does not define a specific exported error for a transaction that is not
	// found, just a "tx (%X) not found" as of cometbft v0.37. The Tx docs also
	// indicate that "`nil` could mean the transaction is in the mempool", so
	// this API should be used with caution, not failing on error AND checking
	// the result for nilness.
	res, err := wc.cl.Tx(ctx, hash, prove)
	if err == nil && res != nil {
		return res, nil
	}

	// The transaction could be in mempool.
	limit := math.MaxInt                             // cmt is bugged, -1 doesn't actually work (see rpc/core.validatePerPage and how it goes with 30 instead of no limit)
	unconf, err := wc.cl.UnconfirmedTxs(ctx, &limit) // SLOW quite often!
	if err != nil {
		return nil, err
	}
	for _, tx := range unconf.Txs {
		if bytes.Equal(tx.Hash(), hash) {
			// Found it. Shoe-horn into a ResultTx with -1 height, and the zero
			// values for ResponseDeliverTx and TxProof (it's checked and
			// accepted to mempool, but not delivered in a block yet).
			return &cmtCoreTypes.ResultTx{
				Hash:   hash,
				Height: -1,
				Tx:     tx,
			}, nil
		}
	}
	return nil, abci.ErrTxNotFound
}

// atomicReadWriter implements the CometBFT AtomicReadWriter interface.
// This should probably be done with a file instead of a KV store,
// but we already have a good implementation of an atomic KV store.
type atomicReadWriter struct {
	kv  kv.KVStore
	key []byte
}

var _ privval.AtomicReadWriter = (*atomicReadWriter)(nil)

func (a *atomicReadWriter) Read() ([]byte, error) {
	res, err := a.kv.Get(a.key)
	if errors.Is(err, kv.ErrKeyNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (a *atomicReadWriter) Write(val []byte) error {
	return a.kv.Set(a.key, val)
}
