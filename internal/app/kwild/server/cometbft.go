package server

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/pkg/abci/cometbft"

	cmtCfg "github.com/cometbft/cometbft/config"
	cmtEd "github.com/cometbft/cometbft/crypto/ed25519"
)

// newCometConfig creates a new CometBFT config for use with NewCometBftNode.
// This applies The operator's settings from the Kwil config as well as applying
// some overrides to the defaults for Kwil.
//
// NOTE: this is somewhat error prone, so care must be taken to update this
// function when fields are added to KwildConfig.ChainCfg.
func newCometConfig(cfg *config.KwildConfig) *cmtCfg.Config {
	// Begin with CometBFT's default chain config.
	nodeCfg := cmtCfg.DefaultConfig()

	// Override defaults with our own if we do not expose them to the user.

	// As all we are validating are tx signatures, no need to go through
	// Validation again. To be set to true when we have validations based on
	// gas, nonces, account balance, etc.
	nodeCfg.Mempool.Recheck = false

	// Translate the entire config.
	userChainCfg := cfg.ChainCfg

	if userChainCfg.Moniker != "" {
		nodeCfg.Moniker = userChainCfg.Moniker
	}

	nodeCfg.RPC.ListenAddress = userChainCfg.RPC.ListenAddress
	nodeCfg.RPC.GRPCListenAddress = userChainCfg.RPC.GRPCListenAddress
	nodeCfg.RPC.TimeoutBroadcastTxCommit = userChainCfg.RPC.TimeoutBroadcastTxCommit
	nodeCfg.RPC.TLSCertFile = cfg.AppCfg.TLSCertFile
	nodeCfg.RPC.TLSKeyFile = cfg.AppCfg.TLSKeyFile

	nodeCfg.P2P.ListenAddress = userChainCfg.P2P.ListenAddress
	nodeCfg.P2P.ExternalAddress = userChainCfg.P2P.ExternalAddress
	nodeCfg.P2P.PersistentPeers = userChainCfg.P2P.PersistentPeers
	nodeCfg.P2P.AddrBookStrict = userChainCfg.P2P.AddrBookStrict
	nodeCfg.P2P.MaxNumInboundPeers = userChainCfg.P2P.MaxNumInboundPeers
	nodeCfg.P2P.MaxNumOutboundPeers = userChainCfg.P2P.MaxNumOutboundPeers
	nodeCfg.P2P.UnconditionalPeerIDs = userChainCfg.P2P.UnconditionalPeerIDs
	nodeCfg.P2P.PexReactor = userChainCfg.P2P.PexReactor
	nodeCfg.P2P.AllowDuplicateIP = cfg.ChainCfg.P2P.AllowDuplicateIP
	nodeCfg.P2P.HandshakeTimeout = userChainCfg.P2P.HandshakeTimeout
	nodeCfg.P2P.DialTimeout = userChainCfg.P2P.DialTimeout

	nodeCfg.Mempool.Size = userChainCfg.Mempool.Size
	nodeCfg.Mempool.CacheSize = userChainCfg.Mempool.CacheSize

	nodeCfg.Consensus.TimeoutPropose = userChainCfg.Consensus.TimeoutPropose
	nodeCfg.Consensus.TimeoutPrevote = userChainCfg.Consensus.TimeoutPrevote
	nodeCfg.Consensus.TimeoutPrecommit = userChainCfg.Consensus.TimeoutPrecommit
	nodeCfg.Consensus.TimeoutCommit = userChainCfg.Consensus.TimeoutCommit

	nodeCfg.StateSync.Enable = userChainCfg.StateSync.Enable
	nodeCfg.StateSync.TempDir = userChainCfg.StateSync.TempDir
	nodeCfg.StateSync.RPCServers = userChainCfg.StateSync.RPCServers
	nodeCfg.StateSync.DiscoveryTime = userChainCfg.StateSync.DiscoveryTime
	nodeCfg.StateSync.ChunkRequestTimeout = userChainCfg.StateSync.ChunkRequestTimeout

	// Standardize the paths.
	nodeCfg.DBPath = cometbft.DataDir // i.e. "data", we do not allow users to change

	chainRoot := filepath.Join(cfg.RootDir, abciDirName)
	nodeCfg.SetRoot(chainRoot)
	nodeCfg.Genesis = cometbft.GenesisPath(chainRoot)
	nodeCfg.P2P.AddrBook = cometbft.AddrBookPath(chainRoot)

	return nodeCfg
}

// loadGenesisAndPrivateKey generates private key and genesis file if not exist
//
//   - If genesis file exists but not private key file, it will generate private
//     key and start the node as a non-validator.
//   - Otherwise, the genesis file is generated based on the private key and
//     starts the node as a validator.
func loadGenesisAndPrivateKey(autoGen bool, privKeyPath, chainRootDir string) (privKey cmtEd.PrivKey, newKey, newGenesis bool, err error) {
	// Get private key:
	//  - if private key file exists, load it.
	//  - else if in autogen mode, generate private key and write to file.
	//  - else fail
	if fileExists(privKeyPath) {
		privKeyHexB, err0 := os.ReadFile(privKeyPath)
		if err0 != nil {
			err = fmt.Errorf("error reading private key file: %v", err0)
			return
		}
		privKeyHex := string(bytes.TrimSpace(privKeyHexB))
		privB, err0 := hex.DecodeString(privKeyHex)
		if err0 != nil {
			err = fmt.Errorf("error decoding private key: %v", err0)
			return
		}
		privKey = cmtEd.PrivKey(privB)
	} else if autoGen {
		privKey, err = cometbft.GeneratePrivateKeyFile(privKeyPath)
		if err != nil {
			err = fmt.Errorf("error creating private key file: %v", err)
			return
		}
		newKey = true
	} else {
		return nil, false, false, fmt.Errorf("private key not found")
	}

	abciCfgDir := filepath.Join(chainRootDir, cometbft.ConfigDir)
	genFile := filepath.Join(abciCfgDir, cometbft.GenesisJSONName) // i.e. <root>/abci/config/genesis.json
	if !fileExists(genFile) {
		if !autoGen {
			err = fmt.Errorf("genesis file not found: %s", genFile)
			return
		}

		if err = os.MkdirAll(abciCfgDir, 0755); err != nil {
			err = fmt.Errorf("error creating abci config dir %s: %v", abciCfgDir, err)
			return
		}

		err = cometbft.GenerateGenesisFile(genFile, []cmtEd.PrivKey{privKey}, "kwil-chain-")
		if err != nil {
			err = fmt.Errorf("unable to write genesis file %s: %v", genFile, err)
			return
		}
		newGenesis = true
	}

	return
}
