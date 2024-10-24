package types

type Role int

const (
	RoleLeader Role = iota
	RoleValidator
	RoleSentry
)

type Validator struct {
	Role   Role
	PubKey string
	Power  int64
}
