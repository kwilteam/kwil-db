package dto

import (
	"encoding/json"
	"fmt"
)

type PayloadType int32

const (
	INVALID_PAYLOAD_TYPE PayloadType = iota
	DEPLOY_DATABASE
	MODIFY_DATABASE
	DROP_DATABASE
	EXECUTE_QUERY
	WITHDRAW
	END_PAYLOAD_TYPE
)

func DecodePayload[T any](tx *Transaction) (T, error) {
	var p T
	err := json.Unmarshal(tx.Payload, &p)
	if err != nil {
		return p, fmt.Errorf("failed to unmarshal payload: %w", err)
	}
	return p, nil
}
