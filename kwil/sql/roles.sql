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

-- name: RoleApplyWallet :exec
INSERT INTO
    role_wallets (role_id, wallet_id)
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
                wallets
            WHERE
                wallet = $2
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