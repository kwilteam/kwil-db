-- name: CreateRole :exec
INSERT INTO
    roles (role_name, db_id, is_default)
VALUES
    (
        $1,
        $2,
        $3
    );

-- name: ApplyPermissionToRole :exec
INSERT INTO
    role_queries (role_id, query_id)
VALUES
    (
        (
            SELECT
                id
            FROM
                roles r
            WHERE
                r.role_name = $2
                AND r.db_id = $1
        ),
        (
            SELECT
                id
            FROM
                queries q
            WHERE
                q.query_name = $3
                AND q.db_id = $1
        )
    );

-- name: GetRoles :many
SELECT
    role_name,
    id,
    is_default
FROM
    roles
WHERE
    db_id = $1;

-- name: GetRolePermissions :many
SELECT
    query_name
FROM
    queries
    JOIN role_queries ON queries.id = role_queries.query_id
WHERE
    role_queries.role_id = $1;