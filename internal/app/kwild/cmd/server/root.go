package server

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	// shorthand for chain client service

	"github.com/cstockton/go-conv"
	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
	"github.com/kwilteam/kwil-db/internal/app/kwild/server"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/spf13/cobra"
)

var kwildConf = config.DefaultConfig()

func NewStartCmd() *cobra.Command {
	AddKwildFlags(startCmd)
	return startCmd
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "kwil grpc server",
	Long:  "Starts node with Kwild and CometBFT services",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		ctx, cancel := context.WithCancel(ctx)

		go func() {
			<-signalChan
			cancel()
		}()

		svr, err := server.BuildKwildServer(ctx, kwildConf)
		if err != nil {
			return err
		}

		return svr.Start(ctx)

	},
}

func init() {
	rootDir, err := RootDir()
	if err != nil {
		panic(err)
	}
	err = kwildConf.LoadKwildConfig(rootDir)
	if err != nil {
		panic(err)
	}

	privateKey, err := crypto.Ed25519PrivateKeyFromHex(kwildConf.AppCfg.PrivateKey)
	if err != nil {
		panic(err)
	}
	kwildConf.PrivateKey = privateKey
}

func RootDir() (string, error) {
	val := os.Getenv("KWIL_ROOT_DIR")
	if val == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			// if `home` env(depends on OS) is not set, complain
			// we can use '/tmp/.kwil' or '.kwil' in this case, but it's not a good idea
			return "", err
		}
		return filepath.Join(home, ".kwil"), err
	}

	root_dir, err := conv.String(val)
	if err != nil {
		return "", err
	}

	return filepath.Clean(root_dir), nil
}
func AddKwildFlags(cmd *cobra.Command) {
	// General APP flags:
	cmd.Flags().StringVar(&kwildConf.AppCfg.PrivateKey, "app.private_key", kwildConf.AppCfg.PrivateKey, "Kwild app's private key")
	cmd.Flags().StringVar(&kwildConf.AppCfg.GrpcListenAddress, "app.grpc_laddr", kwildConf.AppCfg.GrpcListenAddress, "Kwild app gRPC listen address")
	cmd.Flags().StringVar(&kwildConf.AppCfg.HttpListenAddress, "app.http_laddr", kwildConf.AppCfg.HttpListenAddress, "Kwild app HTTP listen address")
	cmd.Flags().StringVar(&kwildConf.AppCfg.SqliteFilePath, "sqlite_file_path", kwildConf.AppCfg.SqliteFilePath, "Kwild app sqlite file path")
	cmd.Flags().BoolVar(&kwildConf.AppCfg.WithoutGasCosts, "app.without_gas_costs", kwildConf.AppCfg.WithoutGasCosts, "Kwild app without gas costs")
	cmd.Flags().BoolVar(&kwildConf.AppCfg.WithoutNonces, "app.without_nonces", kwildConf.AppCfg.WithoutNonces, "Kwild app without nonces")
	cmd.Flags().StringVar(&kwildConf.AppCfg.TLSCertFile, "app.tls_cert_file", kwildConf.AppCfg.TLSCertFile, "TLS certificate file path for RPC Server")
	cmd.Flags().StringVar(&kwildConf.AppCfg.TLSKeyFile, "app.tls_key_file", kwildConf.AppCfg.TLSKeyFile, "TLS key file path for RPC Server")

	// APP logging
	cmd.Flags().StringVar(&kwildConf.Logging.LogLevel, "log_level", kwildConf.Logging.LogLevel, "Kwild app log level")
	cmd.Flags().StringVar(&kwildConf.Logging.LogFormat, "log_format", kwildConf.Logging.LogFormat, "Kwild app log format")
	cmd.Flags().StringSliceVar(&kwildConf.Logging.OutputPaths, "log_output_paths", kwildConf.Logging.OutputPaths, "Kwild app log output paths")

	// Extension endpoints flags
	cmd.Flags().StringSliceVar(&kwildConf.AppCfg.ExtensionEndpoints, "extension_endpoints", kwildConf.AppCfg.ExtensionEndpoints, "Kwild app extension endpoints")

	// Snapshot Config flags
	cmd.Flags().BoolVar(&kwildConf.AppCfg.SnapshotConfig.Enabled, "snapshots.enabled", kwildConf.AppCfg.SnapshotConfig.Enabled, "Enable snapshots")
	cmd.Flags().Uint64Var(&kwildConf.AppCfg.SnapshotConfig.RecurringHeight, "snapshots.recurring_height", kwildConf.AppCfg.SnapshotConfig.RecurringHeight, "Recurring snapshot height")
	cmd.Flags().Uint64Var(&kwildConf.AppCfg.SnapshotConfig.MaxSnapshots, "snapshots.max_snapshots", kwildConf.AppCfg.SnapshotConfig.MaxSnapshots, "Maximum snapshots")
	cmd.Flags().StringVar(&kwildConf.AppCfg.SnapshotConfig.SnapshotDir, "snapshots.snapshot_dir", kwildConf.AppCfg.SnapshotConfig.SnapshotDir, "Snapshot directory path")

	//  Basic Chain Config flags
	cmd.Flags().StringVar(&kwildConf.ChainCfg.Moniker, "moniker", kwildConf.ChainCfg.Moniker, "Chain moniker")

	// Chain RPC flags
	cmd.Flags().StringVar(&kwildConf.ChainCfg.RPC.ListenAddress, "chain.rpc_laddr", kwildConf.ChainCfg.RPC.ListenAddress, "Chain RPC listen address")
	cmd.Flags().IntVar(&kwildConf.ChainCfg.RPC.MaxOpenConnections, "chain.max_open_connections", kwildConf.ChainCfg.RPC.MaxOpenConnections, "chain maximum open connections")
	cmd.Flags().DurationVar(&kwildConf.ChainCfg.RPC.TimeoutBroadcastTxCommit, "chain.timeout_broadcast_tx_commit", kwildConf.ChainCfg.RPC.TimeoutBroadcastTxCommit, "chain timeout broadcast tx commit")

	// Chain P2P flags
	cmd.Flags().StringVar(&kwildConf.ChainCfg.P2P.ListenAddress, "chain.p2p.laddr", kwildConf.ChainCfg.P2P.ListenAddress, "chain P2P listen address")
	cmd.Flags().StringVar(&kwildConf.ChainCfg.P2P.ExternalAddress, "chain.p2p.external_address", kwildConf.ChainCfg.P2P.ExternalAddress, "chain P2P external address")
	cmd.Flags().StringVar(&kwildConf.ChainCfg.P2P.Seeds, "chain.p2p.seeds", kwildConf.ChainCfg.P2P.Seeds, "chain P2P seeds")
	cmd.Flags().StringVar(&kwildConf.ChainCfg.P2P.PersistentPeers, "chain.p2p.persistent_peers", kwildConf.ChainCfg.P2P.PersistentPeers, "chain P2P persistent peers")
	cmd.Flags().BoolVar(&kwildConf.ChainCfg.P2P.UPNP, "chain.upnp", kwildConf.ChainCfg.P2P.UPNP, "chain P2P UPNP")
	cmd.Flags().BoolVar(&kwildConf.ChainCfg.P2P.AddrBookStrict, "chain.addr_book_strict", kwildConf.ChainCfg.P2P.AddrBookStrict, "chain P2P address book strict")
	cmd.Flags().StringVar(&kwildConf.ChainCfg.P2P.UnconditionalPeerIDs, "chain.unconditional_peer_ids", kwildConf.ChainCfg.P2P.UnconditionalPeerIDs, "chain P2P unconditional peer IDs")

	cmd.Flags().IntVar(&kwildConf.ChainCfg.P2P.MaxNumInboundPeers, "chain.max_num_inbound_peers", kwildConf.ChainCfg.P2P.MaxNumInboundPeers, "chain P2P maximum number of inbound peers")
	cmd.Flags().IntVar(&kwildConf.ChainCfg.P2P.MaxNumOutboundPeers, "chain.max_num_outbound_peers", kwildConf.ChainCfg.P2P.MaxNumOutboundPeers, "chain P2P maximum number of outbound peers")
	cmd.Flags().DurationVar(&kwildConf.ChainCfg.P2P.FlushThrottleTimeout, "chain.flush_throttle_timeout", kwildConf.ChainCfg.P2P.FlushThrottleTimeout, "chain P2P flush throttle timeout")

	cmd.Flags().IntVar(&kwildConf.ChainCfg.P2P.MaxPacketMsgPayloadSize, "chain.max_packet_msg_payload_size", kwildConf.ChainCfg.P2P.MaxPacketMsgPayloadSize, "chain P2P maximum packet message payload size")
	cmd.Flags().Int64Var(&kwildConf.ChainCfg.P2P.SendRate, "chain.send_rate", kwildConf.ChainCfg.P2P.SendRate, "chain P2P send rate")
	cmd.Flags().Int64Var(&kwildConf.ChainCfg.P2P.RecvRate, "chain.recv_rate", kwildConf.ChainCfg.P2P.RecvRate, "chain P2P receive rate")

	cmd.Flags().BoolVar(&kwildConf.ChainCfg.P2P.SeedMode, "chain.seed_mode", kwildConf.ChainCfg.P2P.SeedMode, "chain P2P seed mode")
	cmd.Flags().StringVar(&kwildConf.ChainCfg.P2P.PrivatePeerIDs, "chain.private_peer_ids", kwildConf.ChainCfg.P2P.PrivatePeerIDs, "chain P2P private peer IDs")
	cmd.Flags().BoolVar(&kwildConf.ChainCfg.P2P.AllowDuplicateIP, "chain.allow_duplicate_ip", kwildConf.ChainCfg.P2P.AllowDuplicateIP, "chain P2P allow duplicate IP")

	// Chain Mempool flags
	cmd.Flags().IntVar(&kwildConf.ChainCfg.Mempool.Size, "chain.mempool.size", kwildConf.ChainCfg.Mempool.Size, "chain mempool size")
	cmd.Flags().IntVar(&kwildConf.ChainCfg.Mempool.CacheSize, "chain.mempool.cache_size", kwildConf.ChainCfg.Mempool.CacheSize, "chain mempool cache size")
	cmd.Flags().Int64Var(&kwildConf.ChainCfg.Mempool.MaxTxsBytes, "chain.mempool.max_txs_bytes", kwildConf.ChainCfg.Mempool.MaxTxsBytes, "chain mempool maximum transactions bytes")

	// Chain Consensus flags
	cmd.Flags().DurationVar(&kwildConf.ChainCfg.Consensus.TimeoutPropose, "chain.consensus.timeout_propose", kwildConf.ChainCfg.Consensus.TimeoutPropose, "chain consensus timeout propose")
	cmd.Flags().DurationVar(&kwildConf.ChainCfg.Consensus.TimeoutPrevote, "chain.consensus.timeout_prevote", kwildConf.ChainCfg.Consensus.TimeoutPrevote, "chain consensus timeout prevote")
	cmd.Flags().DurationVar(&kwildConf.ChainCfg.Consensus.TimeoutPrecommit, "chain.consensus.timeout_precommit", kwildConf.ChainCfg.Consensus.TimeoutPrecommit, "chain consensus timeout precommit")
	cmd.Flags().DurationVar(&kwildConf.ChainCfg.Consensus.TimeoutCommit, "chain.consensus.timeout_commit", kwildConf.ChainCfg.Consensus.TimeoutCommit, "chain consensus timeout commit")

	// State Sync flags
	cmd.Flags().BoolVar(&kwildConf.ChainCfg.StateSync.Enable, "chain.state_sync.enable", kwildConf.ChainCfg.StateSync.Enable, "chain state sync enable")
	cmd.Flags().StringVar(&kwildConf.ChainCfg.StateSync.TempDir, "chain.state_sync.temp_dir", kwildConf.ChainCfg.StateSync.TempDir, "chain state sync temp dir")
	cmd.Flags().StringSliceVar(&kwildConf.ChainCfg.StateSync.RPCServers, "chain.state_sync.rpc_servers", kwildConf.ChainCfg.StateSync.RPCServers, "chain state sync rpc servers")
	cmd.Flags().DurationVar(&kwildConf.ChainCfg.StateSync.DiscoveryTime, "chain.state_sync.discovery_time", kwildConf.ChainCfg.StateSync.DiscoveryTime, "chain state sync discovery time")
	cmd.Flags().DurationVar(&kwildConf.ChainCfg.StateSync.ChunkRequestTimeout, "chain.state_sync.chunk_request_timeout", kwildConf.ChainCfg.StateSync.ChunkRequestTimeout, "chain state sync chunk request timeout")

	// Block sync can be added later (when they have more version of it)

	// Instrumentation Config flags
	cmd.Flags().BoolVar(&kwildConf.ChainCfg.Instrumentation.Prometheus, "chain.instrumentation.prometheus", kwildConf.ChainCfg.Instrumentation.Prometheus, "chain instrumentation prometheus")
	cmd.Flags().StringVar(&kwildConf.ChainCfg.Instrumentation.PrometheusListenAddr, "chain.instrumentation.prometheus_laddr", kwildConf.ChainCfg.Instrumentation.PrometheusListenAddr, "chain instrumentation prometheus listen address")
}
