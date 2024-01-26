FROM postgres:16.1

# Override the default entrypoint/command to include the additional configuration
CMD ["postgres", "-c", "wal_level=logical", "-c", "max_wal_senders=10", "-c", "max_replication_slots=10", "-c", "max_prepared_transactions=2"]
