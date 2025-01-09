/*
    This section contains all of the DDL for creating the schema for `kwild_engine`, which is
    the internal schema for the engine. This stores all metadata for actions.
*/
CREATE SCHEMA IF NOT EXISTS kwild_engine;

DO $$ 
BEGIN
    -- scalar_data_type is an enumeration of all scalar data types supported by the engine
    BEGIN
        CREATE TYPE kwild_engine.scalar_data_type AS ENUM (
            'INT8', 'TEXT', 'BOOL', 'UUID', 'NUMERIC', 'BYTEA'
        );
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END;

    -- modifiers is an enumeration of all modifiers that can be applied to an action
    BEGIN
        CREATE TYPE kwild_engine.modifiers AS ENUM (
            'VIEW', 'OWNER', 'PUBLIC', 'PRIVATE', 'SYSTEM'
        );
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END;

    -- privilege_type is an enumeration of all privilege types that can be applied to a role
    BEGIN
        CREATE TYPE kwild_engine.privilege_type AS ENUM (
            'SELECT', 'INSERT', 'UPDATE', 'DELETE', 'CREATE', 'DROP', 'ALTER', 'CALL', 'ROLES', 'USE'
        );
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END;

    BEGIN
        CREATE TYPE kwild_engine.namespace_type AS ENUM (
            'USER', 'SYSTEM', 'EXTENSION'
        );
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END;
END $$;

-- namespaces is a table that stores all user schemas in the engine
CREATE TABLE IF NOT EXISTS kwild_engine.namespaces (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    type kwild_engine.namespace_type NOT NULL DEFAULT 'USER'
);


CREATE TABLE IF NOT EXISTS kwild_engine.initialized_extensions (
    id BIGSERIAL PRIMARY KEY,
    namespace_id INT8 NOT NULL REFERENCES kwild_engine.namespaces(id) ON UPDATE CASCADE ON DELETE CASCADE,
    base_extension TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS kwild_engine.extension_initialization_parameters (
    id BIGSERIAL PRIMARY KEY,
    extension_id INT8 NOT NULL REFERENCES kwild_engine.initialized_extensions(id) ON UPDATE CASCADE ON DELETE CASCADE,
    key TEXT NOT NULL,
    value TEXT NOT NULL,
    data_type TEXT NOT NULL,
    UNIQUE (extension_id, key)
);

-- actions is a table that stores all actions in the engine
CREATE TABLE IF NOT EXISTS kwild_engine.actions (
    id BIGSERIAL PRIMARY KEY,
    namespace TEXT NOT NULL REFERENCES kwild_engine.namespaces(name) ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (name = lower(name)),
    raw_statement TEXT NOT NULL,
    returns_table BOOLEAN NOT NULL DEFAULT FALSE,
    modifiers kwild_engine.modifiers[],
    UNIQUE (namespace, name)
);

-- parameters is a table that stores all parameters for actions in the engine
CREATE TABLE IF NOT EXISTS kwild_engine.parameters (
    id BIGSERIAL PRIMARY KEY,
    action_id INT8 NOT NULL REFERENCES kwild_engine.actions(id) ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (name = lower(name)),
    position INT8 NOT NULL,
    scalar_type kwild_engine.scalar_data_type NOT NULL,
    is_array BOOLEAN NOT NULL,
    metadata BYTEA DEFAULT NULL
);

-- return_types is a table that stores all return types for actions in the engine
CREATE TABLE IF NOT EXISTS kwild_engine.return_fields (
    id BIGSERIAL PRIMARY KEY,
    action_id INT8 NOT NULL REFERENCES kwild_engine.actions(id) ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (name = lower(name)),
    position INT8 NOT NULL,
    scalar_type kwild_engine.scalar_data_type NOT NULL,
    is_array BOOLEAN NOT NULL,
    metadata BYTEA DEFAULT NULL
);

-- roles_table is a table that stores all role information.
-- since Kwil uses it's own roles system that is in no way related to the Postgres roles system, we need to store this information
CREATE TABLE IF NOT EXISTS kwild_engine.roles (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS kwild_engine.role_privileges (
    id BIGSERIAL PRIMARY KEY,
    privilege_type kwild_engine.privilege_type NOT NULL,
    namespace_id INT8 REFERENCES kwild_engine.namespaces(id) ON UPDATE CASCADE ON DELETE CASCADE, -- the namespace it is targeting. Can be null if it is a global privilege
    role_id INT8 NOT NULL REFERENCES kwild_engine.roles(id) ON UPDATE CASCADE ON DELETE CASCADE
);

-- user_roles is a table that stores all users who have been assigned roles
CREATE TABLE IF NOT EXISTS kwild_engine.user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_identifier TEXT NOT NULL,
    role_id INT8 NOT NULL
);

-- an index here helps with performance when querying for a user's roles
CREATE INDEX IF NOT EXISTS user_roles_user_identifier_idx ON kwild_engine.user_roles(user_identifier);

-- create a single default role that will be used for all users
INSERT INTO kwild_engine.roles (name) VALUES ('default') ON CONFLICT DO NOTHING;
-- default role can select and call by default
INSERT INTO kwild_engine.role_privileges (privilege_type, role_id) VALUES ('SELECT', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'default'
)), ('CALL', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'default'
)) ON CONFLICT DO NOTHING;

