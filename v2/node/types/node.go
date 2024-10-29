package types

type Role int

const (
	RoleLeader Role = iota
	RoleValidator
	RoleSentry
)

type Validator struct {
	PubKey []byte
	Power  int64
}
