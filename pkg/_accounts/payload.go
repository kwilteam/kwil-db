package accounts

import "fmt"

type PayloadType int32

const (
	INVALID_PAYLOAD_TYPE PayloadType = iota + 100
	DEPLOY_DATABASE
	MODIFY_DATABASE
	DROP_DATABASE
	EXECUTE_QUERY
	WITHDRAW
	END_PAYLOAD_TYPE
)

func (x PayloadType) String() string {
	switch x {
	case INVALID_PAYLOAD_TYPE:
		return "INVALID_PAYLOAD_TYPE"
	case DEPLOY_DATABASE:
		return "DEPLOY_DATABASE"
	case MODIFY_DATABASE:
		return "MODIFY_DATABASE"
	case DROP_DATABASE:
		return "DROP_DATABASE"
	case EXECUTE_QUERY:
		return "EXECUTE_QUERY"
	case WITHDRAW:
		return "WITHDRAW"
	case END_PAYLOAD_TYPE:
		return "END_PAYLOAD_TYPE"
	default:
		return fmt.Sprintf("PayloadType(%d)", x)
	}
}
