package validator

import (
	"fmt"
	databases2 "kwil/pkg/types/databases"
)

func (v *Validator) validateRoles() error {
	// validate role count
	err := validateRoleCount(v.db.Roles)
	if err != nil {
		return fmt.Errorf(`invalid role count: %w`, err)
	}

	roleNames := make(map[string]struct{})
	for _, role := range v.db.Roles {
		// validate role name is unique
		if _, ok := roleNames[role.Name]; ok {
			return fmt.Errorf(`duplicate role name "%s"`, role.Name)
		}
		roleNames[role.Name] = struct{}{}

		// validate role
		err := v.validateRole(role)
		if err != nil {
			return fmt.Errorf(`error on role %v: %w`, role.Name, err)
		}
	}

	return nil
}

func validateRoleCount(roles []*databases2.Role) error {
	if len(roles) > databases2.MAX_ROLE_COUNT {
		return fmt.Errorf(`too many roles: %v > %v`, len(roles), databases2.MAX_ROLE_COUNT)
	}

	return nil
}

func (v *Validator) validateRole(role *databases2.Role) error {
	// validate role name
	err := validateRoleName(role)
	if err != nil {
		return fmt.Errorf(`invalid role name: %w`, err)
	}

	// check permission uniqueness
	perms := make(map[string]struct{})
	for _, perm := range role.Permissions {
		if _, ok := perms[perm]; ok {
			return fmt.Errorf(`duplicate permission "%s"`, perm)
		}
		perms[perm] = struct{}{}

		// check if permission exists
		qry := v.db.GetQuery(perm)
		if qry == nil {
			return fmt.Errorf(`query "%s" does not exist`, perm)
		}
	}

	return nil
}

func validateRoleName(role *databases2.Role) error {
	return CheckName(role.Name, databases2.MAX_ROLE_NAME_LENGTH)
}
