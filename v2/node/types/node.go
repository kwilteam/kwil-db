package types

import (
	"fmt"
)

type Role int

const (
	RoleLeader Role = iota
	RoleValidator
	RoleSentry
)

func (r Role) String() string {
	switch r {
	case RoleLeader:
		return "leader"
	case RoleValidator:
		return "validator"
	case RoleSentry:
		return "sentry"
	default:
		return "unknown"
	}
}

type Validator struct {
	PubKey HexBytes `json:"pubkey"`
	Power  int64    `json:"power"`
}

func (v Validator) String() string {
	return fmt.Sprintf("Validator{PubKey: %s, Power: %d}", v.PubKey, v.Power)
}
