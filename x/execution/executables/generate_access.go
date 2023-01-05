package executables

import "kwil/x/execution/dto"

func GenerateAccessParameters(db *dto.Database) map[string]map[string]struct{} {
	access := make(map[string]map[string]struct{})
	for _, role := range db.Roles {
		access[role.Name] = make(map[string]struct{})
		for _, permission := range role.Permissions {
			access[role.Name][permission] = struct{}{}
		}
	}
	return access
}