-- create an owner role, which has all privileges
INSERT INTO kwild_engine.roles (name) VALUES ('owner') ON CONFLICT DO NOTHING;

-- owner role can do everything
INSERT INTO kwild_engine.role_privileges (privilege_type, role_id) VALUES ('SELECT', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'owner'
)), ('INSERT', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'owner'
)), ('UPDATE', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'owner'
)), ('DELETE', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'owner'
)), ('CREATE', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'owner'
)), ('DROP', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'owner'
)), ('ALTER', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'owner'
)), ('CALL', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'owner'
)), ('ROLES', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'owner'
)), ('USE', (
    SELECT id
    FROM kwild_engine.roles
    WHERE name = 'owner'
)) ON CONFLICT DO NOTHING;


-- format_type is a function that formats a data type for display
CREATE OR REPLACE FUNCTION kwild_engine.format_type(scal kwild_engine.scalar_data_type, is_arr BOOLEAN, meta BYTEA)
RETURNS TEXT AS $$
DECLARE
    result TEXT;
BEGIN
    result := lower(scal::text);

    if result = 'numeric' THEN
        if octet_length(meta) = 4 THEN
            -- precision and scale are uint16, precision is first 2 bytes, scale is next 2 bytes
            result := result || '(' ||
                ((get_byte(meta, 0) << 8 | get_byte(meta, 1))::TEXT) || ',' ||
                ((get_byte(meta, 2) << 8 | get_byte(meta, 3))::TEXT) || ')';
        ELSIF octet_length(meta) != 0 THEN
            -- should never happen, would suggest some sort of serious internal error
            RAISE EXCEPTION 'Invalid metadata length for numeric data type';
        END IF;
    ELSIF octet_length(meta) != 0 THEN
        -- should never happen, would suggest some sort of serious internal error
        RAISE EXCEPTION 'Invalid metadata length for non-numeric data type';
    END IF;

    if is_arr THEN
        result := result || '[]';
    END IF;

    RETURN result;
END;
$$ LANGUAGE plpgsql;

/*
    This section creates the schema the `kwild` schema, which is the public user-facing schema.
    End users can access the views in this schema to get information about the database.

    All views are ordered to ensure that they are deterministic when queried.
*/
CREATE SCHEMA IF NOT EXISTS info;
INSERT INTO kwild_engine.namespaces (name, type) VALUES ('info', 'SYSTEM') ON CONFLICT DO NOTHING;

-- info.namespaces is a public view that provides a list of all namespaces in the database
CREATE VIEW info.namespaces AS
SELECT 
    name,
    type::TEXT
FROM
    kwild_engine.namespaces
ORDER BY
    name;

-- info.tables is a public view that provides a list of all tables in the database
CREATE VIEW info.tables AS
SELECT
    t.tablename::text   AS name,
    t.schemaname::text  AS namespace
FROM pg_tables t
JOIN kwild_engine.namespaces us
    ON t.schemaname = us.name

UNION ALL

SELECT
    v.viewname::text    AS name,
    v.schemaname::text  AS namespace
FROM pg_views v
JOIN kwild_engine.namespaces us
    ON v.schemaname = us.name

ORDER BY 1, 2;

-- info.columns is a public view that provides a list of all columns in the database
CREATE VIEW info.columns AS
SELECT 
    c.table_schema::TEXT AS namespace,
    c.table_name::TEXT AS table_name,
    c.column_name::TEXT AS name,
    (CASE 
        WHEN t.typcategory = 'A'
            THEN (pg_catalog.format_type(a.atttypid, a.atttypmod))
        ELSE c.data_type
    END)::TEXT AS data_type,
    c.is_nullable::bool AS is_nullable,
    c.column_default::TEXT AS default_value,
    CASE
        WHEN tc.constraint_type = 'PRIMARY KEY'
            THEN true
        ELSE false
    END AS is_primary_key,
    c.ordinal_position AS ordinal_position
