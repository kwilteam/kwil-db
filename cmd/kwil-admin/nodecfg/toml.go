package nodecfg

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/kwilteam/kwil-db/cmd/kwild/config"
)

var configTemplate *template.Template

func init() {
	var err error
	tmpl := template.New("configFileTemplate").Funcs(template.FuncMap{
		"arrayFormatter": arrayFormatter,
		"configDuration": func(d config.Duration) time.Duration {
			return time.Duration(d)
		},
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

// writeConfigFile writes the config to a file.
func writeConfigFile(configFilePath string, cfg *config.KwildConfig) error {
	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, cfg); err != nil {
		return err
	}

	return os.WriteFile(configFilePath, buffer.Bytes(), nodeDirPerm)
}

const defaultConfigTemplate = `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

# NOTE: Any path below can be absolute (e.g. "/var/myawesomeapp/data") or
# relative to the home directory (e.g. "data")

# Root Directory Structure:
# RootDir/
#   |- config.toml    (app and chain configuration for running the kwild node)
#   |- private_key   (node's private key)
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
#   |- signing/

# Only the config.toml and genesis file are required to run the kwild node
# The rest of the files & directories are created by the kwild node on startup

#######################################################################
###                    Logging Config Options                       ###
#######################################################################
[log]
# Output level for logging, default is "info". Other options are "debug", "error", "warn", "trace"
level = "{{ .Logging.Level }}"

# Output paths for the logger, can be stdout or a file path
output_paths = {{arrayFormatter .Logging.OutputPaths }}

# Output format: 'plain' or 'json'
format = "{{ .Logging.Format }}"

# Time format: "epochfloat" (default), "epochmilli", or "rfc3339milli"
time_format = "{{ .Logging.TimeEncoding }}"

#######################################################################
###                      App Config Options                         ###
#######################################################################

[app]
# Node's Private key
private_key_path = "{{ .AppCfg.PrivateKeyPath }}"

# TCP address for the KWILD App's GRPC server to listen on
grpc_listen_addr = "{{ .AppCfg.GrpcListenAddress }}"

# TCP address for the KWILD App's HTTP server to listen on
http_listen_addr = "{{ .AppCfg.HTTPListenAddress }}"

# Unix socket or TCP address for the KWILD App's Admin GRPC server to listen on
admin_listen_addr = "{{ .AppCfg.AdminListenAddress }}"

# List of Extension endpoints to be enabled ex: ["localhost:50052", "169.198.102.34:50053"]
extension_endpoints = {{arrayFormatter .AppCfg.ExtensionEndpoints}}

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
hostname = "{{ .AppCfg.Hostname }}"

#######################################################################
###                 Chain  Main Base Config Options                 ###
#######################################################################
[chain]

# A custom human readable name for this node
moniker = "{{ .ChainCfg.Moniker }}"

#######################################################################
###                 Advanced Configuration Options                  ###
#######################################################################

#######################################################
###       RPC Server Configuration Options          ###
#######################################################
[chain.rpc]

# TCP or UNIX socket address for the RPC server to listen on
listen_addr = "{{ .ChainCfg.RPC.ListenAddress }}"

# Timeout for each broadcast tx commit
broadcast_tx_timeout = "{{configDuration .ChainCfg.RPC.BroadcastTxTimeout }}"

#######################################################
###         Consensus Configuration Options         ###
#######################################################
[chain.consensus]

# How long we wait for a proposal block before prevoting nil
timeout_propose = "{{configDuration .ChainCfg.Consensus.TimeoutPropose }}"

# How long we wait after receiving +2/3 prevotes for “anything” (ie. not a single block or nil)
timeout_prevote = "{{configDuration .ChainCfg.Consensus.TimeoutPrevote }}"

# How long we wait after receiving +2/3 precommits for “anything” (ie. not a single block or nil)
timeout_precommit = "{{configDuration .ChainCfg.Consensus.TimeoutPrecommit }}"

# How long we wait after committing a block, before starting on the new
# height (this gives us a chance to receive some more precommits, even
# though we already have +2/3).
timeout_commit = "{{configDuration .ChainCfg.Consensus.TimeoutCommit }}"

#######################################################
###           P2P Configuration Options             ###
#######################################################
[chain.p2p]

# Address to listen for incoming connections
listen_addr = "{{ .ChainCfg.P2P.ListenAddress }}"

# Address to advertise to peers for them to dial
# If empty, will use the same port as the listening address,
# and will introspect on the listener or use UPnP
# to figure out the address. ip and port are required
# example: 159.89.10.97:26656
external_address = "{{ .ChainCfg.P2P.ExternalAddress }}"

# Comma separated list of nodes to keep persistent connections to (used for bootstrapping)
# Nodes should be identified as id@host:port, where id is the hex encoded CometBFT address.
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

# Enable gossiping of peer information
pex = {{ .ChainCfg.P2P.PexReactor }}

# Seed nodes used to obtain peer addresses. Only used if the peers in the
# address book are unreachable.
seeds = "{{ .ChainCfg.P2P.Seeds }}"

# Seed mode, in which node constantly crawls the network and looks for
# peers. If another node asks it for addresses, it responds and disconnects.
#
# It is recommended to instead run a dedicated seeder like https://github.com/kwilteam/cometseed.
#
# Requires peer-exchange to be enabled.
seed_mode = {{ .ChainCfg.P2P.SeedMode }}

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

# Limit the size of any one transaction in mempool.
max_tx_bytes = {{ .ChainCfg.Mempool.MaxTxBytes }}

# Size of the cache (used to filter transactions we saw earlier) in transactions
cache_size = {{ .ChainCfg.Mempool.CacheSize }}
`
