CREATE OR REPLACE FUNCTION set_replica_identity_full()
RETURNS event_trigger AS $$
DECLARE
    obj record;
    has_key_or_index boolean;
BEGIN
    FOR obj IN SELECT * FROM pg_event_trigger_ddl_commands()
               WHERE command_tag = 'CREATE TABLE' AND object_type = 'table'
    LOOP
        SELECT EXISTS (
            SELECT 1 
            FROM pg_class c
            JOIN pg_namespace n ON c.relnamespace = n.oid
            LEFT JOIN pg_index i ON c.oid = i.indrelid
            WHERE n.nspname || '.' || c.relname = obj.object_identity
              AND (i.indisprimary OR i.indisunique)
        ) INTO has_key_or_index;

        -- alter table only if there is (no primary key) and (no unique index)
        IF NOT has_key_or_index THEN
            EXECUTE 'ALTER TABLE ' || obj.object_identity || ' REPLICA IDENTITY FULL';
            RAISE NOTICE 'Altered table: % to set REPLICA IDENTITY FULL.', obj.object_identity;
        ELSE
            RAISE NOTICE 'Table: % already has a primary key or unique index.', obj.object_identity;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
