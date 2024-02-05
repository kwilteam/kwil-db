FROM postgres:16.1

# Might do these envs, but for now leave it to compose:
# ENV POSTGRES_USER kwild
# ENV POSTGRES_PASSWORD kwild
# ENV POSTGRES_DB kwild

# Override the default entrypoint/command to include the additional configuration
CMD ["postgres", "-c", "wal_level=logical", "-c", "max_wal_senders=10", "-c", "max_replication_slots=10", "-c", "max_prepared_transactions=2"]
