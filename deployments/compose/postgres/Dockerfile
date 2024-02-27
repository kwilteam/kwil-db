FROM postgres:16.1

# Inject the init script that makes the kwild superuser and a kwild database
# owned by that kwild user, as well as a kwil_test_db database for tests.
COPY ./init.sql /docker-entrypoint-initdb.d/init.sql

# Override the default entrypoint/command to include the additional configuration
CMD ["postgres", "-c", "wal_level=logical", "-c", "max_wal_senders=10", "-c", "max_replication_slots=10", "-c", "max_prepared_transactions=2"]
