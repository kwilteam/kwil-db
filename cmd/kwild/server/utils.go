package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/kwilteam/kwil-db/core/log"
	types "github.com/kwilteam/kwil-db/core/types/admin"
	"github.com/kwilteam/kwil-db/extensions/precompiles"
	"github.com/kwilteam/kwil-db/internal/abci"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft/privval"
	remoteExtn "github.com/kwilteam/kwil-db/internal/extensions"
	"github.com/kwilteam/kwil-db/internal/kv"

	abciTypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/p2p"
	cmtlocal "github.com/cometbft/cometbft/rpc/client/local"
	cmtCoreTypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"
)

func increaseLogLevel(name string, logger *log.Logger, level string) *log.Logger {
	logger = logger.Named(name)
	if level == "" {
		return logger
	}

	lvl, err := log.ParseLevel(level)
	if err != nil {
		logger.Warnf("invalid log level %q for logger %q: %v", level, name, err)
		return logger
	}

	if parentLevel := logger.Level(); lvl < parentLevel {
		logger.Warnf("cannot increase logger level for %q to %v from %v", name, level, parentLevel)
	} else { // this would be a no-op
		logger = logger.IncreasedLevel(lvl)
	}

	return logger
}

// getExtensions returns both the local and remote extensions. Remote extensions are identified by
// connecting to the specified extension URLs.
func getExtensions(ctx context.Context, urls []string) (map[string]precompiles.Initializer, error) {
	exts := make(map[string]precompiles.Initializer)

	for name, ext := range precompiles.RegisteredPrecompiles() {
		_, ok := exts[name]
		if ok {
			return nil, fmt.Errorf("duplicate extension name: %s", name)
		}
		exts[name] = ext
	}

	for _, url := range urls {
		ext := remoteExtn.New(url)
		err := ext.Connect(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to connect extension '%s': %w", ext.Name(), err)
		}

		_, ok := exts[ext.Name()]
		if ok {
			return nil, fmt.Errorf("duplicate extension name: %s", ext.Name())
		}

		exts[ext.Name()] = remoteExtn.AdaptLegacyExtension(ext)
	}
	return exts, nil
}

// wrappedCometBFTClient satisfies the generic txsvc.BlockchainBroadcaster and
// admsvc.Node interfaces, hiding the details of cometBFT.
type wrappedCometBFTClient struct {
	cl    *cmtlocal.Local
	cache mempoolCache
}

type mempoolCache interface {
	TxInMempool([]byte) bool
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
	// chain / cometbft block store status
	cmtStatus, err := wc.cl.Status(ctx)
	if err != nil {
		return nil, err
	}

	// application status
	abciInfo, err := wc.cl.ABCIInfo(ctx)
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
			Syncing:         si.CatchingUp,
		},
		Validator: &types.ValidatorInfo{
			PubKey: vi.PubKey.Bytes(),
			Power:  vi.VotingPower,
		},
		App: &types.AppInfo{
			Height:  abciInfo.Response.LastBlockHeight,
			AppHash: abciInfo.Response.LastBlockAppHash,
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
				if res != nil { // seriously, they do this
					return &cmtCoreTypes.ResultBroadcastTx{
						Code:      res.CheckTx.Code,
						Data:      res.CheckTx.Data,
						Log:       res.CheckTx.Log,
						Codespace: res.CheckTx.Codespace,
						Hash:      res.Hash,
					}, err
				}
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

	// The transaction could be in the mempool, Check with ABCI directly if it heard of the transaction.
	if wc.cache.TxInMempool(hash) {
		return &cmtCoreTypes.ResultTx{
			Hash:   hash,
			Height: -1,
			Tx:     nil, // The transaction is still in the mempool, so not indexed yet. Returning nil to avoid hash computations on all the transactions in the mempool (potential DoS attack vector).
		}, nil
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

// getPostgresMajorVersion retrieve the major version number of postgres client tools (e.g., psql or pg_dump)
func getPostgresMajorVersion(command string) (int, error) {
	cmd := exec.Command(command, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return -1, fmt.Errorf("failed to execute %s: %w", command, err)
	}

	major, _, err := getPGVersion(out.String())
	if err != nil {
		return -1, fmt.Errorf("failed to get version: %w", err)
	}

	return major, nil
}

// getPGVersion extracts the major and minor version numbers from the version output of a PostgreSQL client tool.
func getPGVersion(versionOutput string) (int, int, error) {
	// Expected output format:
	// Mac OS X: psql (PostgreSQL) 16.0
	// Linux: psql (PostgreSQL) 16.4 (Ubuntu 16.4-1.pgdg22.04+1)
	re := regexp.MustCompile(`\(PostgreSQL\) (\d+)\.(\d+)(?:\.(\d+))?`)
	matches := re.FindStringSubmatch(versionOutput)

	if len(matches) == 0 {
		return -1, -1, fmt.Errorf("could not find a valid version in output: %s", versionOutput)
	}

	// Extract major version number
	major, err := strconv.Atoi(matches[1])
	if err != nil {
		return -1, -1, fmt.Errorf("failed to parse major version: %w", err)
	}

	// Extract minor version number
	minor, err := strconv.Atoi(matches[2])
	if err != nil {
		return -1, -1, fmt.Errorf("failed to parse minor version: %w", err)
	}

	return major, minor, nil
}

const (
	PGVersion = 16
)

// checkVersion validates the version of a PostgreSQL client tool against the expected version.
func checkVersion(command string, version int) error {
	major, err := getPostgresMajorVersion(command)
	if err != nil {
		return err
	}

	if major != version {
		return fmt.Errorf("expected %s version %d, got %d", command, version, major)
	}

	return nil
}
