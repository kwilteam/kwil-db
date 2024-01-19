CREATE OR REPLACE FUNCTION set_replica_identity_full()
RETURNS event_trigger AS $$
DECLARE                                      
    obj record;
    has_primary_key boolean;
    has_unique_index boolean;
BEGIN
    FOR obj IN SELECT * FROM pg_event_trigger_ddl_commands() 
                WHERE command_tag = 'CREATE TABLE' AND object_type = 'table'
    LOOP
        SELECT EXISTS (
            SELECT 1 
            FROM information_schema.table_constraints 
            WHERE table_schema || '.' || table_name = obj.object_identity 
                AND constraint_type = 'PRIMARY KEY'
        ) INTO has_primary_key;

        SELECT EXISTS (
            SELECT 1 
            FROM pg_indexes 
            WHERE schemaname || '.' || tablename = obj.object_identity 
                AND indexdef LIKE '%UNIQUE%'
        ) INTO has_unique_index;

        -- alter table only if there is (no primary key) and (no unique index)
        IF NOT has_primary_key AND NOT has_unique_index THEN
            EXECUTE 'ALTER TABLE ' || obj.object_identity || ' REPLICA IDENTITY FULL'; -- note that object_identity is schema qualified
            RAISE NOTICE 'Altered table: % to set REPLICA IDENTITY FULL.', obj.object_identity;
        ELSE
            RAISE NOTICE 'Table: % already has a primary key or unique index.', obj.object_identity;
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
