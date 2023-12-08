package config

import "github.com/spf13/pflag"

// AddConfigFlags adds all flags from KwildConfig to the given flagSet
func AddConfigFlags(flagSet *pflag.FlagSet, cfg *KwildConfig) {
	flagSet.StringVarP(&cfg.RootDir, "root-dir", "r", "~/.kwild", "kwild root directory for config and data")

	// logging
	flagSet.StringVarP(&cfg.Logging.Level, "log.level", "l", cfg.Logging.Level, "kwild log level")
	flagSet.StringVar(&cfg.Logging.Format, "log.format", cfg.Logging.Format, "kwild log format")
	flagSet.StringVar(&cfg.Logging.TimeEncoding, "log.time-format", cfg.Logging.TimeEncoding, "kwild time log format")
	flagSet.StringSliceVar(&cfg.Logging.OutputPaths, "log.output-paths", cfg.Logging.OutputPaths, "kwild log output paths")

	// General APP flags:
	flagSet.StringVar(&cfg.AppCfg.PrivateKeyPath, "app.private-key-path", cfg.AppCfg.PrivateKeyPath, "Path to the node private key file")
	flagSet.StringVar(&cfg.AppCfg.GrpcListenAddress, "app.grpc-listen-addr", cfg.AppCfg.GrpcListenAddress, "kwild gRPC listen address")
	flagSet.StringVar(&cfg.AppCfg.HTTPListenAddress, "app.http-listen-addr", cfg.AppCfg.HTTPListenAddress, "kwild HTTP listen address")
	flagSet.StringVar(&cfg.AppCfg.AdminListenAddress, "app.admin-listen-addr", cfg.AppCfg.AdminListenAddress, "kwild admin listen address (unix or tcp)")
	flagSet.StringVar(&cfg.AppCfg.SqliteFilePath, "app.sqlite-file-path", cfg.AppCfg.SqliteFilePath, "kwild sqlite file path")
	flagSet.StringVar(&cfg.AppCfg.TLSCertFile, "app.tls-cert-file", cfg.AppCfg.TLSCertFile, "TLS certificate file path for RPC Server")
	flagSet.StringVar(&cfg.AppCfg.TLSKeyFile, "app.tls-key-file", cfg.AppCfg.TLSKeyFile, "TLS key file path for RPC Server")
	flagSet.BoolVar(&cfg.AppCfg.EnableRPCTLS, "app.rpctls", cfg.AppCfg.EnableRPCTLS, "Use TLS on the user gRPC server")
	flagSet.StringVar(&cfg.AppCfg.Hostname, "app.hostname", cfg.AppCfg.Hostname, "kwild Server hostname")

	flagSet.StringVar(&cfg.AppCfg.ProfileMode, "app.profile-mode", cfg.AppCfg.ProfileMode, "kwild profile mode (http, cpu, mem, mutex, or block)")
	flagSet.StringVar(&cfg.AppCfg.ProfileFile, "app.profile-file", cfg.AppCfg.ProfileFile, "kwild profile output file path (e.g. cpu.pprof)")

	// Extension endpoints flags
	flagSet.StringSliceVar(&cfg.AppCfg.ExtensionEndpoints, "app.extension-endpoints", cfg.AppCfg.ExtensionEndpoints, "kwild extension endpoints")

	// TODO: Snapshots are not supported yet
	// // Snapshot Config flags
	// flagSet.BoolVar(&cfg.AppCfg.SnapshotConfig.Enabled, "app.snapshots.enabled", cfg.AppCfg.SnapshotConfig.Enabled, "Enable snapshots")
	// flagSet.Uint64Var(&cfg.AppCfg.SnapshotConfig.RecurringHeight, "app.snapshots.recurring-height", cfg.AppCfg.SnapshotConfig.RecurringHeight, "Recurring snapshot height")
	// flagSet.Uint64Var(&cfg.AppCfg.SnapshotConfig.MaxSnapshots, "app.snapshots.max-snapshots", cfg.AppCfg.SnapshotConfig.MaxSnapshots, "Maximum snapshots")
	// flagSet.StringVar(&cfg.AppCfg.SnapshotConfig.SnapshotDir, "app.snapshots.snapshot-dir", cfg.AppCfg.SnapshotConfig.SnapshotDir, "Snapshot directory path")

	// Basic Chain Config flags
	flagSet.StringVar(&cfg.ChainCfg.Moniker, "chain.moniker", cfg.ChainCfg.Moniker, "Node moniker")
	// flagSet.StringVar(&cfg.ChainCfg.DBPath, "chain.db-dir", cfg.ChainCfg.DBPath, "Chain database directory path") // rm?

	// Chain RPC flags
	flagSet.StringVar(&cfg.ChainCfg.RPC.ListenAddress, "chain.rpc.listen-addr", cfg.ChainCfg.RPC.ListenAddress, "Chain RPC listen address")

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
	// TODO: Bring these flags back when we support state sync
	// flagSet.BoolVar(&cfg.ChainCfg.StateSync.Enable, "chain.state-sync.enable", cfg.ChainCfg.StateSync.Enable, "Chain state sync enable")
	// flagSet.StringVar(&cfg.ChainCfg.StateSync.TempDir, "chain.state-sync.temp-dir", cfg.ChainCfg.StateSync.TempDir, "Chain state sync temp dir")
	// flagSet.StringSliceVar(&cfg.ChainCfg.StateSync.RPCServers, "chain.state-sync.rpc-servers", cfg.ChainCfg.StateSync.RPCServers, "Chain state sync rpc servers")
	// flagSet.DurationVar(&cfg.ChainCfg.StateSync.DiscoveryTime, "chain.state-sync.discovery-time", cfg.ChainCfg.StateSync.DiscoveryTime, "Chain state sync discovery time")
	// flagSet.DurationVar(&cfg.ChainCfg.StateSync.ChunkRequestTimeout, "chain.state-sync.chunk-request-timeout", cfg.ChainCfg.StateSync.ChunkRequestTimeout, "Chain state sync chunk request timeout")

	// Block sync can be added later (when they have more version of it)
}
