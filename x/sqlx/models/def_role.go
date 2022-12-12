package models

import (
	"fmt"
	types "kwil/x/sqlx/spec"
)

type Role struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

func (r *Role) Validate(db *Database) error {
	// check if role name is valid
	err := CheckName(r.Name, types.ROLE)
	if err != nil {
		return err
	}

	// check if role permissions are valid
	permMap := make(map[string]struct{})
	for _, perm := range r.Permissions {
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
