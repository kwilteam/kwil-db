package models

type ActionExecution struct {
	Action string              `json:"action"`
	DBID   string              `json:"dbid"`
	Params []map[string][]byte `json:"params"`
}
