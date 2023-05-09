package validator

import (
	"fmt"
	"github.com/kwilteam/kwil-db/pkg/databases"
)

func (v *Validator) validateRoles() error {
	// validate role count
	err := validateRoleCount(v.DB.Roles)
	if err != nil {
		return fmt.Errorf(`invalid role count: %w`, err)
	}

	roleNames := make(map[string]struct{})
	for _, role := range v.DB.Roles {
		// validate role name is unique
		if _, ok := roleNames[role.Name]; ok {
			return violation(errorCode1300, fmt.Errorf(`duplicate role name "%s"`, role.Name))
		}
		roleNames[role.Name] = struct{}{}

		// validate role
		err := v.ValidateRole(role)
		if err != nil {
			return fmt.Errorf(`error on role %v: %w`, role.Name, err)
		}
	}

	return nil
}

func validateRoleCount(roles []*databases.Role) error {
	if len(roles) > MAX_ROLE_COUNT {
		return violation(errorCode1301, fmt.Errorf(`invalid role count: %d`, len(roles)))
	}

	return nil
}

func (v *Validator) ValidateRole(role *databases.Role) error {
	// validate role name
	if err := CheckName(role.Name, MAX_ROLE_NAME_LENGTH); err != nil {
		return violation(errorCode1400, err)
	}

	if isReservedWord(role.Name) {
		return violation(errorCode1403, fmt.Errorf(`role name "%s" is a reserved word`, role.Name))
	}

	// check permission uniqueness
	perms := make(map[string]struct{})
	for _, perm := range role.Permissions {
		if _, ok := perms[perm]; ok {
			return violation(errorCode1402, fmt.Errorf(`duplicate permission "%s"`, perm))
		}
		perms[perm] = struct{}{}

		// check if permission exists
		qry := v.DB.GetQuery(perm)
		if qry == nil {
			return violation(errorCode1401, fmt.Errorf(`query "%s" does not exist`, perm))
		}
	}

	return nil
}
