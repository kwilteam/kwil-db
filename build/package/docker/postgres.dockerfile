FROM postgres:16.2

# Inject the init script that makes the kwild superuser and a kwild database
# owned by that kwild user, as well as a kwil_test_db database for tests.
COPY ./pginit.sql /docker-entrypoint-initdb.d/init.sql

# With the above, there is still a "postgres" superuser for administration, and
# the default database is still postgres. We can set the following variables to
# change those to "kwild", but cleanup needs a initdb since you cannot drop the
# database to which you are connected.
# ENV POSTGRES_USER kwild
# ENV POSTGRES_PASSWORD kwild
# ENV POSTGRES_DB kwild

# Override the default entrypoint/command to include the additional configuration
CMD ["postgres", "-c", "wal_level=logical", "-c", "max_wal_senders=10", "-c", "max_replication_slots=10", \
	"-c", "track_commit_timestamp=true", "-c", "wal_sender_timeout=0", "-c", "max_prepared_transactions=2"]
