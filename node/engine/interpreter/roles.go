package interpreter

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/kwilteam/kwil-db/node/engine"
	"github.com/kwilteam/kwil-db/node/types/sql"
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
		roles:           make(map[string]*perms),
		userRoles:       make(map[string][]string),
		knownNamespaces: make(map[string]struct{}),
	}

	getRolesStmt := `SELECT r.name, array_agg(rp.privilege_type::text), array_agg(n.name)
	FROM kwild_engine.roles r
	LEFT JOIN kwild_engine.role_privileges rp ON rp.role_id = r.id
	LEFT JOIN kwild_engine.namespaces n ON rp.namespace_id = n.id
	GROUP BY r.name
	ORDER BY 1,2,3`

	// list all roles, their perms, and users
	var roleName string
	var privileges []*string
	var namespaces []*string
	err := queryRowFunc(ctx, db, getRolesStmt, []any{&roleName, &privileges, &namespaces}, func() error {
		perm := &perms{
			namespacePrivileges: make(map[string]map[privilege]struct{}),
			globalPrivileges:    make(map[privilege]struct{}),
		}

		if len(privileges) != len(namespaces) {
			return fmt.Errorf(`unexpected error: length of privileges and namespaces do not match. this is an internal bug`)
		}

		for i, priv := range privileges {
			if priv == nil {
				// priv can be nil if the role has no privileges
				if len(namespaces) != 1 {
					return fmt.Errorf(`unexpected error: nil privilege in non-nil list of privileges. this is an internal bug`)
				}
				if namespaces[i] != nil {
					return fmt.Errorf(`unexpected error: nil privilege in non-nil list of namespaces. this is an internal bug`)
				}
				continue
			}

			// check that the privilege exists
			// This should never not be the case, but it is good to check
			_, ok := privilegeNames[privilege(*priv)]
			if !ok {
				return fmt.Errorf(`unknown privilege "%s" stored in DB`, *priv)
			}

			// if namespace is nil, then it is a global privilege
			// We still register all global privileges with each namespace
			if namespaces[i] == nil {
				perm.globalPrivileges[privilege(*priv)] = struct{}{}

				for nsPriv, np := range perm.namespacePrivileges {
					np[privilege(*priv)] = struct{}{}

					perm.namespacePrivileges[nsPriv] = np
				}
			} else {
				if _, ok := perm.namespacePrivileges[*namespaces[i]]; !ok {
					perm.namespacePrivileges[*namespaces[i]] = make(map[privilege]struct{})
				}

				perm.namespacePrivileges[*namespaces[i]][privilege(*priv)] = struct{}{}
			}

		}

		ac.roles[roleName] = perm

		return nil
	})
	if err != nil {
		return nil, err
	}

	// get all users and their roles
	getUsersStmt := `SELECT u.user_identifier, array_agg(r.name)
	FROM kwild_engine.user_roles u
	JOIN kwild_engine.roles r ON r.id = u.role_id
	GROUP BY u.user_identifier
	ORDER BY 1, 2`

	var user string
	var roles []*string
	err = queryRowFunc(ctx, db, getUsersStmt, []any{&user, &roles}, func() error {
		for _, role := range roles {
			if role == nil {
				panic("unexpected error: role is nil")
			}

			ac.userRoles[user] = append(ac.userRoles[user], *role)
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
	roles           map[string]*perms
	userRoles       map[string][]string // a map of user public keys to the roles they have. It does _not_ include the default role.
	knownNamespaces map[string]struct{} // a set of all known namespaces
}

func (a *accessController) copy() *accessController {
	a2 := &accessController{
		roles:           make(map[string]*perms, len(a.roles)),
		userRoles:       make(map[string][]string, len(a.userRoles)),
		knownNamespaces: maps.Clone(a.knownNamespaces),
	}

	for k, v := range a.roles {
		a2.roles[k] = v.copy()
	}

	for k, v := range a.userRoles {
		a2.userRoles[k] = slices.Clone(v)
	}

	return a2
}

// CreateRole adds a new role to the access controller.
func (a *accessController) CreateRole(ctx context.Context, db sql.DB, role string) error {
	if isBuiltInRole(role) {
		return fmt.Errorf(`%w: role "%s" cannot be added`, engine.ErrBuiltInRole, role)
	}

	_, ok := a.roles[role]
	if ok {
		return fmt.Errorf(`role "%s" already exists`, role)
	}

	err := createRole(ctx, db, role)
	if err != nil {
		return err
	}

	p := &perms{
		namespacePrivileges: make(map[string]map[privilege]struct{}),
		globalPrivileges:    make(map[privilege]struct{}),
	}
	for ns := range a.knownNamespaces {
		p.namespacePrivileges[ns] = make(map[privilege]struct{})
	}

	a.roles[role] = p

	return nil
}

func (a *accessController) DeleteRole(ctx context.Context, db sql.DB, role string) error {
	if isBuiltInRole(role) {
		return fmt.Errorf(`%w: role "%s" cannot be dropped`, engine.ErrBuiltInRole, role)
	}

	_, ok := a.roles[role]
	if !ok {
		return fmt.Errorf(`role "%s" does not exist`, role)
	}

	// remove the role from the db
	err := execute(ctx, db, "DELETE FROM kwild_engine.roles WHERE name = $1", role)
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

// registerNamespace creates a new namespace.
// It does not modify any storage; it only updates the cache.
func (a *accessController) registerNamespace(namespace string) {
	for _, perm := range a.roles {
		perm.namespacePrivileges[namespace] = make(map[privilege]struct{})

		for priv := range perm.globalPrivileges {
			perm.namespacePrivileges[namespace][priv] = struct{}{}
		}
	}
	a.knownNamespaces[namespace] = struct{}{}
}

// unregisterNamespace deletes all roles and privileges associated with a namespace.
// It does not modify any storage; it only updates the cache.
func (a *accessController) unregisterNamespace(namespace string) {
	for _, role := range a.roles {
		delete(role.namespacePrivileges, namespace)
	}
	delete(a.knownNamespaces, namespace)
}

func (a *accessController) HasPrivilege(user string, namespace *string, privilege privilege) bool {
	// if it is the owner, they have all privileges
	if a.IsOwner(user) {
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
			panic("Unexpected cache error: role does not exist. This is a bug.")
		}

		if perms.canDo(privilege, namespace) {
			return true
		}
	}

	return false
}

func (a *accessController) GrantPrivileges(ctx context.Context, db sql.DB, role string, privs []privilege, namespace *string, ifNotGranted bool) error {
	if role == ownerRole {
		return fmt.Errorf(`owner role already has all privileges`)
	}

	perms, ok := a.roles[role]
	if !ok {
		return fmt.Errorf(`role "%s" does not exist`, role)
	}

	ungrantedPrivs := make([]privilege, 0, len(privs))
	for _, p := range privs {
		if perms.canDo(p, namespace) {
			if ifNotGranted {
				ungrantedPrivs = append(ungrantedPrivs, p)
			} else {
				return fmt.Errorf(`role "%s" already has some or all of the specified privileges`, role)
			}
		} else {
			ungrantedPrivs = append(ungrantedPrivs, p)
		}
	}

	var err error

	// update the cache if the db operation is successful
	defer func() {
		if err == nil {
			a.roles[role].grant(namespace, ungrantedPrivs...)
		}
	}()

	err = grantPrivilegesSQL(ctx, db, role, ungrantedPrivs, namespace)
	if err != nil {
		return err
	}

	return nil
}

func (a *accessController) RevokePrivileges(ctx context.Context, db sql.DB, role string, privs []privilege, namespace *string, ifGranted bool) error {
	if role == ownerRole {
		return fmt.Errorf(`owner role cannot have privileges revoked`)
	}

	perms, ok := a.roles[role]
	if !ok {
		return fmt.Errorf(`role "%s" does not exist`, role)
	}

	// for each incoming privilege, if the role can already not do it, then
	// we error out.
	for _, p := range privs {
		if !perms.canDo(p, namespace) {
			if ifGranted {
				continue
			} else {
				return fmt.Errorf(`role "%s" does not have some or all of the specified privileges`, role)
			}
		}
	}

	var err error

	// update the cache if the db operation is successful
	defer func() {
		if err == nil {
			a.roles[role].revoke(namespace, privs...)
		}
	}()

	err = revokePrivilegesSQL(ctx, db, role, privs, namespace)
	if err != nil {
		return err
	}

	return nil
}

func (a *accessController) AssignRole(ctx context.Context, db sql.DB, role string, user string, ifNotGranted bool) error {
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
			if ifNotGranted {
				return nil
			}
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

func (a *accessController) UnassignRole(ctx context.Context, db sql.DB, role string, user string, ifGranted bool) error {
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
		if ifGranted {
			return nil
		}
		return fmt.Errorf(`user "%s" does not have role "%s"`, user, role)
	}

	err := unassignRole(ctx, db, role, user)
	if err != nil {
		return err
	}

	return nil
}

func (a *accessController) IsOwner(user string) bool {
	roles, ok := a.userRoles[user]
	if !ok {
		return false
	}

	for _, role := range roles {
		if role == ownerRole {
			return true
		}
	}

	return false
}

// GetOwner returns the owner of the database.
func (a *accessController) GetOwner() (owner string, found bool) {
	i := 0
	for user, roles := range a.userRoles {
		for _, role := range roles {
			if role == ownerRole {
				owner = user
				found = true
				i++
			}
		}
	}

	if i > 1 {
		// suggests a bug in the engine
		panic("unexpected error: multiple owners found")
	}

	return
}

func (a *accessController) RoleExists(role string) bool {
	_, ok := a.roles[role]
	return ok
}

// createRole creates a role in the db
func createRole(ctx context.Context, db sql.DB, roleName string) error {
	err := execute(ctx, db, "INSERT INTO kwild_engine.roles (name) VALUES ($1)", roleName)
	return err
}

// grantPrivilegesSQL grants privileges to a role.
// If the privileges do not exist, it will return an error.
// It can optionally be applied to a specific namespace.
func grantPrivilegesSQL(ctx context.Context, db sql.DB, roleName string, privileges []privilege, namespace *string) error {
	// we need to convert the privileges back to strings so that pgx can find an encode plan
	privStrs := make([]string, len(privileges))
	for i, p := range privileges {
		privStrs[i] = string(p)
	}

	if namespace == nil {
		err := execute(ctx, db, `INSERT INTO kwild_engine.role_privileges (role_id, privilege_type)
		SELECT r.id, unnest($2::text[])::kwild_engine.privilege_type FROM kwild_engine.roles r WHERE r.name = $1`, roleName, privStrs)
		return err
	}

	err := execute(ctx, db, `INSERT INTO kwild_engine.role_privileges (role_id, namespace_id, privilege_type)
	SELECT r.id, n.id, unnest($3::text[])::kwild_engine.privilege_type FROM kwild_engine.roles r
	JOIN kwild_engine.namespaces n ON n.name = $2
	WHERE r.name = $1`, roleName, *namespace, privStrs)
	return err
}

// revokePrivilegesSQL revokes privileges from a role.
// If the privileges do not exist, it will return an error.
// It can optionally be applied to a specific namespace.
func revokePrivilegesSQL(ctx context.Context, db sql.DB, roleName string, privileges []privilege, namespace *string) error {
	// we need to convert the privileges back to strings so that pgx can find an encode plan
	privStrs := make([]string, len(privileges))
	for i, p := range privileges {
		privStrs[i] = string(p)
	}

	if namespace == nil {
		err := execute(ctx, db, `DELETE FROM kwild_engine.role_privileges
	WHERE role_id = (SELECT id FROM kwild_engine.roles WHERE name = $1) AND privilege_type::text = ANY($2::text[])`, roleName, privStrs)
		return err
	}

	err := execute(ctx, db, `DELETE FROM kwild_engine.role_privileges
	WHERE role_id = (SELECT id FROM kwild_engine.roles WHERE name = $1)
	AND namespace_id = (SELECT id FROM kwild_engine.namespaces WHERE name = $2)
	AND privilege_type::text = ANY($3::text[])`, roleName, *namespace, privStrs)
	return err
}

// assignRole assigns a role to a user.
// If the role does not exist, it will return an error.
func assignRole(ctx context.Context, db sql.DB, roleName, user string) error {
	err := execute(ctx, db, `INSERT INTO kwild_engine.user_roles (user_identifier, role_id)
	VALUES ($1, (SELECT id FROM kwild_engine.roles WHERE name = $2))`, user, roleName)
	return err
}

// unassignRole unassigns a role from a user.
// If the role does not exist, it will return an error.
func unassignRole(ctx context.Context, db sql.DB, roleName, user string) error {
	err := execute(ctx, db, `DELETE FROM kwild_engine.user_roles
	WHERE user_identifier = $1 AND role_id = (SELECT id FROM kwild_engine.roles WHERE name = $2)`, user, roleName)
	return err
}

var privilegeNames = map[privilege]struct{}{
	_CALL_PRIVILEGE:   {},
	_SELECT_PRIVILEGE: {},
	_INSERT_PRIVILEGE: {},
	_UPDATE_PRIVILEGE: {},
	_DELETE_PRIVILEGE: {},
	_CREATE_PRIVILEGE: {},
	_DROP_PRIVILEGE:   {},
	_ALTER_PRIVILEGE:  {},
	_ROLES_PRIVILEGE:  {},
	_USE_PRIVILEGE:    {},
}

type privilege string

func (p privilege) String() string {
	return string(p)
}

// the following constants all start with _ to unexport them.
const (
	// Can execute actions
	_CALL_PRIVILEGE privilege = "CALL"
	// can execute ad-hoc select queries
	_SELECT_PRIVILEGE privilege = "SELECT"
	// can insert data
	_INSERT_PRIVILEGE privilege = "INSERT"
	// can update data
	_UPDATE_PRIVILEGE privilege = "UPDATE"
	// can delete data
	_DELETE_PRIVILEGE privilege = "DELETE"
	// can create new objects
	_CREATE_PRIVILEGE privilege = "CREATE"
	// can drop objects
	_DROP_PRIVILEGE privilege = "DROP"
	// use can use extensions
	_USE_PRIVILEGE privilege = "USE"
	// can alter objects
	_ALTER_PRIVILEGE privilege = "ALTER"
	// can manage roles.
	// roles are global, and are not tied to a specific namespace or object.
	_ROLES_PRIVILEGE privilege = "ROLES"
)

// perms is a struct that holds the permissions for a role.
type perms struct {
	// namespacePrivileges is a map of namespace names to the privileges that are allowed on that namespace.
	// It does NOT include inherited privileges.
	namespacePrivileges map[string]map[privilege]struct{}
	// globalPrivileges is a set of privileges that are allowed globally.
	// it does NOT include inherited privileges.
	// This map should NOT be used to check if a user has a privilege.
	// Instead, it is used when a new namespace is created, so that
	// the new namespace (within namespacePrivileges) can inherit the global privileges.
	// This is because a global privilege can later be revoked for a certain namespace.
	globalPrivileges map[privilege]struct{}
}

func (p *perms) copy() *perms {
	p2 := &perms{
		namespacePrivileges: make(map[string]map[privilege]struct{}),
		globalPrivileges:    maps.Clone(p.globalPrivileges),
	}

	for k, v := range p.namespacePrivileges {
		p2.namespacePrivileges[k] = maps.Clone(v)
	}

	return p2
}

// canDo returns true if the role can perform the specified action.
func (p *perms) canDo(priv privilege, namespace *string) bool {
	if namespace == nil {
		// if no namespace is provided, then this must be something
		// that is not namespaceable. In that case, check if they have
		// the global privilege
		_, hasGlobal := p.globalPrivileges[priv]
		return hasGlobal
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

		for _, np := range p.namespacePrivileges {
			for _, priv := range privs {
				np[priv] = struct{}{}
			}
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

		for _, np := range p.namespacePrivileges {
			for _, priv := range privs {
				delete(np, priv)
			}
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
		switch p {
		case _ROLES_PRIVILEGE, _USE_PRIVILEGE:
			return fmt.Errorf(`%w: %s`, engine.ErrCannotBeNamespaced, p)
		}
	}

	return nil
}

// validatePrivileges returns a nil error if the privileges are valid.
func validatePrivileges(ps ...string) ([]privilege, error) {
	ps2 := make([]privilege, len(ps))
	for i, p := range ps {
		p = strings.ToUpper(p)
		_, ok := privilegeNames[privilege(p)]
		if !ok {
			return nil, fmt.Errorf(`privilege "%s" does not exist`, p)
		}

		ps2[i] = privilege(p)
	}

	return ps2, nil
}
