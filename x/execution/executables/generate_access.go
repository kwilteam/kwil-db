package executables

import (
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
)

func generateAccessParameters(db *databases.Database[anytype.KwilAny]) map[string]map[string]struct{} {
	access := make(map[string]map[string]struct{})
	for _, role := range db.Roles {
		access[role.Name] = make(map[string]struct{})
		for _, permission := range role.Permissions {
			access[role.Name][permission] = struct{}{}
		}
	}
	return access
}
