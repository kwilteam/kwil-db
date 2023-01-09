-- name: CreateRole :exec
INSERT INTO
    roles (role_name, db_id, is_default)
VALUES
    (
        $1,
        (
            SELECT
                id
            FROM
                databases
            WHERE
                db_name = $2
        ),
        $3
    );

-- name: RoleApplyAccount :exec
INSERT INTO
    role_accounts (role_id, account_id)
VALUES
    (
        (
            SELECT
                id
            FROM
                roles
            WHERE
                role_name = $1
        ),
        (
            SELECT
                id
            FROM
                accounts
            WHERE
                account_address = $2
        )
    );

-- name: RoleApplyQuery :exec
INSERT INTO
    role_queries (role_id, query_id)
VALUES
    (
        (
            SELECT
                id
            FROM
                roles
            WHERE
                role_name = $1
        ),
        (
            SELECT
                id
            FROM
                queries
            WHERE
                query_name = $2
        )
    );