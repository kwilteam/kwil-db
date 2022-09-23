package engine

type Kind string

const (
	EngineMySQL      Kind = "mysql"
	EnginePostgreSQL Kind = "postgresql"
	EngineSQLite     Kind = "sqlite"
)
