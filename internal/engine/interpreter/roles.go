package interpreter

import (
	"context"
	"fmt"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/internal/sql/pg"
)

// the following are built-in roles that are always available.
const (
	ownerRole   = "owner"
	defaultRole = "default"
)

func isBuiltInRole(role string) bool {
	return role == ownerRole || role == defaultRole
}

/*
	This file includes a lot of the functionality for roles and access control.
	This could be made much more efficient by caching role information, but for simplicity
	we will access the db each time. The only exception is for the default role, since it will cover
	the majority of cases.
*/

func newAccessController(ctx context.Context, db sql.DB) (*accessController, error) {
	ac := &accessController{
		roles:     make(map[string]*perms),
		userRoles: make(map[string][]string),
	}

	// get the owner
	err := pg.QueryRowFunc(ctx, db, "SELECT value FROM kwild_engine.metadata WHERE key = $1", []any{&ac.owner}, func() error {
		return nil
	}, ownerKey)
	if err != nil {
		return nil, err
	}

	getRolesStmt := `
	SELECT
		r.name AS name,
		array_agg(ur.user_identifier) AS users,
		array_agg(rp.privilege_type) AS privileges,
		array_agg(n.name) AS namespaces
	FROM kwild_engine.roles r
	LEFT JOIN kwild_engine.role_privileges rp ON rp.role_id = r.id
	JOIN kwild_engine.namespaces n ON rp.namespace_id = n.id
	LEFT JOIN kwild_engine.user_roles ur ON ur.role_id = r.id
	GROUP BY r.id
	`

	// list all roles, their perms, and users
	var roleName string
	var users []string
	var privileges []string
	var namespaces []string
	err = pg.QueryRowFunc(ctx, db, getRolesStmt, []any{&roleName, &users, &privileges, &namespaces}, func() error {
		perm := &perms{
			namespacePrivileges: make(map[string]map[privilege]struct{}),
			globalPrivileges:    make(map[privilege]struct{}),
		}

		for i, priv := range privileges {
			// check that the privilege exists
			// This should never not be the case, but it is good to check
			_, ok := privilegeNames[privilege(priv)]
			if !ok {
				return fmt.Errorf(`unknown privilege "%s" stored in DB`, priv)
			}

			if namespaces[i] == "" {
				perm.globalPrivileges[privilege(priv)] = struct{}{}
			} else {
				if _, ok := perm.namespacePrivileges[namespaces[i]]; !ok {
					perm.namespacePrivileges[namespaces[i]] = make(map[privilege]struct{})
				}

				perm.namespacePrivileges[namespaces[i]][privilege(priv)] = struct{}{}
			}
		}

		ac.roles[roleName] = perm

		for _, user := range users {
			// we dont need to check for existence in the userRoles map since if it does not exist,
			// it will be a 0 value slice
			ac.userRoles[user] = append(ac.userRoles[user], roleName)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return ac, nil
}

// accessController enforces access control on the database.
type accessController struct {
	owner     string // the db owner
	roles     map[string]*perms
	userRoles map[string][]string // a map of user public keys to the roles they have. It does _not_ include the default role.
}

// CreateRole adds a new role to the access controller.
func (a *accessController) CreateRole(ctx context.Context, db sql.DB, role string) error {
	if isBuiltInRole(role) {
		return fmt.Errorf(`role "%s" is a built-in role and cannot be added`, role)
	}

	_, ok := a.roles[role]
	if ok {
		return fmt.Errorf(`role "%s" already exists`, role)
	}

	err := createRole(ctx, db, role)
	if err != nil {
		return err
	}

	a.roles[role] = &perms{
		namespacePrivileges: make(map[string]map[privilege]struct{}),
		globalPrivileges:    make(map[privilege]struct{}),
	}

	return nil
}

func (a *accessController) DeleteRole(ctx context.Context, db sql.DB, role string) error {
	if isBuiltInRole(role) {
		return fmt.Errorf(`role "%s" is a built-in role and cannot be removed`, role)
	}

	_, ok := a.roles[role]
	if !ok {
		return fmt.Errorf(`role "%s" does not exist`, role)
	}

	// remove the role from the db
	_, err := db.Execute(ctx, "DELETE FROM kwild_engine.roles WHERE name = $1", role)
	if err != nil {
		return err
	}

	delete(a.roles, role)

	// iterate over all users and remove the role from them
	for user, roles := range a.userRoles {
		for i, r := range roles {
			if r == role {
				a.userRoles[user] = append(roles[:i], roles[i+1:]...)
				break
			}
		}
	}

	return nil
}

// DeleteNamespace deletes all roles and privileges associated with a namespace.
func (a *accessController) DeleteNamespace(namespace string) {
	for _, role := range a.roles {
		delete(role.namespacePrivileges, namespace)
	}
}

func (a *accessController) HasPrivilege(user string, namespace *string, privilege privilege) bool {
	// if it is the owner, they have all privileges
	if user == a.owner {
		return true
	}

	// since all users have the default role, we can check that first
	if a.roles[defaultRole].canDo(privilege, namespace) {
		return true
	}

	// otherwise, we need to check the user's roles
	roles, ok := a.userRoles[user]
	if !ok {
		return false
	}

	for _, role := range roles {
		perms, ok := a.roles[role]
		if !ok {
			fmt.Println("Unexpected cache error: role does not exist. This is a bug.")
			continue
		}

		if perms.canDo(privilege, namespace) {
			return true
		}
	}

	return false
}

func (a *accessController) GrantPrivileges(ctx context.Context, db sql.DB, role string, privs []string, namespace *string) error {
	if role == ownerRole {
		return fmt.Errorf(`owner role already has all privileges`)
	}

	perms, ok := a.roles[role]
	if !ok {
		return fmt.Errorf(`role "%s" does not exist`, role)
	}

	// verify that the privileges are valid
	convPrivs, err := validatePrivileges(privs...)
	if err != nil {
		return err
	}

	// if a namespace is provided, check that it exists and that all privileges can be namespaced
	if namespace != nil {
		_, ok := perms.namespacePrivileges[*namespace]
		if !ok {
			return fmt.Errorf(`namespace "%s" does not exist`, *namespace)
		}

		err = canBeNamespaced(convPrivs...)
		if err != nil {
			return err
		}
	}

	for _, p := range convPrivs {
		if perms.canDo(p, namespace) {
			return fmt.Errorf(`role "%s" already has some or all of the specified privileges`, role)
		}
	}

	// update the cache if the db operation is successful
	defer func() {
		if err == nil {
			a.roles[role].grant(namespace, convPrivs...)
		}
	}()

	err = grantPrivileges(ctx, db, role, privs, namespace)
	if err != nil {
		return err
	}

	return nil
}

func (a *accessController) RevokePrivileges(ctx context.Context, db sql.DB, role string, privs []string, namespace *string) error {
	if role == ownerRole {
		return fmt.Errorf(`owner role cannot have privileges revoked`)
	}

	perms, ok := a.roles[role]
	if !ok {
		return fmt.Errorf(`role "%s" does not exist`, role)
	}

	// verify that the privileges are valid
	convPrivs, err := validatePrivileges(privs...)
	if err != nil {
		return err
	}

	// if a namespace is provided, check that it exists and that all privileges can be namespaced
	if namespace != nil {
		_, ok := perms.namespacePrivileges[*namespace]
		if !ok {
			return fmt.Errorf(`namespace "%s" does not exist`, *namespace)
		}

		err = canBeNamespaced(convPrivs...)
		if err != nil {
			return err
		}
	}

	for _, p := range convPrivs {
		if !perms.canDo(p, namespace) {
			return fmt.Errorf(`role "%s" does not have some or all of the specified privileges`, role)
		}
	}

	// update the cache if the db operation is successful
	defer func() {
		if err == nil {
			a.roles[role].revoke(namespace, convPrivs...)
		}
	}()

	err = revokePrivileges(ctx, db, role, privs, namespace)
	if err != nil {
		return err
	}

	return nil
}

func (a *accessController) AssignRole(ctx context.Context, db sql.DB, role string, user string) error {
	if isBuiltInRole(role) {
		return fmt.Errorf(`role "%s" is a built-in role and cannot be assigned`, role)
	}

	// check that the role exists
	_, ok := a.roles[role]
	if !ok {
		return fmt.Errorf(`role "%s" does not exist`, role)
	}

	// ensure that the user exists
	_, ok = a.userRoles[user]
	if !ok {
		a.userRoles[user] = []string{}
	}

	// check if the user already has the role
	for _, r := range a.userRoles[user] {
		if r == role {
			return fmt.Errorf(`user "%s" already has role "%s"`, user, role)
		}
	}

	var err error
	// update the cache if the db operation is successful
	defer func() {
		if err == nil {
			a.userRoles[user] = append(a.userRoles[user], role)
		}
	}()

	err = assignRole(ctx, db, role, user)
	if err != nil {
		return err
	}

	return nil
}

func (a *accessController) UnassignRole(ctx context.Context, db sql.DB, role string, user string) error {
	if isBuiltInRole(role) {
		return fmt.Errorf(`role "%s" is a built-in role and cannot be unassigned`, role)
	}

	_, ok := a.roles[role]
	if !ok {
		return fmt.Errorf(`role "%s" does not exist`, role)
	}

	roles, ok := a.userRoles[user]
	if !ok {
		return fmt.Errorf(`user "%s" does not exist`, user)
	}

	// check if the user has the role
	var hasRole bool
	for i, r := range roles {
		if r == role {
			hasRole = true
			// remove the role from the user's roles
			a.userRoles[user] = append(roles[:i], roles[i+1:]...)
			break
		}
	}

	if !hasRole {
		return fmt.Errorf(`user "%s" does not have role "%s"`, user, role)
	}

	err := unassignRole(ctx, db, role, user)
	if err != nil {
		return err
	}

	return nil
}

const ownerKey = "db_owner"

// SetOwnership sets the owner of the database.
// It will overwrite the current owner.
func (a *accessController) SetOwnership(ctx context.Context, db sql.DB, user string) error {
	// update the db
	_, err := db.Execute(ctx, "INSERT INTO kwild_engine.metadata (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2", ownerKey, user)
	if err != nil {
		return err
	}

	// update the cache
	a.owner = user

	return nil
}

func (a *accessController) IsOwner(user string) bool {
	return user == a.owner
}

func (a *accessController) RoleExists(role string) bool {
	_, ok := a.roles[role]
	return ok
}

// createRole creates a role in the db
func createRole(ctx context.Context, db sql.DB, roleName string) error {
	_, err := db.Execute(ctx, "INSERT INTO kwild_engine.roles (name) VALUES ($1)", roleName)
	return err
}

// grantPrivileges grants privileges to a role.
// If the privileges do not exist, it will return an error.
// It can optionally be applied to a specific namespace.
func grantPrivileges(ctx context.Context, db sql.DB, roleName string, privileges []string, namespace *string) error {
	if namespace == nil {
		_, err := db.Execute(ctx, `INSERT INTO kwild_engine.role_privileges (role_id, privilege_type)
		SELECT r.id, unnest($2::text[]) FROM kwild_engine.roles r WHERE r.name = $1`, roleName, privileges)
		return err
	}

	_, err := db.Execute(ctx, `INSERT INTO kwild_engine.namespace_role_privileges (role_id, namespace_id, privilege_type)
	SELECT r.id, n.id, unnest($3::text[]) FROM kwild_engine.roles r
	JOIN kwild_engine.namespaces n ON n.name = $2
	WHERE r.name = $1`, roleName, *namespace, privileges)
	return err
}

// revokePrivileges revokes privileges from a role.
// If the privileges do not exist, it will return an error.
// It can optionally be applied to a specific namespace.
func revokePrivileges(ctx context.Context, db sql.DB, roleName string, privileges []string, namespace *string) error {
	if namespace == nil {
		_, err := db.Execute(ctx, `DELETE FROM kwild_engine.role_privileges
	WHERE role_id = (SELECT id FROM kwild_engine.roles WHERE name = $1) AND privilege_type = ANY($2::text[])`, roleName, privileges)
		return err
	}

	_, err := db.Execute(ctx, `DELETE FROM kwild_engine.namespace_role_privileges
	WHERE role_id = (SELECT id FROM kwild_engine.roles WHERE name = $1)
	AND namespace_id = (SELECT id FROM kwild_engine.namespaces WHERE name = $2)
	AND privilege_type = ANY($3::text[])`, roleName, *namespace, privileges)
	return err
}

// assignRole assigns a role to a user.
// If the role does not exist, it will return an error.
func assignRole(ctx context.Context, db sql.DB, roleName, user string) error {
	_, err := db.Execute(ctx, `INSERT INTO kwild_engine.user_roles (user_id, role_id)
	VALUES ($1, (SELECT id FROM kwild_engine.roles WHERE name = $2))`, user, roleName)
	return err
}

// unassignRole unassigns a role from a user.
// If the role does not exist, it will return an error.
func unassignRole(ctx context.Context, db sql.DB, roleName, user string) error {
	_, err := db.Execute(ctx, `DELETE FROM kwild_engine.user_roles
	WHERE user_id = $1 AND role_id = (SELECT id FROM kwild_engine.roles WHERE name = $2)`, user, roleName)
	return err
}

var privilegeNames = map[privilege]struct{}{
	CallPrivilege:   {},
	SelectPrivilege: {},
	InsertPrivilege: {},
	UpdatePrivilege: {},
	DeletePrivilege: {},
	CreatePrivilege: {},
	DropPrivilege:   {},
	AlterPrivilege:  {},
	RolesPrivilege:  {},
	UsePrivilege:    {},
}

type privilege string

func (p privilege) String() string {
	return string(p)
}

const (
	// Can execute actions
	CallPrivilege privilege = "CALL"
	// can execute ad-hoc select queries
	SelectPrivilege privilege = "SELECT"
	// can insert data
	InsertPrivilege privilege = "INSERT"
	// can update data
	UpdatePrivilege privilege = "UPDATE"
	// can delete data
	DeletePrivilege privilege = "DELETE"
	// can create new objects
	CreatePrivilege privilege = "CREATE"
	// can drop objects
	DropPrivilege privilege = "DROP"
	// use can use extensions
	UsePrivilege privilege = "USE"
	// can alter objects
	AlterPrivilege privilege = "ALTER"
	// can manage roles.
	// roles are global, and are not tied to a specific namespace or object.
	RolesPrivilege privilege = "ROLES"
)

// perms is a struct that holds the permissions for a role.
type perms struct {
	// namespacePrivileges is a map of namespace names to the privileges that are allowed on that namespace.
	// It does NOT include inherited privileges.
	namespacePrivileges map[string]map[privilege]struct{}
	// globalPrivileges is a set of privileges that are allowed globally.
	// it does NOT include inherited privileges.
	globalPrivileges map[privilege]struct{}
}

// canDo returns true if the role can perform the specified action.
func (p *perms) canDo(priv privilege, namespace *string) bool {
	// if the user has the global privilege, return true
	_, hasGlobal := p.globalPrivileges[priv]
	if hasGlobal {
		return true
	}
	// if the user does not have global and no namespace is provided, return false
	if namespace == nil {
		return false
	}

	// otherwise, check the namespace
	np, ok := p.namespacePrivileges[*namespace]
	if !ok {
		return false
	}

	_, has := np[priv]
	return has
}

// grant adds the privileges to the set.
func (p *perms) grant(namespace *string, privs ...privilege) {
	if namespace == nil {
		for _, priv := range privs {
			p.globalPrivileges[priv] = struct{}{}
		}
	} else {
		np, ok := p.namespacePrivileges[*namespace]
		if !ok {
			panic("unexpected error: namespace does not exist")
		}

		for _, priv := range privs {
			np[priv] = struct{}{}
		}

		p.namespacePrivileges[*namespace] = np
	}
}

// revoke removes the privileges from the set.
func (p *perms) revoke(namespace *string, privs ...privilege) {
	if namespace == nil {
		for _, priv := range privs {
			delete(p.globalPrivileges, priv)
		}
	} else {
		np, ok := p.namespacePrivileges[*namespace]
		if !ok {
			panic("unexpected error: namespace does not exist")
		}

		for _, priv := range privs {
			delete(np, priv)
		}

		p.namespacePrivileges[*namespace] = np
	}
}

// canBeNamespaced returns a nil error if the privilege can be namespaced.
func canBeNamespaced(ps ...privilege) error {
	for _, p := range ps {
		if p == RolesPrivilege {
			return fmt.Errorf(`privilege "%s" cannot be namespaced`, p)
		}
	}

	return nil
}

// validatePrivileges returns a nil error if the privileges are valid.
func validatePrivileges(ps ...string) ([]privilege, error) {
	ps2 := make([]privilege, len(ps))
	for i, p := range ps {
		_, ok := privilegeNames[privilege(p)]
		if !ok {
			return nil, fmt.Errorf(`privilege "%s" does not exist`, p)
		}

		ps2[i] = privilege(p)
	}

	return ps2, nil
}
