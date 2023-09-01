package nodecfg

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/kwilteam/kwil-db/internal/app/kwild/config"
)

const defaultDirPerm = 0755

var configTemplate *template.Template

func init() {
	var err error
	tmpl := template.New("configFileTemplate").Funcs(template.FuncMap{
		"arrayFormatter": arrayFormatter,
	})
	if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

// arrayFormatter is a template function that formats a array of strings in to `["str1", "str2", ...]` in toml file.
func arrayFormatter(items []string) string {
	var formattedStrings []string
	for _, word := range items {
		formattedStrings = append(formattedStrings, fmt.Sprintf(`"%s"`, word))
	}
	return "[" + strings.Join(formattedStrings, ", ") + "]"
}

// kwildTemplateConfig
func writeConfigFile(configFilePath string, cfg *config.KwildConfig) {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, cfg); err != nil {
		panic(err)
	}

	os.WriteFile(configFilePath, buffer.Bytes(), defaultDirPerm)
}

const defaultConfigTemplate = `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

# NOTE: Any path below can be absolute (e.g. "/var/myawesomeapp/data") or
# relative to the home directory (e.g. "data")

# Home Directory Structure:
# HomeDir/
#   |- config.toml    (app and chain configuration for running the kwild node)
#   |- abci/
#   |   |- config/
#   |   |   |- genesis.json   (genesis file for the network)
#   |   |   |- addrbook.json  (peer routable addresses for the kwild node)
#   |   |- data/
#   |   |   |- blockchain db files/dir (blockstore.db, state.db, etc)
#   |   |- info/
#   |- application/wal
#   |- data
#   |   |- kwild.db/
#   |- snapshots/
#   |- signing/
#   |- rcvdSnaps/   (includes the chunks rcvd from the state sync module during db restoration process, its a temp dir)

# Only the config.toml and genesis file are required to run the kwild node
# The rest of the files & directories are created by the kwild node on startup

#######################################################################
###                    Logging Config Options                       ###
#######################################################################
[log]
# Output level for logging, default is "info". Other options are "debug", "error", "warn", "trace"
log_level = "{{ .Logging.LogLevel }}"

# Output paths for the logger, can be stdout or a file path
output_paths = {{arrayFormatter .Logging.OutputPaths }}

# Output format: 'plain' or 'json'
log_format = "{{ .Logging.LogFormat }}"

#######################################################################
###                      App Config Options                         ###
#######################################################################

[app]
# Node's Private key
private_key = "{{ .AppCfg.PrivateKey }}"

# TCP or UNIX socket address for the KWILD App's GRPC server to listen on
grpc_listen_addr = "{{ .AppCfg.GrpcListenAddress }}"

# TCP or UNIX socket address for the KWILD App's HTTP server to listen on
http_listen_addr = "{{ .AppCfg.HttpListenAddress }}"

# List of Extension endpoints to be enabled ex: ["localhost:50052", "169.198.102.34:50053"]
extension_endpoints = {{arrayFormatter .AppCfg.ExtensionEndpoints}}

# Toggle to enable gas costs for transactions and queries
without_gas_costs = {{ .AppCfg.WithoutGasCosts }}

# Toggle to disable nonces for transactions and queries
without_nonces = {{ .AppCfg.WithoutNonces }}

# KWILD Sqlite database file path
sqlite_file_path = "{{ .AppCfg.SqliteFilePath }}"

# The path to a file containing certificate that is used to create the HTTPS server.
# Might be either absolute path or path related to the kwild root directory.
# If the certificate is signed by a certificate authority,
# the certFile should be the concatenation of the server's certificate, any intermediates,
# and the CA's certificate.
# NOTE: both tls_cert_file and tls_key_file must be present for CometBFT to create HTTPS server.
# Otherwise, HTTP server is run.
tls_cert_file = "{{ .AppCfg.TLSCertFile }}"

# The path to a file containing matching private key that is used to create the HTTPS server.
# Might be either absolute path or path related to the kwild root directory.
# NOTE: both tls_cert_file and tls_key_file must be present for CometBFT to create HTTPS server.
# Otherwise, HTTP server is run.
tls_key_file = "{{ .AppCfg.TLSKeyFile }}"

# Kwild Server hostname
hostname = ""

#######################################################################
###                Snapshot store Config Options                    ###
#######################################################################
[app.snapshots]
# Toggle to enable snapshot store
# This would snapshot the application state at every snapshot_heights blocks
# and keep max_snapshots number of snapshots in the snapshot_dir
# Application state includes the databases deployed, accounts and the Validators db
enabled = {{ .AppCfg.SnapshotConfig.Enabled }}

# The height at which the snapshot is taken
snapshot_heights = {{ .AppCfg.SnapshotConfig.RecurringHeight }}

# Maximum number of snapshots to be kept in the snapshot_dir.
# If max limit is reached, the oldest would be deleted and replaced by the latest snapshot
max_snapshots = {{ .AppCfg.SnapshotConfig.MaxSnapshots}}

# The directory where the snapshots are stored. Can be absolute or relative to the kwild root directory
snapshot_dir = "{{ .AppCfg.SnapshotConfig.SnapshotDir }}"

#######################################################################
###                 Chain  Main Base Config Options                 ###
#######################################################################
[chain]
# A custom human readable name for this node
moniker = "{{ .ChainCfg.Moniker }}"

# Blockchain Genesis file
genesis_file = "{{ .ChainCfg.Genesis }}"

# Blockchain database directory
db_dir = "{{ .ChainCfg.DBPath }}"

#######################################################################
###                 Advanced Configuration Options                  ###
#######################################################################

#######################################################
###       RPC Server Configuration Options          ###
#######################################################
[chain.rpc]

# TCP or UNIX socket address for the RPC server to listen on
laddr = "{{ .ChainCfg.RPC.ListenAddress }}"

# How long to wait for a tx to be committed during /broadcast_tx_commit.
# WARNING: Using a value larger than 10s will result in increasing the
# global HTTP write timeout, which applies to all connections and endpoints.
# See https://github.com/tendermint/tendermint/issues/3435
timeout_broadcast_tx_commit = "{{ .ChainCfg.RPC.TimeoutBroadcastTxCommit }}"

#######################################################
###         Consensus Configuration Options         ###
#######################################################
[chain.consensus]

# How long we wait for a proposal block before prevoting nil
timeout_propose = "{{ .ChainCfg.Consensus.TimeoutPropose }}"

# How long we wait after receiving +2/3 prevotes for “anything” (ie. not a single block or nil)
timeout_prevote = "{{ .ChainCfg.Consensus.TimeoutPrevote }}"

# How long we wait after receiving +2/3 precommits for “anything” (ie. not a single block or nil)
timeout_precommit = "{{ .ChainCfg.Consensus.TimeoutPrecommit }}"

# How long we wait after committing a block, before starting on the new
# height (this gives us a chance to receive some more precommits, even
# though we already have +2/3).
timeout_commit = "{{ .ChainCfg.Consensus.TimeoutCommit }}"

#######################################################
###           P2P Configuration Options             ###
#######################################################
[chain.p2p]

# Address to listen for incoming connections
laddr = "{{ .ChainCfg.P2P.ListenAddress }}"

# Address to advertise to peers for them to dial
# If empty, will use the same port as the laddr,
# and will introspect on the listener or use UPnP
# to figure out the address. ip and port are required
# example: 159.89.10.97:26656
external_address = "{{ .ChainCfg.P2P.ExternalAddress }}"

# Comma separated list of nodes to keep persistent connections to (used for bootstrapping)
# Example: "d128266b8b9f64c313de466cf29e0a6182dba54d@172.10.100.2:26656,9440f4a8059cf7ff31454973c4f9c68de65fe526@172.10.100.3:26656"
persistent_peers = "{{ .ChainCfg.P2P.PersistentPeers }}"

# Set true for strict address routability rules
# Set false for private or local networks
addr_book_strict = {{ .ChainCfg.P2P.AddrBookStrict }}

# Maximum number of inbound peers
max_num_inbound_peers = {{ .ChainCfg.P2P.MaxNumInboundPeers }}

# Maximum number of outbound peers to connect to, excluding persistent peers
max_num_outbound_peers = {{ .ChainCfg.P2P.MaxNumOutboundPeers }}

# List of node IDs, to which a connection will be (re)established ignoring any existing limits
unconditional_peer_ids = "{{ .ChainCfg.P2P.UnconditionalPeerIDs }}"

# Toggle to disable guard against peers connecting from the same ip.
allow_duplicate_ip = {{ .ChainCfg.P2P.AllowDuplicateIP }}

#######################################################
###          Mempool Configuration Options          ###
#######################################################
[chain.mempool]
# Maximum number of transactions in the mempool
size = {{ .ChainCfg.Mempool.Size }}

# Limit the total size of all txs in the mempool.
# This only accounts for raw transactions (e.g. given 1MB transactions and
# max_txs_bytes=5MB, mempool will only accept 5 transactions).
max_txs_bytes = {{ .ChainCfg.Mempool.MaxTxsBytes }}

# Size of the cache (used to filter transactions we saw earlier) in transactions
cache_size = {{ .ChainCfg.Mempool.CacheSize }}

#######################################################
###         State Sync Configuration Options        ###
#######################################################
[chain.statesync]
# State sync rapidly bootstraps a new node by discovering, fetching, and restoring a state machine
# snapshot from peers instead of fetching and replaying historical blocks. Requires some peers in
# the network to take and serve state machine snapshots. State sync is not attempted if the node
# has any local state (LastBlockHeight > 0). The node will have a truncated block history,
# starting from the height of the snapshot.
enable = {{ .ChainCfg.StateSync.Enable }}

# RPC servers (comma-separated) for light client verification of the synced state machine and
# retrieval of state data for node bootstrapping. Also needs a trusted height and corresponding
# header hash obtained from a trusted source, and a period during which validators can be trusted.
#
# For Cosmos SDK-based chains, trust_period should usually be about 2/3 of the unbonding time (~2
# weeks) during which they can be financially punished (slashed) for misbehavior.
rpc_servers = {{arrayFormatter .ChainCfg.StateSync.RPCServers }}

# Temporary directory for state sync snapshot chunks, defaults to the OS tempdir (typically /tmp).
# Will create a new, randomly named directory within, and remove it when done.
temp_dir = "{{ .ChainCfg.StateSync.TempDir }}"

# Time to spend discovering snapshots before initiating a restore.
discovery_time = "{{ .ChainCfg.StateSync.DiscoveryTime }}"

# The timeout duration before re-requesting a chunk, possibly from a different
# peer (default: 1 minute).
chunk_request_timeout = "{{ .ChainCfg.StateSync.ChunkRequestTimeout }}"
`
