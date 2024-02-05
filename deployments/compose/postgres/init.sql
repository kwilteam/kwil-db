-- These commands are run with psql inside the container after postgres starts.
CREATE USER kwild WITH PASSWORD 'kwild' SUPERUSER REPLICATION;
CREATE DATABASE kwild OWNER kwild;
-- the tests db:
CREATE DATABASE kwil_test_db OWNER kwild;
