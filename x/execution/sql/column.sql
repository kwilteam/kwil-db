-- name: CreateColumn :exec
INSERT INTO
    COLUMNS (table_id, column_name, column_type)
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
        $3
    );

-- name: CreateAttribute :exec
INSERT INTO
    attributes (column_id, attribute_type, attribute_value)
VALUES
    (
        (
            SELECT
                id
            FROM
                COLUMNS
            WHERE
                column_name = $1
        ),
        $2,
        $3
    );