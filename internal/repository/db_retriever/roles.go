package dbretriever

import (
	"context"
	"fmt"
	"kwil/pkg/types/databases"
)

func (q *dbRetriever) GetRoles(ctx context.Context, dbid int32) ([]*databases.Role, error) {
	roleList, err := q.gen.GetRoles(ctx, dbid)
	if err != nil {
		return nil, fmt.Errorf(`error getting roles for dbid %d: %w`, dbid, err)
	}

	// sqlc can't return array_aggs so we have to get the permissions separately

	roles := make([]*databases.Role, len(roleList))
	for i, role := range roleList {
		perms, err := q.gen.GetRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf(`error getting permissions for role %s: %w`, role.RoleName, err)
		}

		roles[i] = &databases.Role{
			Name:        role.RoleName,
			Default:     role.IsDefault,
			Permissions: perms,
		}
	}

	return roles, nil
}
