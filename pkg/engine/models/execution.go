package models

type ActionExecution struct {
	Action string
	DBID   string
	Params []map[string][]byte
}
