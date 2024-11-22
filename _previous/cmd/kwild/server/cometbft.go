package server

import (
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	cmtCfg "github.com/cometbft/cometbft/config"
	cmtEd "github.com/cometbft/cometbft/crypto/ed25519"
	cmttypes "github.com/cometbft/cometbft/types"
	kconfig "github.com/kwilteam/kwil-db/cmd/kwild/config"
	"github.com/kwilteam/kwil-db/common/chain"
	config "github.com/kwilteam/kwil-db/common/config"
	"github.com/kwilteam/kwil-db/core/utils/url"
	"github.com/kwilteam/kwil-db/internal/abci/cometbft"
)

// cleanListenAddr tries to ensure the address has a scheme and port, as
// required by cometbft for its listen address settings. If it cannot parse, it
// is returned as-is so cometbft can try it (this is a best effort helper).
func cleanListenAddr(addr, defaultPort string) string {
	u, err := url.ParseURL(addr)
	if err != nil { // just see if cometbft takes it
		return addr
	}

	parsed := u.URL()
	if u.Port == 0 {
		// If port not included or explicitly set to 0, use the default.
		_, port, _ := net.SplitHostPort(u.Target)
		if port == "" {
			parsed.Host = net.JoinHostPort(u.Target, defaultPort)
		}
	}
	return parsed.String()
}

func portFromURL(u string) string {
	rpcAddress, err := url.ParseURL(u)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(rpcAddress.Port)
}

