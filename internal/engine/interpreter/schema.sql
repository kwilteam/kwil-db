/*
    This section contains all of the DDL for creating the schema for `kwild_engine`, which is
    the internal schema for the engine. This stores all metadata for actions.
*/
CREATE SCHEMA kwild_engine;

DO $$ 
BEGIN
    -- scalar_data_type is an enumeration of all scalar data types supported by the engine
    BEGIN
        CREATE TYPE kwild_engine.scalar_data_type AS ENUM (
            'int8', 'text', 'bool', 'uuid', 'numeric', 'bytea'
        );
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END;

    -- modifiers is an enumeration of all modifiers that can be applied to an action
    BEGIN
        CREATE TYPE kwild_engine.modifiers AS ENUM (
            'VIEW', 'OWNER'
        );
    EXCEPTION
        WHEN duplicate_object THEN NULL;
    END;
END $$;

-- user_namespaces is a table that stores all user schemas in the engine
CREATE TABLE IF NOT EXISTS kwild_engine.user_namespaces (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    owner BYTEA NOT NULL
);

-- actions is a table that stores all actions in the engine
CREATE TABLE IF NOT EXISTS kwild_engine.actions (
    id BIGSERIAL PRIMARY KEY,
    schema_name TEXT NOT NULL REFERENCES kwild_engine.user_namespaces(name) ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL UNIQUE,
    public BOOLEAN NOT NULL DEFAULT FALSE,
    raw_body TEXT NOT NULL,
    returns_table BOOLEAN NOT NULL DEFAULT FALSE,
    modifiers kwild_engine.modifiers[]
);

-- parameters is a table that stores all parameters for actions in the engine
CREATE TABLE IF NOT EXISTS kwild_engine.parameters (
    id BIGSERIAL PRIMARY KEY,
    action_id INT8 NOT NULL REFERENCES kwild_engine.actions(id) ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (name = lower(name)),
    scalar_type kwild_engine.scalar_data_type NOT NULL,
    is_array BOOLEAN NOT NULL,
    metadata BYTEA DEFAULT NULL
);

-- return_types is a table that stores all return types for actions in the engine
CREATE TABLE IF NOT EXISTS kwild_engine.return_fields (
    id BIGSERIAL PRIMARY KEY,
    action_id INT8 NOT NULL REFERENCES kwild_engine.actions(id) ON UPDATE CASCADE ON DELETE CASCADE,
    name TEXT NOT NULL CHECK (name = lower(name)),
    scalar_type kwild_engine.scalar_data_type NOT NULL,
    is_array BOOLEAN NOT NULL,
    metadata BYTEA DEFAULT NULL
);

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


-- drop_action is a function that drops an action from the engine
CREATE FUNCTION kwild_engine.drop_action(
    action_name_input TEXT
) RETURNS VOID AS $$
BEGIN
    DELETE FROM kwild_engine.actions
    WHERE name = action_name_input;
END;
$$ LANGUAGE plpgsql;

/*
    This section creates the schema the `kwild` schema, which is the public user-facing schema.
    End users can access the views in this schema to get information about the database.

    All views are ordered to ensure that they are deterministic when queried.
*/
CREATE SCHEMA kwild;

SET search_path TO kwild;

-- kwil_tables is a public view that provides a list of all tables in the database
CREATE VIEW kwil_tables AS
SELECT tablename::TEXT AS name, schemaname::TEXT AS schema
FROM pg_tables
JOIN kwild_engine.user_namespaces us
    ON schemaname = us.name
ORDER BY 1, 2;

-- kwil_columns is a public view that provides a list of all columns in the database
CREATE VIEW kwil_columns AS
SELECT 
    c.table_schema::TEXT AS schema_name,
    c.table_name::TEXT AS table_name,
    c.column_name::TEXT AS column_name,
    CASE 
        WHEN t.typcategory = 'A'
            THEN (pg_catalog.format_type(a.atttypid, a.atttypmod))
        ELSE c.data_type
    END AS data_type,
    c.is_nullable::bool AS is_nullable,
    c.column_default AS default_value,
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
    kwild_engine.user_namespaces us ON n.nspname::TEXT = us.name
