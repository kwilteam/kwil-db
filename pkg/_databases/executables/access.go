package executables

import (
	"fmt"
	"kwil/pkg/databases"
	"kwil/pkg/databases/spec"
)

func generateAccessParameters(db *databases.Database[*spec.KwilAny]) map[string]map[string]struct{} {
	access := make(map[string]map[string]struct{})
	for _, role := range db.Roles {
		access[role.Name] = make(map[string]struct{})
		for _, permission := range role.Permissions {
			access[role.Name][permission] = struct{}{}
		}
	}
	return access
}

func (d *DatabaseInterface) CanExecute(wallet, query string) bool {

	// check if the default roles have permission
	for _, role := range d.defaultRoles {
		v, ok := d.access[role][query]
		fmt.Println(v, ok)

		_, ok = d.access[role][query]
		if ok {
			return true
		}
	}

	// check if wallet is the owner
	return d.Owner == wallet

	// since we do not currently have ways of defining non-default roles, I will not implement any more logic here
}
