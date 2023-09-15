package main

import (
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"

	"github.com/spf13/pflag"
)

func addKwildFlags(flagSet *pflag.FlagSet, cfg *config.KwildConfig) {
	flagSet.BoolVarP(&cfg.AutoGen, "autogen", "a", false,
		"auto generate private key and genesis file if not exist")
	flagSet.StringVarP(&cfg.RootDir, "root_dir", "r", "~/.kwild", "kwild root directory for config and data")

	// logging
	flagSet.StringVarP(&cfg.Logging.Level, "log.level", "l", cfg.Logging.Level, "Kwild app log level")
	flagSet.StringVar(&cfg.Logging.Format, "log.format", cfg.Logging.Format, "Kwild app log format")
	flagSet.StringVar(&cfg.Logging.TimeEncoding, "log.time_format", cfg.Logging.TimeEncoding, "Kwild app time log format")
	flagSet.StringSliceVar(&cfg.Logging.OutputPaths, "log.output_paths", cfg.Logging.OutputPaths, "Kwild app log output paths")

	// General APP flags:
	flagSet.StringVar(&cfg.AppCfg.PrivateKeyPath, "app.private_key_path", cfg.AppCfg.PrivateKeyPath, "Path to private key file")
	flagSet.StringVar(&cfg.AppCfg.GrpcListenAddress, "app.grpc_listen_addr", cfg.AppCfg.GrpcListenAddress, "Kwild app gRPC listen address")
	flagSet.StringVar(&cfg.AppCfg.HTTPListenAddress, "app.http_listen_addr", cfg.AppCfg.HTTPListenAddress, "Kwild app HTTP listen address")
	flagSet.StringVar(&cfg.AppCfg.AdminListenAddress, "app.admin_listen_addr", cfg.AppCfg.AdminListenAddress, "Kwild app gRPC listen address")
	flagSet.StringVar(&cfg.AppCfg.SqliteFilePath, "app.sqlite_file_path", cfg.AppCfg.SqliteFilePath, "Kwild app sqlite file path")
	flagSet.BoolVar(&cfg.AppCfg.WithoutGasCosts, "app.without_gas_costs", cfg.AppCfg.WithoutGasCosts, "Kwild app without gas costs")
	flagSet.BoolVar(&cfg.AppCfg.WithoutNonces, "app.without_nonces", cfg.AppCfg.WithoutNonces, "Kwild app without nonces")
	flagSet.StringVar(&cfg.AppCfg.TLSCertFile, "app.tls_cert_file", cfg.AppCfg.TLSCertFile, "TLS certificate file path for RPC Server")
	flagSet.StringVar(&cfg.AppCfg.TLSKeyFile, "app.tls_key_file", cfg.AppCfg.TLSKeyFile, "TLS key file path for RPC Server")
	flagSet.BoolVar(&cfg.AppCfg.EnableRPCTLS, "app.rpctls", cfg.AppCfg.EnableRPCTLS, "Kwild app without nonces")
	flagSet.StringVar(&cfg.AppCfg.Hostname, "app.hostname", cfg.AppCfg.Hostname, "Kwild Server hostname")

	// Extension endpoints flags
	flagSet.StringSliceVar(&cfg.AppCfg.ExtensionEndpoints, "app.extension_endpoints", cfg.AppCfg.ExtensionEndpoints, "Kwild app extension endpoints")

	// Snapshot Config flags
	flagSet.BoolVar(&cfg.AppCfg.SnapshotConfig.Enabled, "app.snapshots.enabled", cfg.AppCfg.SnapshotConfig.Enabled, "Enable snapshots")
	flagSet.Uint64Var(&cfg.AppCfg.SnapshotConfig.RecurringHeight, "app.snapshots.recurring_height", cfg.AppCfg.SnapshotConfig.RecurringHeight, "Recurring snapshot height")
	flagSet.Uint64Var(&cfg.AppCfg.SnapshotConfig.MaxSnapshots, "app.snapshots.max_snapshots", cfg.AppCfg.SnapshotConfig.MaxSnapshots, "Maximum snapshots")
	flagSet.StringVar(&cfg.AppCfg.SnapshotConfig.SnapshotDir, "app.snapshots.snapshot_dir", cfg.AppCfg.SnapshotConfig.SnapshotDir, "Snapshot directory path")

	// Basic Chain Config flags
	flagSet.StringVar(&cfg.ChainCfg.Moniker, "chain.moniker", cfg.ChainCfg.Moniker, "Node moniker")
	// flagSet.StringVar(&cfg.ChainCfg.DBPath, "chain.db_dir", cfg.ChainCfg.DBPath, "Chain database directory path") // rm?

	// Chain RPC flags
	flagSet.StringVar(&cfg.ChainCfg.RPC.ListenAddress, "chain.rpc.laddr", cfg.ChainCfg.RPC.ListenAddress, "Chain RPC listen address")
	flagSet.DurationVar(&cfg.ChainCfg.RPC.TimeoutBroadcastTxCommit, "chain.rpc.timeout_broadcast_tx_commit", cfg.ChainCfg.RPC.TimeoutBroadcastTxCommit, "chain timeout broadcast tx commit")

	// Chain P2P flags
	flagSet.StringVar(&cfg.ChainCfg.P2P.ListenAddress, "chain.p2p.laddr", cfg.ChainCfg.P2P.ListenAddress, "chain P2P listen address")
	flagSet.StringVar(&cfg.ChainCfg.P2P.ExternalAddress, "chain.p2p.external_address", cfg.ChainCfg.P2P.ExternalAddress, "chain P2P external address")
	flagSet.StringVar(&cfg.ChainCfg.P2P.PersistentPeers, "chain.p2p.persistent_peers", cfg.ChainCfg.P2P.PersistentPeers, "chain P2P persistent peers")
	flagSet.BoolVar(&cfg.ChainCfg.P2P.AddrBookStrict, "chain.p2p.addr_book_strict", cfg.ChainCfg.P2P.AddrBookStrict, "chain P2P address book strict")
	flagSet.StringVar(&cfg.ChainCfg.P2P.UnconditionalPeerIDs, "chain.p2p.unconditional_peer_ids", cfg.ChainCfg.P2P.UnconditionalPeerIDs, "chain P2P unconditional peer IDs")
	flagSet.IntVar(&cfg.ChainCfg.P2P.MaxNumInboundPeers, "chain.p2p.max_num_inbound_peers", cfg.ChainCfg.P2P.MaxNumInboundPeers, "chain P2P maximum number of inbound peers")
	flagSet.IntVar(&cfg.ChainCfg.P2P.MaxNumOutboundPeers, "chain.p2p.max_num_outbound_peers", cfg.ChainCfg.P2P.MaxNumOutboundPeers, "chain P2P maximum number of outbound peers")
	flagSet.BoolVar(&cfg.ChainCfg.P2P.AllowDuplicateIP, "chain.p2p.allow_duplicate_ip", cfg.ChainCfg.P2P.AllowDuplicateIP, "chain P2P allow duplicate IP")

	// Chain Mempool flags
	flagSet.IntVar(&cfg.ChainCfg.Mempool.Size, "chain.mempool.size", cfg.ChainCfg.Mempool.Size, "chain mempool size")
	flagSet.IntVar(&cfg.ChainCfg.Mempool.CacheSize, "chain.mempool.cache_size", cfg.ChainCfg.Mempool.CacheSize, "chain mempool cache size")
	// flagSet.Int64Var(&cfg.ChainCfg.Mempool.MaxTxsBytes, "chain.mempool.max_txs_bytes", cfg.ChainCfg.Mempool.MaxTxsBytes, "chain mempool maximum transactions bytes")

	// Chain Consensus flags
	flagSet.DurationVar(&cfg.ChainCfg.Consensus.TimeoutPropose, "chain.consensus.timeout_propose", cfg.ChainCfg.Consensus.TimeoutPropose, "chain consensus timeout propose")
	flagSet.DurationVar(&cfg.ChainCfg.Consensus.TimeoutPrevote, "chain.consensus.timeout_prevote", cfg.ChainCfg.Consensus.TimeoutPrevote, "chain consensus timeout prevote")
	flagSet.DurationVar(&cfg.ChainCfg.Consensus.TimeoutPrecommit, "chain.consensus.timeout_precommit", cfg.ChainCfg.Consensus.TimeoutPrecommit, "chain consensus timeout precommit")
	flagSet.DurationVar(&cfg.ChainCfg.Consensus.TimeoutCommit, "chain.consensus.timeout_commit", cfg.ChainCfg.Consensus.TimeoutCommit, "chain consensus timeout commit")

	// State Sync flags
	flagSet.BoolVar(&cfg.ChainCfg.StateSync.Enable, "chain.state_sync.enable", cfg.ChainCfg.StateSync.Enable, "chain state sync enable")
	flagSet.StringVar(&cfg.ChainCfg.StateSync.TempDir, "chain.state_sync.temp_dir", cfg.ChainCfg.StateSync.TempDir, "chain state sync temp dir")
	flagSet.StringSliceVar(&cfg.ChainCfg.StateSync.RPCServers, "chain.state_sync.rpc_servers", cfg.ChainCfg.StateSync.RPCServers, "chain state sync rpc servers")
	flagSet.DurationVar(&cfg.ChainCfg.StateSync.DiscoveryTime, "chain.state_sync.discovery_time", cfg.ChainCfg.StateSync.DiscoveryTime, "chain state sync discovery time")
	flagSet.DurationVar(&cfg.ChainCfg.StateSync.ChunkRequestTimeout, "chain.state_sync.chunk_request_timeout", cfg.ChainCfg.StateSync.ChunkRequestTimeout, "chain state sync chunk request timeout")

	// Block sync can be added later (when they have more version of it)
}