FROM information_schema.columns c
JOIN pg_namespace n
    ON c.table_schema = n.nspname::TEXT
JOIN pg_class cl
    ON cl.relname = c.table_name
        AND cl.relnamespace = n.oid
JOIN pg_attribute a
    ON a.attname = c.column_name 
        AND a.attrelid = cl.oid
JOIN pg_type t
    ON t.oid = a.atttypid
LEFT JOIN information_schema.key_column_usage kcu
    ON c.table_name = kcu.table_name 
        AND c.column_name = kcu.column_name
        AND c.table_schema = kcu.table_schema
LEFT JOIN information_schema.table_constraints tc
    ON kcu.constraint_name = tc.constraint_name
        AND tc.constraint_type = 'PRIMARY KEY'
        AND tc.table_schema = c.table_schema
JOIN 
    kwild_engine.namespaces us ON n.nspname::TEXT = us.name
WHERE cl.relkind IN ('r', 'v') -- only tables and views
ORDER BY 
    c.table_name, 
    c.ordinal_position,
    1, 2, 3, 4, 5, 6, 7, 8;

-- info.indexes is a public view that provides a list of all indexes in the database
CREATE VIEW info.indexes AS
SELECT 
    n.nspname::TEXT AS namespace,
    c.relname::TEXT AS table_name,
    ic.relname::TEXT AS name,
    i.indisprimary AS is_pk,
    i.indisunique AS is_unique,
    array_agg(a.attname ORDER BY x.ordinality)::TEXT[] AS column_names
FROM pg_index i
JOIN pg_class c ON c.oid = i.indrelid
JOIN pg_class ic ON ic.oid = i.indexrelid
JOIN pg_namespace n ON c.relnamespace = n.oid
JOIN pg_am am ON ic.relam = am.oid
JOIN pg_attribute a ON a.attnum = ANY(i.indkey) AND a.attrelid = c.oid
JOIN LATERAL unnest(i.indkey) WITH ORDINALITY AS x(colnum, ordinality) ON x.colnum = a.attnum
JOIN 
    kwild_engine.namespaces us ON n.nspname::TEXT = us.name
GROUP BY n.nspname, c.relname, ic.relname, i.indisprimary, i.indisunique
ORDER BY 1,2,3,4,5,6;

-- info.constraints is a public view that provides a list of all constraints in the database
CREATE VIEW info.constraints AS
SELECT 
    pg_namespace.nspname::TEXT AS namespace,
    conname::TEXT AS constraint_name,
    split_part(conrelid::regclass::text, '.', 2) AS table_name,
    array_agg(attname)::TEXT[] AS columns,
    pg_get_constraintdef(pg_constraint.oid) AS expression,
    CASE contype
        WHEN 'c' THEN 'CHECK'
        WHEN 'u' THEN 'UNIQUE'
    END AS constraint_type
FROM 
    pg_constraint
JOIN 
    pg_class ON conrelid = pg_class.oid
JOIN 
    pg_namespace ON pg_class.relnamespace = pg_namespace.oid
LEFT JOIN 
    unnest(conkey) AS cols(colnum) ON true
LEFT JOIN 
    pg_attribute ON pg_attribute.attnum = cols.colnum AND pg_attribute.attrelid = pg_class.oid
JOIN 
    kwild_engine.namespaces us ON pg_namespace.nspname::TEXT = us.name
WHERE 
    contype = 'c'  -- Only check constraints
    OR contype = 'u'  -- Only unique constraints 
GROUP BY 
    pg_namespace.nspname, conname, conrelid, pg_constraint.oid
ORDER BY 
    1, 2, 3, 4, 5, 6;

