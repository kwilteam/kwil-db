package hasura

const (
	AdminSecretName     = "hasuraadminsecret"
	AdminSecretEnv      = "HASURA_GRAPHQL_ADMIN_SECRET"
	UnauthorizedRole    = "HASURA_GRAPHQL_UNAUTHORIZED_ROLE"
	DefaultSource       = "default"
	DefaultSchema       = "public"
	EndpointEnv         = "KWIL_GRAPHQL_ENDPOINT"
	GraphqlEndpointName = "graphql"
)

type ErrorResp struct {
	Code  string `json:"code"`
	Error string `json:"error"`
	Path  string `json:"path"`
}

type TableConf struct {
	CustomName string `json:"custom_name"`
}

type QualifiedTable struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
}

type PgTrackTableArgs struct {
	Table         QualifiedTable `json:"table"`
	Source        string         `json:"source"`
	Configuration TableConf      `json:"configuration"`
}

type Resource struct{}

type Metadata struct {
	Version string     `json:"version"`
	Sources []Resource `json:"sources"`
}

type ExportMetadataResp struct {
	ResourceVersion int      `json:"resource_version"`
	Metadata        Metadata `json:"metadata"`
}

type ExplainResp struct {
	Field string   `json:"field"`
	Sql   string   `json:"sql"`
	Plan  []string `json:"plan"`
}

type PgTrackTableParams struct {
	Type string           `json:"type"`
	Args PgTrackTableArgs `json:"args"`
}

func newHasuraPgTrackTableParams(source, schema, table string) PgTrackTableParams {
	return PgTrackTableParams{
		Type: "pg_track_table",
		Args: PgTrackTableArgs{
			Source: source,
			Table: QualifiedTable{
				Schema: schema,
				Name:   table,
			},
			Configuration: TableConf{
				CustomName: customTableName(schema, table),
			},
		},
	}
}

type PgUntrackTableArgs struct {
	Source  string         `json:"source"`
	Table   QualifiedTable `json:"table"`
	Cascade bool           `json:"cascade"`
}

type PgUntrackTableParams struct {
	Type string             `json:"type"`
	Args PgUntrackTableArgs `json:"args"`
}

func newHasuraPgUntrackTableParams(source, schema, table string) PgUntrackTableParams {
	return PgUntrackTableParams{
		Type: "pg_untrack_table",
		Args: PgUntrackTableArgs{
			Source: source,
			Table: QualifiedTable{
				Schema: schema,
				Name:   table,
			},
			Cascade: false,
		},
	}
}
