-- name: CreateQuery :exec
INSERT INTO
    queries (query_name, query, table_id) VALUES
    (
        $1,
        $2,
        (
            SELECT
                id
            FROM
                tables
            WHERE
                table_name = $3
        )
    );