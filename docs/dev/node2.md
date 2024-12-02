# Kwil v2 Orientation

## Repo Layout

### Top Level

Pretty much the same.

```text
.
├── app             shared components of CLI apps
├── cmd             CLI apps, composed from app pkgs
├── common          context and data structs for extensions
├── config          the main Config struct with toml tags
├── contrib         helper scripts, docker files, non-code, etc.
├── core            top-level utils, structs, logger, and client types
├── docs
├── extensions      extension packages for binary customization
├── node            the node (p2p and all the dependencies inc. consensus)
├── parse
├── test
├── testing
└── version
```

### Expanded

```text
kwil-db
├── app                     shared components of CLI apps
│   ├── custom
│   ├── key
│   ├── node                the "builder" for the actual node
│   │   └── conf            the **merged** config for the node (viper dies here)
│   ├── setup
│   └── shared
│       ├── bind
│       ├── display
│       ├── generate
│       └── version
├── cmd                     CLI apps, composed from app pkgs
│   ├── kwil-cli
│   │   ├── client          DialClient and the old "RoundTripper"
│   │   ├── cmds
│   │   ├── config          think ~/.kwil-cli/config.json
│   │   ├── csv
│   │   ├── generate
│   │   └── helpers
│   └── kwild
│       ├── generate
│       └── internal
├── common                  context and data structs for extensions
├── config                  the main Config struct with toml tags
├── contrib                 helper scripts, docker files, non-code, etc.
│   ├── docker
│   │   └── compose
│   ├── scripts
│   │   ├── build
│   │   ├── kuneiform
│   │   ├── mods
│   │   └── publish
│   └── systemd
├── core                    top-level utils, structs, logger, and client types
│   ├── client
│   │   └── types
│   ├── crypto
│   │   └── auth
│   ├── gatewayclient
│   ├── log
│   ├── rpc
│   │   ├── client
│   │   ├── json
│   │   └── transport
│   ├── types
│   │   ├── admin
│   │   ├── decimal
│   │   ├── serialize
│   │   └── validation
│   └── utils
│       ├── json
│       ├── order
│       └── random
├── docs
├── extensions              extension packages for binary customization
│   ├── auth
│   ├── consensus
│   ├── listeners
│   ├── precompiles
│   └── resolutions
├── node                    the node (p2p and all the dependencies inc. consensus)
│   ├── accounts
│   ├── admin
│   ├── consensus           the consensus engine (CE) where decisions are made
│   ├── engine              interpreter of kuneiform and user dataset SQL
│   │   ├── execution
│   │   ├── generate
│   │   ├── integration
│   │   └── testdata
│   ├── ident               end-user identity based core/crypto{,auth}
│   ├── mempool
│   ├── meta                chain metadata (current state)
│   ├── peers               node peer manager (PEX etc.)
│   ├── pg
│   ├── services
│   │   └── jsonrpc
│   ├── store               block store / index / tx index
│   ├── txapp
│   ├── types
│   │   └── sql
│   ├── utils
│   ├── versioning
│   └── voting              voting store a.k.a. validator+event store
├── parse
│   ├── gen
│   ├── grammar
│   ├── planner
│   ├── postgres
│   └── wasm
├── testing                 kuneiform testing framework
└── version                 global (sem)versioning
```

## Status

### New

Consensus Engine (CE) is newly in-house.

`node/consensus` contains the new consensus engine. This implements the logic outlined in the planning docs for the new >50% leader-based consensus algorithms.

`node/store` is the new block and tx store+index.

`node` contains the p2p layer and glue type (`Node`), which interfaces with our new CE and other systems including block+tx store.

`config` is the top level node config package, toml-tagged

`app/...` has many new packages to support both node (`kwild`) and user CLI (`kwil-cli`).

- `app/shared/bind` - higher level for viper-like config merge
- `app/node/config` - node-specific (`kwild` or custom binary) config merge based on `app/shared/bind` and `config`

`cmd/kwild` includes the commands from `kwil-admin` (comet's baggage is gone, and users want less bins).

### Needs Work

Validator/voting store needs methods for RPC et al.

`test` module is currently gone.  Needs to be recreated **after** tests using `go test` and mock networks are created. We should **TEST EVERYTHING** "e2e" using `Node` types in `go test ...` ***without Docker***! Test *fast* without CLI and network stack BS.

- `node/node_test.go` uses mock *everything* to test node's stream handlers (the p2p primitives)
- `node/node_live_test.go` uses the real CE to do a near-E2E test. This is a basis for more expansive test without Docker.

**REPEAT:** We should **TEST EVERYTHING** "e2e" using `Node` types in `go test ...` ***without Docker!*** -- **FAST** -- **Without CLI** -- **Without network stack**.

`engine` exists, but is weakly integrated.

`txapp` exists, but is not well integrated.

`services/{,*rpc}` exist but are poorly integrated. RPC server and clients exist, but have many `nil` dependencies and unimplemented methods.

`mempool` is a shell.

`app/node` builder / dep injection code is lifted for speed and can use re-do.

We are constantly discovering and realizing new p2p/comms requirements...