-- info.foreign_keys is a public view that provides a list of all foreign keys in the database
CREATE VIEW info.foreign_keys AS
SELECT 
    pg_namespace.nspname::TEXT AS namespace,
    conname::TEXT AS constraint_name,
    split_part(conrelid::regclass::text, '.', 2) AS table_name,
    array_agg(attname)::TEXT[] AS columns,
    CASE confupdtype
        WHEN 'a' THEN 'NO ACTION'
        WHEN 'r' THEN 'RESTRICT'
        WHEN 'c' THEN 'CASCADE'
        WHEN 'n' THEN 'SET NULL'
        WHEN 'd' THEN 'SET DEFAULT'
    END AS on_update,
    CASE confdeltype
        WHEN 'a' THEN 'NO ACTION'
        WHEN 'r' THEN 'RESTRICT'
        WHEN 'c' THEN 'CASCADE'
        WHEN 'n' THEN 'SET NULL'
        WHEN 'd' THEN 'SET DEFAULT'
    END AS on_delete
FROM 
    pg_constraint
JOIN 
    pg_class ON conrelid = pg_class.oid
JOIN 
    pg_namespace ON pg_class.relnamespace = pg_namespace.oid
LEFT JOIN 
    unnest(conkey) AS cols(colnum) ON true
LEFT JOIN 
    pg_attribute ON pg_attribute.attnum = cols.colnum AND pg_attribute.attrelid = pg_class.oid
JOIN 
    kwild_engine.namespaces us ON pg_namespace.nspname::TEXT = us.name
WHERE 
    contype = 'f'  -- Only foreign key constraints
GROUP BY 
    pg_namespace.nspname, conname, conrelid, confupdtype, confdeltype
ORDER BY 
    table_name, constraint_name,
    1, 2, 3, 4, 5, 6;


-- actions is a public view that provides a list of all actions in the database
CREATE VIEW info.actions AS
SELECT 
    a.namespace AS namespace,
    a.name::TEXT AS name,
    a.raw_statement,
    a.modifiers::TEXT[] AS modifiers,
    a.returns_table,
    array_agg(
        json_build_object(
            'name', p.name,
            'data_type', kwild_engine.format_type(p.scalar_type, p.is_array, p.metadata)
        )::TEXT
        ORDER BY p.position, p.name, kwild_engine.format_type(p.scalar_type, p.is_array, p.metadata)
    ) AS parameters,
    array_agg(
        json_build_object(
            'name', r.name,
            'data_type', kwild_engine.format_type(r.scalar_type, r.is_array, r.metadata)
        )::TEXT
        ORDER BY r.position, r.name, kwild_engine.format_type(r.scalar_type, r.is_array, r.metadata)
    ) AS return_types
FROM kwild_engine.actions a
LEFT JOIN kwild_engine.parameters p
    ON a.id = p.action_id
LEFT JOIN kwild_engine.return_fields r
    ON a.id = r.action_id
GROUP BY a.namespace, a.name, a.modifiers, a.raw_statement, a.returns_table
ORDER BY a.name,
    1, 2, 3, 4, 5
;

-- roles is a public view that provides a list of all roles in the database
CREATE VIEW info.roles AS
SELECT 
    name
FROM
    kwild_engine.roles
ORDER BY
    name;

CREATE VIEW info.user_roles AS
SELECT 
    user_identifier,
    r.name AS role
FROM
    kwild_engine.user_roles ur
JOIN
    kwild_engine.roles r
    ON ur.role_id = r.id
ORDER BY
    1, 2;

-- role_privileges is a public view that provides a list of all role privileges in the database
CREATE VIEW info.role_privileges AS
SELECT 
    r.name AS role,
    p.privilege_type::text AS privilege,
    n.name AS namespace
FROM
    kwild_engine.role_privileges p
JOIN
    kwild_engine.roles r
    ON p.role_id = r.id
LEFT JOIN
    kwild_engine.namespaces n
    ON p.namespace_id = n.id
ORDER BY
    1, 2, 3;

CREATE VIEW info.extensions AS
SELECT 
    n.name AS namespace,
    ie.base_extension AS extension,
    array_agg(
        json_build_object(
            'key', eip.key,
            'value', eip.value
        )::TEXT
        ORDER BY eip.key, eip.value
    ) AS parameters
FROM
    kwild_engine.initialized_extensions ie
JOIN
    kwild_engine.namespaces n
    ON ie.namespace_id = n.id
LEFT JOIN
    kwild_engine.extension_initialization_parameters eip
    ON ie.id = eip.extension_id
GROUP BY
    n.name, ie.base_extension
ORDER BY
    1, 2;

-- lastly, we need to create a default namespace for the user
CREATE SCHEMA IF NOT EXISTS main;
INSERT INTO kwild_engine.namespaces (name, type) VALUES ('main', 'SYSTEM') ON CONFLICT DO NOTHING;