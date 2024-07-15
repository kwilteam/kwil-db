package userjson

import jsonrpc "github.com/kwilteam/kwil-db/core/rpc/json"

const (
	MethodUserVersion           jsonrpc.Method = "user.version"
	MethodPing                  jsonrpc.Method = "user.ping"
	MethodChainInfo             jsonrpc.Method = "user.chain_info"
	MethodAccount               jsonrpc.Method = "user.account"
	MethodBroadcast             jsonrpc.Method = "user.broadcast"
	MethodCall                  jsonrpc.Method = "user.call"
	MethodDatabases             jsonrpc.Method = "user.databases"
	MethodPrice                 jsonrpc.Method = "user.estimate_price"
	MethodQuery                 jsonrpc.Method = "user.query"
	MethodTxQuery               jsonrpc.Method = "user.tx_query"
	MethodSchema                jsonrpc.Method = "user.schema"
	MethodMigrationStatus       jsonrpc.Method = "user.migration_status"
	MethodListMigrations        jsonrpc.Method = "user.list_migrations"
	MethodLoadChangeset         jsonrpc.Method = "user.changeset"
	MethodLoadChangesetMetadata jsonrpc.Method = "user.changeset_metadata"
	MethodMigrationMetadata     jsonrpc.Method = "user.migration_metadata"
	MethodMigrationGenesisChunk jsonrpc.Method = "user.migration_genesis_chunk"
)
