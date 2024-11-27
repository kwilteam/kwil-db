#!/bin/bash

# TIP: set password with env:
#  PGPASSWORD="kwild" DB_PORT=5454 ./pg_shared_mem.sh

# Database connection details (adjust as needed)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-kwild}"
DB_NAME="${DB_NAME:-kwild}"

# Function to fetch PostgreSQL settings
get_pg_setting() {
    psql --no-psqlrc -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SHOW $1;"
}

# Fetch PostgreSQL configuration settings
shared_buffers=$(get_pg_setting "shared_buffers")
wal_buffers=$(get_pg_setting "wal_buffers")
max_connections=$(get_pg_setting "max_connections")
max_locks_per_transaction=$(get_pg_setting "max_locks_per_transaction")
max_prepared_transactions=$(get_pg_setting "max_prepared_transactions")
autovacuum_max_workers=$(get_pg_setting "autovacuum_max_workers")
max_worker_processes=$(get_pg_setting "max_worker_processes")

echo "Key settings:"
echo "---------------------------------------"
echo "shared_buffers:             $shared_buffers"
echo "wal_buffers:                $wal_buffers"
echo "max_locks_per_transaction:  $max_locks_per_transaction"
echo "max_prepared_transactions:  $max_prepared_transactions"
echo "autovacuum_max_workers:     $autovacuum_max_workers"
echo "max_worker_processes:       ${max_worker_processes}"
echo ""

# Convert memory units to bytes
convert_to_bytes() {
    local value=$1
    case "$value" in
        *kB) echo $(( ${value%kB} * 1024 )) ;;
        *MB) echo $(( ${value%MB} * 1024 * 1024 )) ;;
        *GB) echo $(( ${value%GB} * 1024 * 1024 * 1024 )) ;;
        *) echo "$value" ;;
    esac
}

shared_buffers_bytes=$(convert_to_bytes "$shared_buffers")
wal_buffers_bytes=$(convert_to_bytes "$wal_buffers")

# Constants for memory usage
lock_size=200          # bytes per lock
worker_process_memory=$((8 * 1024)) # 8 KB per worker process
prepared_transaction_memory=$((6 * 1024)) # 6 KB per prepared transaction
fixed_overhead=$((1 * 1024 * 1024))  # 1 MB for fixed overhead

# Calculate memory usage
lock_memory=$((max_locks_per_transaction * (max_connections + max_prepared_transactions) * lock_size))
worker_memory=$(( (autovacuum_max_workers + max_worker_processes) * worker_process_memory ))
prepared_memory=$(( max_prepared_transactions * prepared_transaction_memory ))

total_shared_memory=$(( shared_buffers_bytes + wal_buffers_bytes + lock_memory + worker_memory + prepared_memory + fixed_overhead ))

# Convert the result to MB for easier readability
total_shared_memory_mb=$(( total_shared_memory / (1024) ))

# Display results
echo "Estimated PostgreSQL Shared Memory Usage:"
echo "---------------------------------------"
echo "Shared Buffers:               $((shared_buffers_bytes / (1024))) KB"
echo "WAL Buffers:                  $((wal_buffers_bytes / (1024))) KB"
echo "Locks Memory:                 $((lock_memory / (1024))) KB"
echo "Background Workers Memory:    $((worker_memory / (1024))) KB"
echo "Prepared Transactions Memory: $((prepared_memory / (1024))) KB"
echo "Fixed Overhead:               $((fixed_overhead / (1024))) KB"
echo "---------------------------------------"
echo "Total Estimated Shared Memory: ${total_shared_memory_mb} KB"