WHERE cl.relkind = 'r'  -- Only include regular tables
ORDER BY 
    c.table_name, 
    c.ordinal_position,
    1, 2, 3, 4, 5, 6, 7, 8;

-- kwil_indexes is a public view that provides a list of all indexes in the database
CREATE VIEW kwil_indexes AS
SELECT 
    n.nspname::TEXT AS schema_name,
    c.relname::TEXT AS table_name,
    ic.relname::TEXT AS index_name,
    i.indisprimary AS is_pk,
    i.indisunique AS is_unique,
    array_agg(a.attname ORDER BY x.ordinality) AS column_names
FROM pg_index i
JOIN pg_class c ON c.oid = i.indrelid
JOIN pg_class ic ON ic.oid = i.indexrelid
JOIN pg_namespace n ON c.relnamespace = n.oid
JOIN pg_am am ON ic.relam = am.oid
JOIN pg_attribute a ON a.attnum = ANY(i.indkey) AND a.attrelid = c.oid
JOIN LATERAL unnest(i.indkey) WITH ORDINALITY AS x(colnum, ordinality) ON x.colnum = a.attnum
JOIN 
    kwild_engine.user_namespaces us ON n.nspname::TEXT = us.name
GROUP BY n.nspname, c.relname, ic.relname, i.indisprimary, i.indisunique
ORDER BY 1,2,3,4,5,6;

-- kwil_constraints is a public view that provides a list of all constraints in the database
CREATE VIEW kwil_constraints AS
SELECT 
    pg_namespace.nspname::TEXT AS schema_name,
    conname AS constraint_name,
    split_part(conrelid::regclass::text, '.', 2) AS table_name,
    array_agg(attname) AS columns,
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
    kwild_engine.user_namespaces us ON pg_namespace.nspname::TEXT = us.name
WHERE 
    contype = 'c'  -- Only check constraints
    OR contype = 'u'  -- Only unique constraints 
GROUP BY 
    pg_namespace.nspname, conname, conrelid, pg_constraint.oid
ORDER BY 
    1, 2, 3, 4, 5, 6;

-- kwil_foreign_keys is a public view that provides a list of all foreign keys in the database
CREATE VIEW kwil_foreign_keys AS
SELECT 
    pg_namespace.nspname::TEXT AS schema_name,
    conname AS constraint_name,
    split_part(conrelid::regclass::text, '.', 2) AS table_name,
    array_agg(attname) AS columns,
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
    kwild_engine.user_namespaces us ON pg_namespace.nspname::TEXT = us.name
WHERE 
    contype = 'f'  -- Only foreign key constraints
GROUP BY 
    pg_namespace.nspname, conname, conrelid, confupdtype, confdeltype
ORDER BY 
    table_name, constraint_name,
    1, 2, 3, 4, 5, 6;

-- kwil_actions is a public view that provides a list of all actions in the database
CREATE VIEW kwil_actions AS
SELECT 
    a.schema_name,
    a.id,
    a.name::TEXT,
    a.public,
    a.raw_body,
    a.modifiers::TEXT[] AS modifiers,
    a.returns_table,
    array_agg(
        json_build_object(
            'name', p.name,
            'data_type', kwild_engine.format_type(p.scalar_type, p.is_array, p.metadata)
        )
        ORDER BY p.name, kwild_engine.format_type(p.scalar_type, p.is_array, p.metadata)
    ) AS parameters,
    array_agg(
        json_build_object(
            'name', r.name,
            'data_type', kwild_engine.format_type(r.scalar_type, r.is_array, r.metadata)
        )
        ORDER BY r.name, kwild_engine.format_type(r.scalar_type, r.is_array, r.metadata)
    ) AS return_types
FROM kwild_engine.actions a
JOIN kwild_engine.parameters p
    ON a.id = p.action_id
LEFT JOIN kwild_engine.return_fields r
    ON a.id = r.action_id
GROUP BY a.schema_name, a.id, a.name, a.public, a.raw_body, a.returns_table
ORDER BY a.name,
    1, 2, 3, 4, 5, 6; --TODO: do we need to order 7, 8?
