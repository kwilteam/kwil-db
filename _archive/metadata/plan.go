package metadata

type SchemaRequest struct {
	Wallet     string
	Database   string
	SchemaData []byte
}

type PlanInfo struct {
	Wallet   string
	Database string
	Data     []byte
}

type Plan struct {
	Changes []Change
}

type Change struct {
	Cmd     string
	Comment string
}
