package validation

import (
	"fmt"
	"kwil/x/execution"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
)

func validateRoles(d *databases.Database[anytype.KwilAny]) error {
	// check amount of roles
	if len(d.Roles) > execution.MAX_ROLE_COUNT {
		return fmt.Errorf(`database must have at most %d roles`, execution.MAX_ROLE_COUNT)
	}

	// check unique role names and validate roles
	roles := make(map[string]struct{})
	for _, role := range d.Roles {
		// check if role name is unique
		if _, ok := roles[role.Name]; ok {
			return fmt.Errorf(`duplicate role name "%s"`, role.Name)
		}
		roles[role.Name] = struct{}{}

		err := ValidateRole(role, d)
		if err != nil {
			return fmt.Errorf(`error on role "%s": %w`, role.Name, err)
		}
	}

	return nil
}

func ValidateRole(role *databases.Role, db *databases.Database[anytype.KwilAny]) error {
	// check if role name is valid
	err := CheckName(role.Name, execution.MAX_ROLE_NAME_LENGTH)
	if err != nil {
		return err
	}

	// check if role permissions are valid
	permMap := make(map[string]struct{})
	for _, perm := range role.Permissions {
		// check if permission is unique
		if _, ok := permMap[perm]; ok {
			return fmt.Errorf(`duplicate permission "%s"`, perm)
		}
		permMap[perm] = struct{}{}

		// check if permission is valid
		qry := db.GetQuery(perm)
		if qry == nil {
			return fmt.Errorf(`query "%s" does not exist`, perm)
		}
	}

	return nil
}
