module github.com/kwilteam/kwil-db/cmd/cli

go 1.18

replace (
	github.com/kwilteam/kwil-db/internal => ../../internal
	github.com/kwilteam/kwil-db/cmd/cli/commands => ./cli/commands

	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)

require github.com/spf13/cobra v1.4.0

require (
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)
