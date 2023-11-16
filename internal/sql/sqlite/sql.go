package sqlite

const (
	sqlPragmaSync    = "PRAGMA synchronous = FULL;"
	sqlCheckpoint    = "PRAGMA wal_checkpoint(TRUNCATE);"
	sqlEnableFK      = "PRAGMA foreign_keys = ON;"
	sqlDisableFK     = "PRAGMA foreign_keys = OFF;"
	sqlIfTableExists = `SELECT * FROM sqlite_master WHERE type='table' AND name=$name;`
	sqlListTables    = `SELECT name FROM sqlite_master WHERE type='table';`
)
