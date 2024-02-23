module github.com/kwilteam/kwil-db/parse

go 1.21

require (
	github.com/antlr/antlr4/runtime/Go/antlr/v4 v4.0.0-20230512164433-5d1fd1a340c9
	github.com/kwilteam/action-grammar-go v0.0.1-0.20240221235853-2b171733810e
	github.com/kwilteam/sql-grammar-go v0.0.3-0.20240222213209-52b51de2eaf8
	github.com/pganalyze/pg_query_go/v5 v5.1.0 // This is only for unit testing
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/exp v0.0.0-20230817173708-d852ddb80c63 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
