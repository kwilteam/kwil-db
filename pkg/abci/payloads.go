package abci

import engineTypes "github.com/kwilteam/kwil-db/pkg/engine/types"

type PayloadDeployDatabase struct {
	Schema *engineTypes.Schema
}

type PayloadDropDatabase struct {
	Name  string
	Owner string
}

type PayloadExecuteAction struct {
	Action string
	DBID   string
	Params []map[string]any
}

type PayloadCallAction struct {
	Action string
	DBID   string
	Params map[string]any
}

type PayloadValidatorJoin struct {
	Address string
}

type PayloadValidatorApprove struct {
	ValidatorToApprove string
	ApprovedBy         string
}

type PayloadEncoder interface {
	Encode(v any) ([]byte, error)
	Decode(data []byte, v any) error
}
