module github.com/kwilteam/kwil-db/extensions/encoding/borsh

go 1.23.5

replace github.com/kwilteam/kwil-db/core => ../../../core

replace github.com/kwilteam/kwil-db => ../../..

require (
	github.com/kwilteam/kwil-db v0.9.2-0.20250127164258-c637b4dcd403
	github.com/kwilteam/kwil-db/core v0.3.1-0.20250127164258-c637b4dcd403
	github.com/near/borsh-go v0.3.1
)

require (
	github.com/ethereum/go-ethereum v1.14.12 // indirect
	github.com/holiman/uint256 v1.3.2 // indirect

)
