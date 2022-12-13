package cache

import (
	"kwil/x/sqlx/models"
)

type Role struct {
	Name        string
	Permissions map[string]bool
}

func (c *Role) From(role *models.Role) error {
	c.Name = role.Name
	c.Permissions = make(map[string]bool)
	for _, perm := range role.Permissions {
		c.Permissions[perm] = true
	}
	return nil // making this return an error since it will in the future and the rest of the methods do
}

func (c *Role) HasPermission(perm string) bool {
	_, ok := c.Permissions[perm]
	return ok
}
