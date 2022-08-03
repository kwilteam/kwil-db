module github.com/kwilteam/kwil-db/internal

go 1.18

require github.com/tidwall/wal v1.1.7

require (
	github.com/tidwall/gjson v1.10.2 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	github.com/tidwall/tinylru v1.1.0 // indirect
)

// This will mean an explicit 'require' is 
// needed for usage of local packages/modules.
// Intended to reduce implicit coupling.
replace (
	github.com/kwilteam/kwil-db => ./FORBIDDEN_DEPENDENCY
)