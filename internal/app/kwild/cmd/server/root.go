package server

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	// shorthand for chain client service

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/internal/app/kwild/server"
	"github.com/spf13/cobra"
)

func NewStartCmd(cfg *config.KwildConfig) *cobra.Command {
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "kwil grpc server",
		Long:  "Starts node with Kwild and CometBFT services",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cfg.AppCfg.PrivateKey == "" {
				return errors.New("private key is not set")
			}

			if cfg.RootDir == "" {
				return errors.New("kwild home directory not set")
			}

			signalChan := make(chan os.Signal, 1)
			signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
			ctx, cancel := context.WithCancel(cmd.Context())

			go func() {
				<-signalChan
				cancel()
			}()

			svr, err := server.BuildKwildServer(ctx, cfg)
			if err != nil {
				return err
			}

			return svr.Start(ctx)
		},
	}

	AddKwildFlags(startCmd, cfg)
	return startCmd
}

func AddKwildFlags(cmd *cobra.Command, cfg *config.KwildConfig) {
	// General APP flags:
	cmd.Flags().StringVar(&cfg.RootDir, "home", cfg.RootDir, "Kwild home directory to store blockchain, kwildb and other data")

	cmd.Flags().StringVar(&cfg.AppCfg.PrivateKey, "app.private_key", cfg.AppCfg.PrivateKey, "Kwild app's private key")

	cmd.Flags().StringVar(&cfg.AppCfg.GrpcListenAddress, "app.grpc_listen_addr", cfg.AppCfg.GrpcListenAddress, "Kwild app gRPC listen address")

	cmd.Flags().StringVar(&cfg.AppCfg.HttpListenAddress, "app.http_listen_addr", cfg.AppCfg.HttpListenAddress, "Kwild app HTTP listen address")

	cmd.Flags().StringVar(&cfg.AppCfg.SqliteFilePath, "sqlite_file_path", cfg.AppCfg.SqliteFilePath, "Kwild app sqlite file path")

	cmd.Flags().BoolVar(&cfg.AppCfg.WithoutGasCosts, "app.without_gas_costs", cfg.AppCfg.WithoutGasCosts, "Kwild app without gas costs")

	cmd.Flags().BoolVar(&cfg.AppCfg.WithoutNonces, "app.without_nonces", cfg.AppCfg.WithoutNonces, "Kwild app without nonces")

	cmd.Flags().StringVar(&cfg.AppCfg.TLSCertFile, "app.tls_cert_file", cfg.AppCfg.TLSCertFile, "TLS certificate file path for RPC Server")

	cmd.Flags().StringVar(&cfg.AppCfg.TLSKeyFile, "app.tls_key_file", cfg.AppCfg.TLSKeyFile, "TLS key file path for RPC Server")

	cmd.Flags().StringVar(&cfg.AppCfg.Hostname, "app.hostname", cfg.AppCfg.Hostname, "Kwild Server hostname")

	// APP logging
	cmd.Flags().StringVar(&cfg.Logging.LogLevel, "log.log_level", cfg.Logging.LogLevel, "Kwild app log level")

	cmd.Flags().StringVar(&cfg.Logging.LogFormat, "log.log_format", cfg.Logging.LogFormat, "Kwild app log format")

	cmd.Flags().StringSliceVar(&cfg.Logging.OutputPaths, "log.log_output_paths", cfg.Logging.OutputPaths, "Kwild app log output paths")

	// Extension endpoints flags
	cmd.Flags().StringSliceVar(&cfg.AppCfg.ExtensionEndpoints, "app.extension_endpoints", cfg.AppCfg.ExtensionEndpoints, "Kwild app extension endpoints")

	// Snapshot Config flags
	cmd.Flags().BoolVar(&cfg.AppCfg.SnapshotConfig.Enabled, "app.snapshots.enabled", cfg.AppCfg.SnapshotConfig.Enabled, "Enable snapshots")

	cmd.Flags().Uint64Var(&cfg.AppCfg.SnapshotConfig.RecurringHeight, "app.snapshots.recurring_height", cfg.AppCfg.SnapshotConfig.RecurringHeight, "Recurring snapshot height")

	cmd.Flags().Uint64Var(&cfg.AppCfg.SnapshotConfig.MaxSnapshots, "app.snapshots.max_snapshots", cfg.AppCfg.SnapshotConfig.MaxSnapshots, "Maximum snapshots")

	cmd.Flags().StringVar(&cfg.AppCfg.SnapshotConfig.SnapshotDir, "app.snapshots.snapshot_dir", cfg.AppCfg.SnapshotConfig.SnapshotDir, "Snapshot directory path")

	//  Basic Chain Config flags
	cmd.Flags().StringVar(&cfg.ChainCfg.Moniker, "chain.moniker", cfg.ChainCfg.Moniker, "Chain moniker")

	cmd.Flags().StringVar(&cfg.ChainCfg.Genesis, "chain.genesis", cfg.ChainCfg.Genesis, "Genesis file path")

	cmd.Flags().StringVar(&cfg.ChainCfg.DBPath, "chain.db_dir", cfg.ChainCfg.DBPath, "Chain database directory path")
	// Chain RPC flags
	cmd.Flags().StringVar(&cfg.ChainCfg.RPC.ListenAddress, "chain.rpc.laddr", cfg.ChainCfg.RPC.ListenAddress, "Chain RPC listen address")

	cmd.Flags().DurationVar(&cfg.ChainCfg.RPC.TimeoutBroadcastTxCommit, "chain.timeout_broadcast_tx_commit", cfg.ChainCfg.RPC.TimeoutBroadcastTxCommit, "chain timeout broadcast tx commit")

	// Chain P2P flags
	cmd.Flags().StringVar(&cfg.ChainCfg.P2P.ListenAddress, "chain.p2p.laddr", cfg.ChainCfg.P2P.ListenAddress, "chain P2P listen address")

	cmd.Flags().StringVar(&cfg.ChainCfg.P2P.ExternalAddress, "chain.p2p.external_address", cfg.ChainCfg.P2P.ExternalAddress, "chain P2P external address")

	cmd.Flags().StringVar(&cfg.ChainCfg.P2P.PersistentPeers, "chain.p2p.persistent_peers", cfg.ChainCfg.P2P.PersistentPeers, "chain P2P persistent peers")

	cmd.Flags().BoolVar(&cfg.ChainCfg.P2P.AddrBookStrict, "chain.p2p.addr_book_strict", cfg.ChainCfg.P2P.AddrBookStrict, "chain P2P address book strict")

	cmd.Flags().StringVar(&cfg.ChainCfg.P2P.UnconditionalPeerIDs, "chain.p2p.unconditional_peer_ids", cfg.ChainCfg.P2P.UnconditionalPeerIDs, "chain P2P unconditional peer IDs")

	cmd.Flags().IntVar(&cfg.ChainCfg.P2P.MaxNumInboundPeers, "chain.p2p.max_num_inbound_peers", cfg.ChainCfg.P2P.MaxNumInboundPeers, "chain P2P maximum number of inbound peers")

	cmd.Flags().IntVar(&cfg.ChainCfg.P2P.MaxNumOutboundPeers, "chain.p2p.max_num_outbound_peers", cfg.ChainCfg.P2P.MaxNumOutboundPeers, "chain P2P maximum number of outbound peers")

	cmd.Flags().BoolVar(&cfg.ChainCfg.P2P.AllowDuplicateIP, "chain.p2p.allow_duplicate_ip", cfg.ChainCfg.P2P.AllowDuplicateIP, "chain P2P allow duplicate IP")

	// Chain Mempool flags
	cmd.Flags().IntVar(&cfg.ChainCfg.Mempool.Size, "chain.mempool.size", cfg.ChainCfg.Mempool.Size, "chain mempool size")

	cmd.Flags().IntVar(&cfg.ChainCfg.Mempool.CacheSize, "chain.mempool.cache_size", cfg.ChainCfg.Mempool.CacheSize, "chain mempool cache size")

	cmd.Flags().Int64Var(&cfg.ChainCfg.Mempool.MaxTxsBytes, "chain.mempool.max_txs_bytes", cfg.ChainCfg.Mempool.MaxTxsBytes, "chain mempool maximum transactions bytes")

	// Chain Consensus flags
	cmd.Flags().DurationVar(&cfg.ChainCfg.Consensus.TimeoutPropose, "chain.consensus.timeout_propose", cfg.ChainCfg.Consensus.TimeoutPropose, "chain consensus timeout propose")

	cmd.Flags().DurationVar(&cfg.ChainCfg.Consensus.TimeoutPrevote, "chain.consensus.timeout_prevote", cfg.ChainCfg.Consensus.TimeoutPrevote, "chain consensus timeout prevote")

	cmd.Flags().DurationVar(&cfg.ChainCfg.Consensus.TimeoutPrecommit, "chain.consensus.timeout_precommit", cfg.ChainCfg.Consensus.TimeoutPrecommit, "chain consensus timeout precommit")

	cmd.Flags().DurationVar(&cfg.ChainCfg.Consensus.TimeoutCommit, "chain.consensus.timeout_commit", cfg.ChainCfg.Consensus.TimeoutCommit, "chain consensus timeout commit")

	// State Sync flags
	cmd.Flags().BoolVar(&cfg.ChainCfg.StateSync.Enable, "chain.state_sync.enable", cfg.ChainCfg.StateSync.Enable, "chain state sync enable")

	cmd.Flags().StringVar(&cfg.ChainCfg.StateSync.TempDir, "chain.state_sync.temp_dir", cfg.ChainCfg.StateSync.TempDir, "chain state sync temp dir")

	cmd.Flags().StringSliceVar(&cfg.ChainCfg.StateSync.RPCServers, "chain.state_sync.rpc_servers", cfg.ChainCfg.StateSync.RPCServers, "chain state sync rpc servers")

	cmd.Flags().DurationVar(&cfg.ChainCfg.StateSync.DiscoveryTime, "chain.state_sync.discovery_time", cfg.ChainCfg.StateSync.DiscoveryTime, "chain state sync discovery time")

	cmd.Flags().DurationVar(&cfg.ChainCfg.StateSync.ChunkRequestTimeout, "chain.state_sync.chunk_request_timeout", cfg.ChainCfg.StateSync.ChunkRequestTimeout, "chain state sync chunk request timeout")

	// Block sync can be added later (when they have more version of it)
}
