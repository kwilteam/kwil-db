package config

import "github.com/spf13/pflag"

// AddConfigFlags adds all flags from KwildConfig to the given flagSet
func AddConfigFlags(flagSet *pflag.FlagSet, cfg *KwildConfig) {
	flagSet.StringVarP(&cfg.RootDir, "root-dir", "r", "~/.kwild", "kwild root directory for config and data")

	// logging
	flagSet.StringVarP(&cfg.Logging.Level, "log.level", "l", cfg.Logging.Level, "kwild log level")
	flagSet.StringVar(&cfg.Logging.RPCLevel, "log.rpc_level", cfg.Logging.RPCLevel, "user rpc server log level")
	flagSet.StringVar(&cfg.Logging.ConsensusLevel, "log.consensus_level", cfg.Logging.ConsensusLevel, "consensus (cometbft) log level")
	flagSet.StringVar(&cfg.Logging.DBLevel, "log.db_level", cfg.Logging.DBLevel, "database backend (postgres) log level")
	flagSet.StringVar(&cfg.Logging.Format, "log.format", cfg.Logging.Format, "kwild log format")
	flagSet.StringVar(&cfg.Logging.TimeEncoding, "log.time-format", cfg.Logging.TimeEncoding, "kwild time log format")
	flagSet.StringSliceVar(&cfg.Logging.OutputPaths, "log.output-paths", cfg.Logging.OutputPaths, "kwild log output paths")

	// General APP flags:
	flagSet.StringVar(&cfg.AppCfg.PrivateKeyPath, "app.private-key-path", cfg.AppCfg.PrivateKeyPath, "Path to the node private key file")
	flagSet.StringVar(&cfg.AppCfg.JSONRPCListenAddress, "app.jsonrpc-listen-addr", cfg.AppCfg.JSONRPCListenAddress, "kwild JSON-RPC listen address")
	flagSet.StringVar(&cfg.AppCfg.HTTPListenAddress, "app.http-listen-addr", cfg.AppCfg.HTTPListenAddress, "kwild HTTP listen address")
	flagSet.StringVar(&cfg.AppCfg.AdminListenAddress, "app.admin-listen-addr", cfg.AppCfg.AdminListenAddress, "kwild admin listen address (unix or tcp)")
	flagSet.StringVar(&cfg.AppCfg.AdminRPCPass, "app.admin-pass", cfg.AppCfg.AdminRPCPass, "password for the kwil admin service (may be empty)")
	flagSet.BoolVar(&cfg.AppCfg.NoTLS, "app.admin-notls", cfg.AppCfg.NoTLS, "do not enable TLS on admin server (automatically disabled for unix socket or loopback listen addresses)")
	flagSet.StringVar(&cfg.AppCfg.TLSCertFile, "app.tls-cert-file", cfg.AppCfg.TLSCertFile, "TLS certificate file path for the admin and consensus RPC server (optional)")
	flagSet.StringVar(&cfg.AppCfg.TLSKeyFile, "app.tls-key-file", cfg.AppCfg.TLSKeyFile, "TLS key file path for the admin and consensus RPC servers (optional)")
	flagSet.StringVar(&cfg.AppCfg.Hostname, "app.hostname", cfg.AppCfg.Hostname, "kwild Server hostname")

	flagSet.StringVar(&cfg.AppCfg.DBHost, "app.pg-db-host", cfg.AppCfg.DBHost, "PostgreSQL host address (no port)")
	flagSet.StringVar(&cfg.AppCfg.DBPort, "app.pg-db-port", cfg.AppCfg.DBPort, "PostgreSQL port")
	flagSet.StringVar(&cfg.AppCfg.DBUser, "app.pg-db-user", cfg.AppCfg.DBUser, "PostgreSQL user name")
	flagSet.StringVar(&cfg.AppCfg.DBPass, "app.pg-db-pass", cfg.AppCfg.DBPass, "PostgreSQL password name")
	flagSet.StringVar(&cfg.AppCfg.DBName, "app.pg-db-name", cfg.AppCfg.DBName, "PostgreSQL database name")

	flagSet.StringVar(&cfg.AppCfg.ProfileMode, "app.profile-mode", cfg.AppCfg.ProfileMode, "kwild profile mode (http, cpu, mem, mutex, or block)")
	flagSet.StringVar(&cfg.AppCfg.ProfileFile, "app.profile-file", cfg.AppCfg.ProfileFile, "kwild profile output file path (e.g. cpu.pprof)")

	flagSet.Var(&cfg.AppCfg.RPCTimeout, "app.rpc-timeout", "timeout for RPC requests (through reading the request, handling the request, and sending the response)")
	flagSet.IntVar(&cfg.AppCfg.RPCMaxReqSize, "app.rpc-req-limit", cfg.AppCfg.RPCMaxReqSize, "RPC request size limit")
	flagSet.Var(&cfg.AppCfg.ReadTxTimeout, "app.db-read-timeout", "timeout for database reads initiated by RPC requests")

	// Extension endpoints flags
	flagSet.StringSliceVar(&cfg.AppCfg.ExtensionEndpoints, "app.extension-endpoints", cfg.AppCfg.ExtensionEndpoints, "kwild extension endpoints")

	// Snapshot Config flags
	flagSet.BoolVar(&cfg.AppCfg.Snapshots.Enabled, "app.snapshots.enabled", cfg.AppCfg.Snapshots.Enabled, "Enable snapshots")
	flagSet.Uint64Var(&cfg.AppCfg.Snapshots.RecurringHeight, "app.snapshots.recurring-height", cfg.AppCfg.Snapshots.RecurringHeight, "Recurring heights to create snapshots")
	flagSet.Uint64Var(&cfg.AppCfg.Snapshots.MaxSnapshots, "app.snapshots.max-snapshots", cfg.AppCfg.Snapshots.MaxSnapshots, "Maximum snapshots to store on disk. Default is 3. If max snapshots is reached, the oldest snapshot is deleted.")
	flagSet.StringVar(&cfg.AppCfg.Snapshots.SnapshotDir, "app.snapshots.snapshot-dir", cfg.AppCfg.Snapshots.SnapshotDir, "Snapshot directory path")

	flagSet.StringVar(&cfg.AppCfg.GenesisState, "app.genesis-state", cfg.AppCfg.GenesisState, "Path to the genesis state file")

	// Basic Chain Config flags
	flagSet.StringVar(&cfg.ChainCfg.Moniker, "chain.moniker", cfg.ChainCfg.Moniker, "Node moniker")

	// Chain RPC flags
	flagSet.StringVar(&cfg.ChainCfg.RPC.ListenAddress, "chain.rpc.listen-addr", cfg.ChainCfg.RPC.ListenAddress, "Chain RPC listen address")
	flagSet.Var(&cfg.ChainCfg.RPC.BroadcastTxTimeout, "chain.rpc.broadcast-tx-timeout", "Chain RPC broadcast transaction timeout")

	// Chain P2P flags
	flagSet.StringVar(&cfg.ChainCfg.P2P.ListenAddress, "chain.p2p.listen-addr", cfg.ChainCfg.P2P.ListenAddress, "Chain P2P listen address")
	flagSet.StringVar(&cfg.ChainCfg.P2P.ExternalAddress, "chain.p2p.external-address", cfg.ChainCfg.P2P.ExternalAddress, "Chain P2P external address to advertise")
	flagSet.StringVar(&cfg.ChainCfg.P2P.PersistentPeers, "chain.p2p.persistent-peers", cfg.ChainCfg.P2P.PersistentPeers, "Chain P2P persistent peers")
	flagSet.BoolVar(&cfg.ChainCfg.P2P.AddrBookStrict, "chain.p2p.addr-book-strict", cfg.ChainCfg.P2P.AddrBookStrict, "Chain P2P address book strict")
	flagSet.StringVar(&cfg.ChainCfg.P2P.UnconditionalPeerIDs, "chain.p2p.unconditional-peer-ids", cfg.ChainCfg.P2P.UnconditionalPeerIDs, "Chain P2P unconditional peer IDs")
	flagSet.IntVar(&cfg.ChainCfg.P2P.MaxNumInboundPeers, "chain.p2p.max-num-inbound-peers", cfg.ChainCfg.P2P.MaxNumInboundPeers, "Chain P2P maximum number of inbound peers")
	flagSet.IntVar(&cfg.ChainCfg.P2P.MaxNumOutboundPeers, "chain.p2p.max-num-outbound-peers", cfg.ChainCfg.P2P.MaxNumOutboundPeers, "Chain P2P maximum number of outbound peers")
	flagSet.BoolVar(&cfg.ChainCfg.P2P.AllowDuplicateIP, "chain.p2p.allow-duplicate-ip", cfg.ChainCfg.P2P.AllowDuplicateIP, "Chain P2P allow multiple peers with the same IP address")
	flagSet.BoolVar(&cfg.ChainCfg.P2P.PexReactor, "chain.p2p.pex", cfg.ChainCfg.P2P.PexReactor, "Enables peer information exchange")
	flagSet.StringVar(&cfg.ChainCfg.P2P.Seeds, "chain.p2p.seeds", cfg.ChainCfg.P2P.Seeds, "Seed nodes for obtaining peer addresses, if address book is empty")
	flagSet.BoolVar(&cfg.ChainCfg.P2P.SeedMode, "chain.p2p.seed-mode", cfg.ChainCfg.P2P.SeedMode, `Run kwild in a special "seed" mode where it crawls the network for peer addresses,
sharing them with incoming peers before immediately disconnecting. It is recommended
to instead run a dedicated seeder like https://github.com/kwilteam/cometseed.`)

	// Chain Mempool flags
	flagSet.IntVar(&cfg.ChainCfg.Mempool.Size, "chain.mempool.size", cfg.ChainCfg.Mempool.Size, "Chain mempool size")
	flagSet.IntVar(&cfg.ChainCfg.Mempool.CacheSize, "chain.mempool.cache-size", cfg.ChainCfg.Mempool.CacheSize, "Chain mempool cache size")
	flagSet.IntVar(&cfg.ChainCfg.Mempool.MaxTxBytes, "chain.mempool.max-tx-bytes", cfg.ChainCfg.Mempool.MaxTxBytes, "chain mempool maximum single transaction size in bytes")
	flagSet.IntVar(&cfg.ChainCfg.Mempool.MaxTxsBytes, "chain.mempool.max-txs-bytes", cfg.ChainCfg.Mempool.MaxTxsBytes, "chain mempool maximum total transactions in bytes")

	// Chain Consensus flags
	flagSet.Var(&cfg.ChainCfg.Consensus.TimeoutPropose, "chain.consensus.timeout-propose", "Chain consensus timeout propose")
	flagSet.Var(&cfg.ChainCfg.Consensus.TimeoutPrevote, "chain.consensus.timeout-prevote", "Chain consensus timeout prevote")
	flagSet.Var(&cfg.ChainCfg.Consensus.TimeoutPrecommit, "chain.consensus.timeout-precommit", "Chain consensus timeout precommit")
	flagSet.Var(&cfg.ChainCfg.Consensus.TimeoutCommit, "chain.consensus.timeout-commit", "Chain consensus timeout commit")

	// State Sync flags
	flagSet.BoolVar(&cfg.ChainCfg.StateSync.Enable, "chain.statesync.enable", cfg.ChainCfg.StateSync.Enable, "Chain state sync enable")
	flagSet.StringVar(&cfg.ChainCfg.StateSync.SnapshotDir, "chain.statesync.snapshot-dir", cfg.ChainCfg.StateSync.SnapshotDir, "Chain state sync snapshot directory")
	flagSet.StringVar(&cfg.ChainCfg.StateSync.RPCServers, "chain.statesync.rpc-servers", cfg.ChainCfg.StateSync.RPCServers, "Chain state sync rpc servers")
	flagSet.Var(&cfg.ChainCfg.StateSync.DiscoveryTime, "chain.statesync.discovery-time", "Chain state sync discovery time")
	flagSet.Var(&cfg.ChainCfg.StateSync.ChunkRequestTimeout, "chain.statesync.chunk-request-timeout", "Chain state sync chunk request timeout")
	flagSet.Var(&cfg.ChainCfg.StateSync.TrustPeriod, "chain.statesync.trust-period", "Duration of time for which the snapshots are trusted")
}
