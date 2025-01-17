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
    scalar_type kwild_engine.scalar_data_type NOT NULL,
    is_array BOOLEAN NOT NULL,
    metadata BYTEA DEFAULT NULL,
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
    built_in BOOLEAN DEFAULT FALSE,
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
    name TEXT NOT NULL UNIQUE,
    built_in BOOLEAN DEFAULT FALSE
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
INSERT INTO kwild_engine.roles (name, built_in) VALUES ('default', true) ON CONFLICT DO NOTHING;
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
INSERT INTO kwild_engine.roles (name, built_in) VALUES ('owner', true) ON CONFLICT DO NOTHING;

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
        ELSE
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

-- format_pg_type formats a function read from postgres's information_schema.columns
CREATE OR REPLACE FUNCTION kwild_engine.format_pg_type (type oid, typemod integer)
RETURNS TEXT AS $$
DECLARE
    result TEXT;
BEGIN
    result := pg_catalog.format_type(type, typemod);
    -- we can usually just return this, however there are a few times that we need to format it
    -- to Kwil's native type
    if result = 'character varying' THEN
        result := 'text';
    END IF;
    if result = 'bigint' THEN
        result := 'int8';
    END IF;
    if result = 'character' THEN
        result := 'text';
    END IF;
    if result = 'decimal' THEN
        result := 'numeric';
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
CREATE OR REPLACE VIEW info.columns AS
SELECT
    c.table_schema::text          AS namespace,
    c.table_name::text            AS table_name,
    c.column_name::text           AS name,
    kwild_engine.format_pg_type(a.atttypid, a.atttypmod)::text  AS data_type,
    c.is_nullable::bool           AS is_nullable,
    c.column_default::text        AS default_value,
    -- Instead of joining to table_constraints, do a subselect:
    CASE
        WHEN EXISTS (
            SELECT 1
            FROM information_schema.key_column_usage kc
            JOIN information_schema.table_constraints tc
                 ON kc.constraint_name = tc.constraint_name
                AND kc.table_schema = tc.table_schema
            WHERE
                kc.table_schema = c.table_schema
                AND kc.table_name  = c.table_name
                AND kc.column_name = c.column_name
                AND tc.constraint_type = 'PRIMARY KEY'
        ) THEN true
        ELSE false
    END AS is_primary_key,
    c.ordinal_position::int       AS ordinal_position
FROM information_schema.columns c
JOIN pg_namespace n
    ON c.table_schema = n.nspname::text
JOIN pg_class cl
    ON cl.relname      = c.table_name
   AND cl.relnamespace = n.oid
JOIN pg_attribute a
    ON a.attname  = c.column_name
   AND a.attrelid = cl.oid
JOIN pg_type t
    ON t.oid = a.atttypid
JOIN 
    kwild_engine.namespaces us ON n.nspname::TEXT = us.name
WHERE cl.relkind IN ('r', 'v') -- only tables and views
ORDER BY table_name, ordinal_position;
    

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
ORDER BY 
    table_name, name,
    1,2,3,4,5,6;

-- info.constraints is a public view that provides a list of all constraints in the database
CREATE VIEW info.constraints AS
SELECT 
    pg_namespace.nspname::TEXT AS namespace,
    split_part(conrelid::regclass::text, '.', 2) AS table_name,
    conname::TEXT AS name,
    CASE contype
        WHEN 'c' THEN 'CHECK'
        WHEN 'u' THEN 'UNIQUE'
    END AS constraint_type,
    array_agg(attname)::TEXT[] AS columns,
    pg_get_constraintdef(pg_constraint.oid) AS expression
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
    namespace, table_name, name,
    1, 2, 3, 4, 5, 6;

