-- name: CreateIndex :exec
INSERT INTO
    INDEXES (table_id, index_name, index_type, COLUMNS)
VALUES
    (
        (
            SELECT
                id
            FROM
                tables
            WHERE
                table_name = $1
        ),
        $2,
        $3,
        $4
    );