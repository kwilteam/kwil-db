package service

import v0 "kwil/x/api/v0"

const DATABASE_EMITTER_ALIAS = "service.database.emitter"

type DBRequest struct {
	IdempotentKey string // key to use for de-duplicating commands in the target Db and for retrieving the request info
	RoutingKey    string // represents a command path to ensure messages are processed in order
	DBCommand     string // actual DDL or INSERT/UPDATE/DELETE SQL
}

func getCreateDbRequest(req *v0.CreateDatabaseRequest) *DBRequest {
	panic("not implemented")
}

func getUpdateDbRequest(req *v0.UpdateDatabaseRequest) *DBRequest {
	panic("not implemented")
}
