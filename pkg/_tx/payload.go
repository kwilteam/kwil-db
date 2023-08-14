package tx

import (
	"fmt"
)

type PayloadType int32

const (
	INVALID_PAYLOAD_TYPE PayloadType = iota + 100
	DEPLOY_DATABASE
	DROP_DATABASE
	EXECUTE_ACTION
	VALIDATOR_JOIN
	VALIDATOR_LEAVE
	VALIDATOR_APPROVE
	END_PAYLOAD_TYPE
)

func (x PayloadType) String() string {
	switch x {
	case INVALID_PAYLOAD_TYPE:
		return "INVALID_PAYLOAD_TYPE"
	case DEPLOY_DATABASE:
		return "DEPLOY_DATABASE"
	case DROP_DATABASE:
		return "DROP_DATABASE"
	case EXECUTE_ACTION:
		return "EXECUTE_QUERY"
	case END_PAYLOAD_TYPE:
		return "END_PAYLOAD_TYPE"
	case VALIDATOR_JOIN:
		return "VALIDATOR_JOIN"
	case VALIDATOR_LEAVE:
		return "VALIDATOR_LEAVE"
	case VALIDATOR_APPROVE:
		return "VALIDATOR_APPROVE"
	default:
		return fmt.Sprintf("PayloadType(%d)", x)
	}
}

func (x PayloadType) IsValid() error {
	if x < INVALID_PAYLOAD_TYPE || x >= END_PAYLOAD_TYPE {
		return fmt.Errorf("invalid payload type '%d'", x)
	}
	return nil
}

func (x PayloadType) Int32() int32 {
	return int32(x)
}
