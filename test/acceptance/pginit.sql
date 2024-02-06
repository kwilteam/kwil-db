-- These commands are run with psql inside the container after postgres starts.
CREATE USER kwild WITH PASSWORD 'kwild' SUPERUSER REPLICATION;
CREATE DATABASE kwild OWNER kwild;
\c kwild
CREATE PUBLICATION kwild_repl FOR ALL TABLES;