// newCometConfig creates a new CometBFT config for use with NewCometBftNode.
// This applies the operator's settings from the Kwil config as well as applying
// some overrides to the defaults for Kwil.
//
// NOTE: this is somewhat error prone, so care must be taken to update this
// function when fields are added to KwildConfig.ChainCfg.
func newCometConfig(cfg *config.KwildConfig) *cmtCfg.Config {
	// Begin with CometBFT's default chain config.
	nodeCfg := cmtCfg.DefaultConfig()

	// Override defaults with our own if we do not expose them to the user.

	// Recheck should be the default, but make sure.
	nodeCfg.Mempool.Recheck = true

	// Translate the entire config.
	userChainCfg := cfg.ChainConfig

	if userChainCfg.Moniker != "" {
		nodeCfg.Moniker = userChainCfg.Moniker
	}

	nodeCfg.FilterPeers = true
	nodeCfg.RPC.ListenAddress = cleanListenAddr(userChainCfg.RPC.ListenAddress,
		portFromURL(nodeCfg.RPC.ListenAddress))
	// NOTE: we would add new config settings to configure CometBFT's RPC server
	// to use TLS, if that is something people want:
	// nodeCfg.RPC.TLSCertFile = cfg.AppConfig.ConsensusTLSCertFile
	// nodeCfg.RPC.TLSKeyFile = cfg.AppConfig.ConsensusAdminTLSKeyFile
	nodeCfg.RPC.TimeoutBroadcastTxCommit = time.Duration(userChainCfg.RPC.BroadcastTxTimeout)

	nodeCfg.P2P.ListenAddress = cleanListenAddr(userChainCfg.P2P.ListenAddress,
		portFromURL(nodeCfg.P2P.ListenAddress))
	nodeCfg.P2P.ExternalAddress = userChainCfg.P2P.ExternalAddress
	nodeCfg.P2P.PersistentPeers = userChainCfg.P2P.PersistentPeers
	nodeCfg.P2P.AddrBookStrict = userChainCfg.P2P.AddrBookStrict
	nodeCfg.P2P.MaxNumInboundPeers = userChainCfg.P2P.MaxNumInboundPeers
	nodeCfg.P2P.MaxNumOutboundPeers = userChainCfg.P2P.MaxNumOutboundPeers
	nodeCfg.P2P.UnconditionalPeerIDs = userChainCfg.P2P.UnconditionalPeerIDs
	nodeCfg.P2P.PexReactor = userChainCfg.P2P.PexReactor
	nodeCfg.P2P.AllowDuplicateIP = cfg.ChainConfig.P2P.AllowDuplicateIP
	nodeCfg.P2P.HandshakeTimeout = time.Duration(userChainCfg.P2P.HandshakeTimeout)
	nodeCfg.P2P.DialTimeout = time.Duration(userChainCfg.P2P.DialTimeout)
	nodeCfg.P2P.SeedMode = userChainCfg.P2P.SeedMode
	nodeCfg.P2P.Seeds = userChainCfg.P2P.Seeds

	nodeCfg.Mempool.Size = userChainCfg.Mempool.Size
	nodeCfg.Mempool.CacheSize = userChainCfg.Mempool.CacheSize
	nodeCfg.Mempool.MaxTxBytes = userChainCfg.Mempool.MaxTxBytes
	nodeCfg.Mempool.MaxTxsBytes = int64(userChainCfg.Mempool.MaxTxsBytes)

	nodeCfg.Consensus.TimeoutPropose = time.Duration(userChainCfg.Consensus.TimeoutPropose)
	nodeCfg.Consensus.TimeoutPrevote = time.Duration(userChainCfg.Consensus.TimeoutPrevote)
	nodeCfg.Consensus.TimeoutPrecommit = time.Duration(userChainCfg.Consensus.TimeoutPrecommit)
	nodeCfg.Consensus.TimeoutCommit = time.Duration(userChainCfg.Consensus.TimeoutCommit)

	nodeCfg.StateSync.Enable = userChainCfg.StateSync.Enable
	nodeCfg.StateSync.RPCServers = strings.Split(userChainCfg.StateSync.RPCServers, ",")
	nodeCfg.StateSync.DiscoveryTime = time.Duration(userChainCfg.StateSync.DiscoveryTime)
	nodeCfg.StateSync.ChunkRequestTimeout = time.Duration(userChainCfg.StateSync.ChunkRequestTimeout)

	nodeCfg.Instrumentation = &cmtCfg.InstrumentationConfig{
		Prometheus:           cfg.Instrumentation.Prometheus,
		PrometheusListenAddr: cfg.Instrumentation.PromListenAddr,
		MaxOpenConnections:   cfg.Instrumentation.MaxConnections,
		Namespace:            "cometbft",
	}

	// Light client verification
	nodeCfg.StateSync.TrustPeriod = time.Duration(userChainCfg.StateSync.TrustPeriod)

	// Standardize the paths.
	nodeCfg.DBPath = cometbft.DataDir // i.e. "data", we do not allow users to change

	chainRoot := kconfig.ABCIDir(cfg.RootDir)
	nodeCfg.SetRoot(chainRoot)
	// NOTE: The Genesis field is the one in cometbft's GenesisDoc, which is
	// different from kwild's, which contains more fields (and not string
	// int64). The documented genesis.json in kwild's root directory is:
	//   filepath.Join(cfg.RootDir, cometbft.GenesisJSONName)
	// This file is only used to reflect the in-memory genesis config provided
	// to cometbft via a GenesisDocProvider. It it is not used by cometbft.
	nodeCfg.Genesis = filepath.Join(chainRoot, "config", cometbft.GenesisJSONName)
	nodeCfg.P2P.AddrBook = cometbft.AddrBookPath(chainRoot)
	// For the same reasons described for the genesis.json path above, clear the
	// node and validator file fields since they are provided in-memory.
	nodeCfg.PrivValidatorKey = ""
	nodeCfg.PrivValidatorState = ""
	nodeCfg.NodeKey = ""

	return nodeCfg
}

// extractGenesisDoc is used by cometbft while initializing the node to extract
// the genesis configuration. Note that cometbft's GenesisDoc is a subset of
// kwild's genesis file. The app version set in AppVersion is used to supply
// cometbft with the application protocol version, which is determined by the
// app code rather than a configurable value in our genesis config.
func extractGenesisDoc(g *chain.GenesisConfig) (*cmttypes.GenesisDoc, error) {
	// BaseConsensusParms => cometbft's ConsensusParms
	consensusParams := cometbft.ExtractConsensusParams(&g.ConsensusParams.BaseConsensusParams, AppVersion)

	genDoc := &cmttypes.GenesisDoc{
		ChainID:         g.ChainID,
		GenesisTime:     g.GenesisTime,
		InitialHeight:   g.InitialHeight,
		AppHash:         g.DataAppHash,
		ConsensusParams: consensusParams,
	}

	for _, v := range g.Validators {
		if len(v.PubKey) != cmtEd.PubKeySize {
			return nil, fmt.Errorf("pubkey is incorrect size: %v", v.PubKey.String())
		}
		pubKey := cmtEd.PubKey(v.PubKey)
		genDoc.Validators = append(genDoc.Validators, cmttypes.GenesisValidator{
			Address: pubKey.Address(),
			PubKey:  pubKey,
			Power:   v.Power,
			Name:    v.Name,
		})
	}
	return genDoc, nil
}
