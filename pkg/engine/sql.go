package engine

const (
	sqlInitTables = `
	CREATE TABLE IF NOT EXISTS databases (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		dbid TEXT NOT NULL UNIQUE,
		owner TEXT NOT NULL,
		name TEXT NOT NULL,
		UNIQUE (owner, name)
	);
	`
)

const (
	sqlListDatabases = "SELECT dbid, name, owner FROM databases;"
	sqlDeleteDataset = "DELETE FROM databases WHERE dbid = $dbid;"
	sqlCreateDataset = "INSERT INTO databases (dbid, name, owner) VALUES ($dbid, $name, $owner);"
	sqlGetDataset    = "SELECT name, owner FROM databases WHERE dbid = $dbid;"
)
