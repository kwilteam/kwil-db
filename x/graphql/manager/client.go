package manager

type Client interface {
	TrackTable(source, schema, table string) error
	UntrackTable(source, schema, table string) error
	AddDefaultSourceAndSchema() error
	ExplainQuery(query string) (string, error)
}