-- info.foreign_keys is a public view that provides a list of all foreign keys in the database
CREATE VIEW info.foreign_keys AS
SELECT 
    pg_namespace.nspname::TEXT AS namespace,
    split_part(conrelid::regclass::text, '.', 2) AS table_name,
    conname::TEXT AS name,
    array_agg(pg_attribute.attname ORDER BY cols.colnum)::TEXT[] AS columns,
    (SELECT split_part(confrelid::regclass::text, '.', 2)) AS ref_table,
    array_agg(ref_attr.attname ORDER BY ref_cols.colnum)::TEXT[] AS ref_columns,
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
    unnest(conkey) WITH ORDINALITY AS cols(colnum, ord) ON true
LEFT JOIN 
    pg_attribute ON pg_attribute.attnum = cols.colnum AND pg_attribute.attrelid = pg_class.oid
LEFT JOIN 
    unnest(confkey) WITH ORDINALITY AS ref_cols(colnum, ord) ON ref_cols.ord = cols.ord
LEFT JOIN 
    pg_attribute ref_attr ON ref_attr.attnum = ref_cols.colnum AND ref_attr.attrelid = confrelid
JOIN 
    kwild_engine.namespaces us ON pg_namespace.nspname::TEXT = us.name
WHERE 
    contype = 'f'  -- Only foreign key constraints
GROUP BY 
    pg_namespace.nspname, conname, conrelid, confrelid, confupdtype, confdeltype
ORDER BY 
    namespace, table_name, name,
    1, 2, 3, 4, 5, 6, 7, 8;



-- actions is a public view that provides a list of all actions in the database
CREATE VIEW info.actions AS
WITH parameters AS (
    SELECT 
        action_id,
        array_agg(
            p.name
            ORDER BY p.position, p.name, kwild_engine.format_type(p.scalar_type, p.is_array, p.metadata)
        ) AS parameter_names,
        array_agg(
            kwild_engine.format_type(p.scalar_type, p.is_array, p.metadata)
            ORDER BY p.position, p.name, kwild_engine.format_type(p.scalar_type, p.is_array, p.metadata)
        ) AS parameter_types
    FROM kwild_engine.parameters p
    GROUP BY action_id
), return_fields AS (
    SELECT 
        action_id,
        array_agg(r.name ORDER BY r.position, r.name, kwild_engine.format_type(r.scalar_type, r.is_array, r.metadata)) AS return_names,
        array_agg(kwild_engine.format_type(r.scalar_type, r.is_array, r.metadata) ORDER BY r.position, r.name, kwild_engine.format_type(r.scalar_type, r.is_array, r.metadata)) AS return_types
    FROM kwild_engine.return_fields r
    GROUP BY action_id
)
SELECT 
    a.namespace AS namespace,
    a.name::TEXT AS name,
    a.raw_statement AS raw_statement,
    a.modifiers::TEXT[] AS access_modifiers,
    COALESCE(p.parameter_names, ARRAY[]::TEXT[]) AS parameter_names,
    COALESCE(p.parameter_types, ARRAY[]::TEXT[]) AS parameter_types,
    COALESCE(r.return_names, ARRAY[]::TEXT[]) AS return_names,
    COALESCE(r.return_types, ARRAY[]::TEXT[]) AS return_types,
    a.returns_table AS returns_table,
    a.built_in AS built_in
FROM kwild_engine.actions a
LEFT JOIN parameters p
    ON a.id = p.action_id
LEFT JOIN return_fields r
    ON a.id = r.action_id
ORDER BY a.namespace, a.name,
    1, 2, 3, 4, 5, 6, 7, 8, 9;


-- roles is a public view that provides a list of all roles in the database
CREATE VIEW info.roles AS
SELECT 
    name,
    built_in
FROM
    kwild_engine.roles
ORDER BY
    1, 2;

CREATE VIEW info.user_roles AS
SELECT 
    r.name AS role_name,
    user_identifier
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
    r.name AS role_name,
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
    array_agg(eip.key ORDER BY eip.key, eip.value) AS parameters,
    array_agg(eip.value ORDER BY eip.key, eip.value) AS values
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