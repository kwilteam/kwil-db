package schema

func (db *Database) ListRoles() []string {
	var roles []string
	for k := range db.Roles {
		roles = append(roles, k)
	}
	return roles
}
