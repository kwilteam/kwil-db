package core

type EngineKind string

const (
	EngineMySQL      EngineKind = "mysql"
	EnginePostgreSQL EngineKind = "postgresql"
	EngineSQLite     EngineKind = "sqlite"
)
