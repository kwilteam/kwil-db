# Upgrading Kwil

This guide provides instructions for upgrading to specific [versions](https://github.com/kwilteam/kwil-db/releases) of `kwild`, `kwil-cli`, and `kwil-admin`. It also includes information on on changes to the Core Module (Go SDK).

## v0.8.0

Kwil v0.8 introduces substantial changes to the Kuneiform language (SQL smart contracts), an improved network migration process, a system for coordinating consensus rules changes, and a JSON-RPC server.

### Upgrading from v0.7.x

Nodes running `kwild` v0.8 are not compatible with nodes running `kwild` v0.7.

If a v0.7 network uses schemas that **are** compatible with v0.8, follow the [network migrations guide](https://docs.kwil.com/docs/node/network-migrations) to export the network data (schemas, accounts, and data within schemas) and import it into a new v0.8 network.

If a v0.7 network uses schemas that **are not** compatible with v0.8, you will need to deploy a fresh v0.8 network and migrate the Postgres data manually using the [batch command](https://docs.kwil.com/docs/ref/kwil-cli/database/batch).

To check if your schemas are compatible with v0.8, use the `kwil-cli utils parse` command in the v0.8 Kwil CLI to parse the schema files and output the schema as JSON. If the command fails, the schema is not compatible with v0.8.

### Kuneiform

- Kuneiform now includes procedures. Procedures are a strongly typed declarative syntax for implementing logic in Kuneiform. Procedures can be used to build access control, perform arithmetic, and manage control flow. - **[Learn more](https://docs.kwil.com/docs/kuneiform/procedures)**

- Kuneiform now includes a set of built-in functions for common operations like string manipulation, arithmetic, and uuid generation - **[Learn more](https://docs.kwil.com/docs/kuneiform/functions)**

- Kuneiform is now handled in kwil-db's `parse` module. Previously, Kuneiform was handled in a separate repository. The `parse` module will be tagged as `parse/v0.2.0` for the v0.8 release of Kwil DB.

#### Breaking Changes:

- SQL statements that are valid SQL but have invalid predicates will fail during parse, instead of at runtime.
- Fixed edge cases where users could perform cartesian joins. Now, one side of a join must be a column of unique values.
- Fixed a handful of cases that could lead to non-deterministic query results:
    - Disallowing conflicting column names when returned from a query.
    - Not returning unnamed columns in query results.
    - Applying deterministic ordering for joins against subqueries.

### Network Migrations

- Kwil has a network migration process that allows for exporting Postgres Data (schemas, accounts, and data within schemas) and importing it into a new network. This can be used to upgrade a network to a new version (assuming the old version does not use features that are incompatible with the new version) or to create a new network with the same data. - **[Learn more](https://docs.kwil.com/docs/node/network-migrations)**

### Coordinated Consensus Rules Changes

- Using the [extension system](https://docs.kwil.com/docs/extensions/overview), Kwil now has a system for nodes to coordinate changing consensus rules at a specific block height. This allows for a more flexible upgrades, such as changing the block size or adding new extensions. Previously, these changes could only be made via launching an entirely new network. - **[Learn more](https://docs.kwil.com/docs/extensions/fork-background)**
- Note: Consensus rule changes are still experimental and may change depending on feedback.

### JSON-RPC Server

- Kwil now includes a JSON-RPC server that allows for querying the Kwil node using JSON-RPC.
- The JSON-RPC listen address can be set with `jsonrpc_listen_addr` in the node's `config.toml` file. By default, the JSON-RPC server listens on port 8484.
- Kwil CLI, GO SDK, and JS SDK have been updated to use the JSON-RPC server.

#### Breaking Changes:

- The HTTP server (`http_listen_addr`) is DEPRECATED and will be removed in the susbequent release.
- The gRPC server (`grpc_listen_addr`) is no longer exposed.

### Node Configuration Changes

In the node's `config.toml` file, the following changes have been made:

- Added `app.rpc_timeout`, which imposes a timeout on RPC requests. The default is 45 seconds.
- Added `app.db_read_timeout`, which imposes a timout on read-only DB transactions (i.e. actions/procedures with a `view` tag or an ad-hoc SELECT query). The default 5 seconds.
- Added `rpc.broadcast_tx_timout`, which imposes a timeout for awaiting transaction confirmation when using the `--sync` flag in `kwil-cli`. The default is 10 seconds.
- Added `app.snapshots` to support snapshot creation. See the `SnapshotConfig` struct for more information.
- Added `app.genesis_state` to suport starting a network with a state at genesis. This is to support the Network Migrations feature.
- Added `chain.statesync` to enable new nodes syncing with snapshots instead of replaying each block. See the `StateSyncConfig` struct for more information.

### CLI Changes

- The `kwil-cli` has been updated to use the JSON-RPC server.
- The `--kwil-provider` flag is DEPRECATED and will be removed in the subsequent release. Use the `--provider` flag instead.
- Added `kwil-cli utils parse` to parse a Kuneiform file and output the schema as JSON.
- Added `kwil-cli utils decode-tx` to decode a transaction from a base64 string.

#### Breaking Changes

- Because Kwil-CLI now communicates over JSON-RPC, you should pass the JSON-RPC listen address to the CLI with the `--provider` flag or `kwil-cli configure` command.
- For uncomfirmed transactions, `kwild` now returns `null` for the `tx` field. Using older versions of `kwil-cli` to query unconfirmed transactions may have unexpected behavior.

### Kwil Admin

- Added `kwil-admin snapshot create` command for creating network snapshots (schemas, accounts, and data within schemas).

#### Breaking Changes

- The admin service now listens on `"/tmp/kwild.socket"` by default. Previously, the admin service listened on `"unix:///tmp/kwil_admin.sock"`.

### Core Module (Go SDK)

This section covers changes to the `core` module, which provides the SDK and client for Kwil. If you do not use the Go SDK, you can skip this section. Other Kwil tooling (e.g. CLI, JS SDK, etc.) account for these changes.

The `core` module is tagged as `core/v0.2.0` for the Kwil DB v0.8 release.

#### `core/crypto/auth`

##### Breaking Changes

- The JSON tags on the `Signature` struct have changed:

```go
type Signature struct {
    Signature []byte `json:"sig"` // was `json:"signature_bytes"`
    Type string `json:"type"` // was `json:"signature_type"`
}
```

#### `core/types`

- The `Schema` struct and all of it's composing types are now mirrored in `core/types.Schema`.
- Added many new type definitions to support procedures and strongly typed values. See `core/types/transactions/payload_schema.go`.

##### Breaking Changes

- There are many new types and fields on the `Schema` struct to support procedures and strongly typed values. Changes include:
    - `Owner` is now `HexBytes` instead of `byte[]`. `HexBytes` is an alias for `byte[]`; however, when marshaled to JSON, it is represented as a hex string.
    - `Table.Column.Type` now uses the `DataType` struct instead of a `string` to support strongly typed values.
    - Added the `Procedures` field to the `Schema` struct to support procedures.
    - Added the `ForeignProcedures` field to the `Schema` struct to support calling procedures that are defined in other schemas.

#### `core/types/transactions`

- **Note**: Most users do not need to be concerned with these changes as the `core/client` and `core/rpc/...` packages handle their use.
- Added the helper function `UnmarshalPayload` to assist in unmarshaling payloads from `byte[]` give its `PayloadType`.
- Added new `TxCode` values:

```go
    // engine-related error codes
    CodeInvalidSchema  TxCode = 100
    CodeDatasetMissing TxCode = 110
    CodeDatasetExists  TxCode = 120
```

##### Breaking Changes

- The `CallMessage`, `CallMessageBody`, and `Transaction` structs are now JSON-tagged for consistent snake-case marshaling.
- `CallMessage.Sender` and `Transaction.Sender` are now `HexBytes` instead of `byte[]`. This change was made because "senders" are usually hexadecimal Ethereum addresses.
- The `PayloadTypeExecuteAction` payload type is renamed to `PayloadTypeExecute` to reflect that it is used for executing procedures as well as actions.

#### `core/types/validation`

- Added the `validation` package to handle global limits and reserved keywords.

#### `core/types/transactions/payload_schema.go`

- Added `Call` and `Execute` to the `Client` interface to support calling procedures and actions.
- `CallAction` is DEPRECATED and will be removed in the subsequent release. Use `Call` instead.
- `ExecuteAction` is DEPRECATED and will be removed in the subsequent release. Use `Execute` instead.

##### Breaking Changes

- The `DeployDatabase` method's payload parameter is now `*types.Schema`. It was previously `*transactions.Schema`.
- The `GetSchema` method's return type is now `*types.Schema`. It was previously `*transactions.Schema`.

#### `core/client`

- Added detailed documentation in a `README.md` file.
- An example app in `core/client/example`.

##### Breaking Changes

- The `Client` type now uses the JSON-RPC server instead of the HTTP gateway to the gRPC server. This means that the `Client` only works with v0.8 nodes.

#### `core/rpc/client/...`

- **Note**: The packages unders `core/rpc/...` are relatively low-lvel and most users do not need to be concerned with these changes.
- Added the `core/rpc/client.JSONRPCClient` type to provider infrastructure for the "user", "admin", and "kgw" RPC services implemented in the `core/rpc/client/user/jsonrpc`, `core/rpc/client/admin/jsonrpc`, and `core/rpc/client/gateway/jsonrpc` packages.
- Added the `core/rpc/jsonrpc` package to define the common JSON-RPC types and methods.
- DEPRECATED the `core/rpc/http` package and its contents. This package will be removed in the subsequent release.