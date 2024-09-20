package config

import (
	"fmt"

	"github.com/kwilteam/kwil-db/cmd"
	"github.com/kwilteam/kwil-db/common/config"
	"github.com/spf13/pflag"
)

// AddConfigFlags adds all flags from KwildConfig to the given flagSet.
func AddConfigFlags(flagSet *pflag.FlagSet, cfg *config.KwildConfig) {
	format := func(s string) string {
		return fmt.Sprintf(s, cmd.BinaryConfig.NodeCmd)
	}

	flagSet.StringVarP(&cfg.RootDir, "root-dir", "r", "~/.kwild", format("%s root directory for config and data"))

	// logging
	flagSet.StringVarP(&cfg.Logging.Level, "log.level", "l", cfg.Logging.Level, format("%s log level"))
	flagSet.StringVar(&cfg.Logging.RPCLevel, "log.rpc-level", cfg.Logging.RPCLevel, "user rpc server log level")
	flagSet.StringVar(&cfg.Logging.ConsensusLevel, "log.consensus-level", cfg.Logging.ConsensusLevel, "consensus (cometbft) log level")
	flagSet.StringVar(&cfg.Logging.DBLevel, "log.db-level", cfg.Logging.DBLevel, "database backend (postgres) log level")
	flagSet.StringVar(&cfg.Logging.Format, "log.format", cfg.Logging.Format, format("%s log format"))
	flagSet.StringVar(&cfg.Logging.TimeEncoding, "log.time-format", cfg.Logging.TimeEncoding, format("%s time log format"))
	flagSet.StringSliceVar(&cfg.Logging.OutputPaths, "log.output-paths", cfg.Logging.OutputPaths, format("%s log output paths"))
	flagSet.Int64Var(&cfg.Logging.MaxLogSizeKB, "log.file-roll-size", cfg.Logging.MaxLogSizeKB, "threshold in KB at which the log file rolls over and archives the current one")
	flagSet.IntVar(&cfg.Logging.MaxLogRolls, "log.retain-max-rolls", cfg.Logging.MaxLogRolls, "retention limit on the number of archived log files to keep (default is 0, meaning retain all)")

	// General APP flags:
	flagSet.StringVar(&cfg.AppConfig.PrivateKeyPath, "app.private-key-path", cfg.AppConfig.PrivateKeyPath, "Path to the node private key file")
	flagSet.StringVar(&cfg.AppConfig.JSONRPCListenAddress, "app.jsonrpc-listen-addr", cfg.AppConfig.JSONRPCListenAddress, format("%s JSON-RPC listen address"))
	flagSet.StringVar(&cfg.AppConfig.AdminListenAddress, "app.admin-listen-addr", cfg.AppConfig.AdminListenAddress, format("%s admin listen address (unix or tcp)"))
	flagSet.StringVar(&cfg.AppConfig.AdminRPCPass, "app.admin-pass", cfg.AppConfig.AdminRPCPass, "password for the node's admin service (may be empty)")
	flagSet.BoolVar(&cfg.AppConfig.NoTLS, "app.admin-notls", cfg.AppConfig.NoTLS, "do not enable TLS on admin server (automatically disabled for unix socket or loopback listen addresses)")
	flagSet.StringVar(&cfg.AppConfig.TLSCertFile, "app.tls-cert-file", cfg.AppConfig.TLSCertFile, "TLS certificate file path for the admin and consensus RPC server (optional)")
	flagSet.StringVar(&cfg.AppConfig.TLSKeyFile, "app.tls-key-file", cfg.AppConfig.TLSKeyFile, "TLS key file path for the admin and consensus RPC servers (optional)")
	flagSet.StringVar(&cfg.AppConfig.Hostname, "app.hostname", cfg.AppConfig.Hostname, format("%s Server hostname"))
	flagSet.BoolVar(&cfg.AppConfig.PrivateRPC, "app.private-rpc", cfg.AppConfig.PrivateRPC, "Enforce data privacy with authenticated call request, disabled ad hoc queries, and no raw transaction retrieval")
	flagSet.Var(&cfg.AppConfig.ChallengeExpiry, "app.challenge-expiry", "Time after which a challenge expires")
	flagSet.Float64Var(&cfg.AppConfig.ChallengeRateLimit, "app.challenge-rate-limit", cfg.AppConfig.ChallengeRateLimit, "challenge request rate limit per second per client IP")

	flagSet.StringVar(&cfg.AppConfig.DBHost, "app.pg-db-host", cfg.AppConfig.DBHost, "PostgreSQL host address (no port)")
	flagSet.StringVar(&cfg.AppConfig.DBPort, "app.pg-db-port", cfg.AppConfig.DBPort, "PostgreSQL port")
	flagSet.StringVar(&cfg.AppConfig.DBUser, "app.pg-db-user", cfg.AppConfig.DBUser, "PostgreSQL user name")
	flagSet.StringVar(&cfg.AppConfig.DBPass, "app.pg-db-pass", cfg.AppConfig.DBPass, "PostgreSQL password name")
	flagSet.StringVar(&cfg.AppConfig.DBName, "app.pg-db-name", cfg.AppConfig.DBName, "PostgreSQL database name")

	flagSet.StringVar(&cfg.AppConfig.ProfileMode, "app.profile-mode", cfg.AppConfig.ProfileMode, format("%s profile mode (http, cpu, mem, mutex, or block)"))
	flagSet.StringVar(&cfg.AppConfig.ProfileFile, "app.profile-file", cfg.AppConfig.ProfileFile, format("%s profile output file path (e.g. cpu.pprof)"))

	flagSet.Var(&cfg.AppConfig.RPCTimeout, "app.rpc-timeout", "timeout for RPC requests (through reading the request, handling the request, and sending the response)")
	flagSet.IntVar(&cfg.AppConfig.RPCMaxReqSize, "app.rpc-max-req-size", cfg.AppConfig.RPCMaxReqSize, "RPC request size limit")
	flagSet.IntVar(&cfg.AppConfig.DEPRECATED_RPCReqLimit, "app.rpc-req-limit", cfg.AppConfig.DEPRECATED_RPCReqLimit, "RPC request size limit")
	flagSet.MarkDeprecated("app.rpc-req-limit", "use --app.rpc-max-req-size instead")

	flagSet.Var(&cfg.AppConfig.ReadTxTimeout, "app.db-read-timeout", "timeout for database reads initiated by RPC requests")

	// genesis state / migration flags
	flagSet.StringVar(&cfg.AppConfig.GenesisState, "app.genesis-state", cfg.AppConfig.GenesisState, "Path to the genesis state file")
	flagSet.StringVar(&cfg.AppConfig.MigrateFrom, "app.migrate-from", cfg.AppConfig.MigrateFrom, format("%s JSON-RPC listening address of the node to replicate the state from."))

	// Extension endpoints flags
	flagSet.StringSliceVar(&cfg.AppConfig.ExtensionEndpoints, "app.extension-endpoints", cfg.AppConfig.ExtensionEndpoints, format("%s extension endpoints"))

	// Snapshot Config flags
	flagSet.BoolVar(&cfg.AppConfig.Snapshots.Enabled, "app.snapshots.enabled", cfg.AppConfig.Snapshots.Enabled, "Enable snapshots")
	flagSet.Uint64Var(&cfg.AppConfig.Snapshots.RecurringHeight, "app.snapshots.recurring-height", cfg.AppConfig.Snapshots.RecurringHeight, "Recurring heights to create snapshots")
	flagSet.Uint64Var(&cfg.AppConfig.Snapshots.MaxSnapshots, "app.snapshots.max-snapshots", cfg.AppConfig.Snapshots.MaxSnapshots, "Maximum snapshots to store on disk. Default is 3. If max snapshots is reached, the oldest snapshot is deleted.")

	// Basic Chain Config flags
	flagSet.StringVar(&cfg.ChainConfig.Moniker, "chain.moniker", cfg.ChainConfig.Moniker, "Node moniker")

	// Chain RPC flags
	flagSet.StringVar(&cfg.ChainConfig.RPC.ListenAddress, "chain.rpc.listen-addr", cfg.ChainConfig.RPC.ListenAddress, "Chain RPC listen address")
	flagSet.Var(&cfg.ChainConfig.RPC.BroadcastTxTimeout, "chain.rpc.broadcast-tx-timeout", "Chain RPC broadcast transaction timeout")

	// Chain P2P flags
	flagSet.StringVar(&cfg.ChainConfig.P2P.ListenAddress, "chain.p2p.listen-addr", cfg.ChainConfig.P2P.ListenAddress, "Chain P2P listen address")
	flagSet.StringVar(&cfg.ChainConfig.P2P.ExternalAddress, "chain.p2p.external-address", cfg.ChainConfig.P2P.ExternalAddress, "Chain P2P external address to advertise")
	flagSet.StringVar(&cfg.ChainConfig.P2P.PersistentPeers, "chain.p2p.persistent-peers", cfg.ChainConfig.P2P.PersistentPeers, "Chain P2P persistent peers")
	flagSet.BoolVar(&cfg.ChainConfig.P2P.AddrBookStrict, "chain.p2p.addr-book-strict", cfg.ChainConfig.P2P.AddrBookStrict, "Chain P2P address book strict")
	flagSet.StringVar(&cfg.ChainConfig.P2P.UnconditionalPeerIDs, "chain.p2p.unconditional-peer-ids", cfg.ChainConfig.P2P.UnconditionalPeerIDs, "Chain P2P unconditional peer IDs")
	flagSet.IntVar(&cfg.ChainConfig.P2P.MaxNumInboundPeers, "chain.p2p.max-num-inbound-peers", cfg.ChainConfig.P2P.MaxNumInboundPeers, "Chain P2P maximum number of inbound peers")
	flagSet.IntVar(&cfg.ChainConfig.P2P.MaxNumOutboundPeers, "chain.p2p.max-num-outbound-peers", cfg.ChainConfig.P2P.MaxNumOutboundPeers, "Chain P2P maximum number of outbound peers")
	flagSet.BoolVar(&cfg.ChainConfig.P2P.AllowDuplicateIP, "chain.p2p.allow-duplicate-ip", cfg.ChainConfig.P2P.AllowDuplicateIP, "Chain P2P allow multiple peers with the same IP address")
	flagSet.BoolVar(&cfg.ChainConfig.P2P.PexReactor, "chain.p2p.pex", cfg.ChainConfig.P2P.PexReactor, "Enables peer information exchange")
	flagSet.StringVar(&cfg.ChainConfig.P2P.Seeds, "chain.p2p.seeds", cfg.ChainConfig.P2P.Seeds, "Seed nodes for obtaining peer addresses, if address book is empty")
	flagSet.BoolVar(&cfg.ChainConfig.P2P.SeedMode, "chain.p2p.seed-mode", cfg.ChainConfig.P2P.SeedMode, format(`Run %s in a special "seed" mode where it crawls the network for peer addresses,
sharing them with incoming peers before immediately disconnecting. It is recommended
to instead run a dedicated seeder like https://github.com/kwilteam/cometseed.`))
	flagSet.Var(&cfg.ChainConfig.P2P.HandshakeTimeout, "chain.p2p.handshake-timeout", "Chain P2P handshake timeout")
	flagSet.Var(&cfg.ChainConfig.P2P.DialTimeout, "chain.p2p.dial-timeout", "Chain P2P dial timeout")

	// Network flags
	flagSet.BoolVarP(&cfg.ChainConfig.P2P.PrivateMode, "chain.p2p.private-mode", "p", cfg.ChainConfig.P2P.PrivateMode, "Run the node in private mode. In private mode, the connectivity to the node is restricted to the current validators and whitelist peers.")
	flagSet.StringVar(&cfg.ChainConfig.P2P.WhitelistPeers, "chain.p2p.whitelist-peers", cfg.ChainConfig.P2P.WhitelistPeers, fmt.Sprintf("List of allowed sentry nodes that can connect to the node. Whitelist peers can be updated dynamically using `%s` commands.", cmd.BinaryConfig.AdminUsage()))

	// Chain Mempool flags
	flagSet.IntVar(&cfg.ChainConfig.Mempool.Size, "chain.mempool.size", cfg.ChainConfig.Mempool.Size, "Chain mempool size")
	flagSet.IntVar(&cfg.ChainConfig.Mempool.CacheSize, "chain.mempool.cache-size", cfg.ChainConfig.Mempool.CacheSize, "Chain mempool cache size")
	flagSet.IntVar(&cfg.ChainConfig.Mempool.MaxTxBytes, "chain.mempool.max-tx-bytes", cfg.ChainConfig.Mempool.MaxTxBytes, "chain mempool maximum single transaction size in bytes")
	flagSet.IntVar(&cfg.ChainConfig.Mempool.MaxTxsBytes, "chain.mempool.max-txs-bytes", cfg.ChainConfig.Mempool.MaxTxsBytes, "chain mempool maximum total transactions in bytes")

	// Chain Consensus flags
	flagSet.Var(&cfg.ChainConfig.Consensus.TimeoutPropose, "chain.consensus.timeout-propose", "Chain consensus timeout propose")
	flagSet.Var(&cfg.ChainConfig.Consensus.TimeoutPrevote, "chain.consensus.timeout-prevote", "Chain consensus timeout prevote")
	flagSet.Var(&cfg.ChainConfig.Consensus.TimeoutPrecommit, "chain.consensus.timeout-precommit", "Chain consensus timeout precommit")
	flagSet.Var(&cfg.ChainConfig.Consensus.TimeoutCommit, "chain.consensus.timeout-commit", "Chain consensus timeout commit")

	// State Sync flags
	flagSet.BoolVar(&cfg.ChainConfig.StateSync.Enable, "chain.statesync.enable", cfg.ChainConfig.StateSync.Enable, "Chain state sync enable")
	flagSet.StringVar(&cfg.ChainConfig.StateSync.RPCServers, "chain.statesync.rpc-servers", cfg.ChainConfig.StateSync.RPCServers, "Chain state sync rpc servers")
	flagSet.Var(&cfg.ChainConfig.StateSync.DiscoveryTime, "chain.statesync.discovery-time", "Chain state sync discovery time")
	flagSet.Var(&cfg.ChainConfig.StateSync.ChunkRequestTimeout, "chain.statesync.chunk-request-timeout", "Chain state sync chunk request timeout")

	// TODO: delete in v0.10.0
	flagSet.String("app.snapshots.snapshot-dir", "", "Snapshot directory path")
	flagSet.MarkDeprecated("app.snapshots.snapshot-dir", "this value is no longer configurable")
	flagSet.String("chain.statesync.snapshot-dir", "", "Chain state sync snapshot directory")
	flagSet.MarkDeprecated("chain.statesync.snapshot-dir", "this value is no longer configurable")
}
