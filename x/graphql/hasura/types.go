package hasura

const (
	AdminSecretName = "hasuraadminsecret"
	AdminSecretEnv  = "HASURA_GRAPHQL_ADMIN_SECRET"
	UnathorizedRole = "HASURA_GRAPHQL_UNAUTHORIZED_ROLE"
	DefaultSource   = "default"
	DefaultSchema   = "public"
	EndpointEnv     = "KWIL_GRAPHQL_ENDPOINT"
)

type HasuraErrorResp struct {
	Code  string `json:"code"`
	Error string `json:"error"`
	Path  string `json:"path"`
}

type HasuraTableConf struct {
	CustomName string `json:"custom_name"`
}

type HasuraQulifiedTable struct {
	Schema string `json:"schema"`
	Name   string `json:"name"`
}

type HasuraPgTrackTableArgs struct {
	Table         HasuraQulifiedTable `json:"table"`
	Source        string              `json:"source"`
	Configuration HasuraTableConf     `json:"configuration"`
}

type HasuraResource struct{}

type HasuraMetadata struct {
	Version string           `json:"version"`
	Sources []HasuraResource `json:"sources"`
}

type HasuraExportMetadataResp struct {
	ResourceVersion int            `json:"resource_version"`
	Metadata        HasuraMetadata `json:"metadata"`
}

type HasuraExplainResp struct {
	Field string   `json:"field"`
	Sql   string   `json:"sql"`
	Plan  []string `json:"plan"`
}

type HasuraPgTrackTableParams struct {
	Type string                 `json:"type"`
	Args HasuraPgTrackTableArgs `json:"args"`
}

func newHasuraPgTrackTableParams(source, schema, table string) HasuraPgTrackTableParams {
	return HasuraPgTrackTableParams{
		Type: "pg_track_table",
		Args: HasuraPgTrackTableArgs{
			Source: source,
			Table: HasuraQulifiedTable{
				Schema: schema,
				Name:   table,
			},
			Configuration: HasuraTableConf{
				CustomName: customHasuraTableName(schema, table),
			},
		},
	}
}

type HasuraPgUntrackTableArgs struct {
	Source  string              `json:"source"`
	Table   HasuraQulifiedTable `json:"table"`
	Cascade bool                `json:"cascade"`
}

type HasuraPgUntrackTableParams struct {
	Type string                   `json:"type"`
	Args HasuraPgUntrackTableArgs `json:"args"`
}

func newHasuraPgUntrackTableParams(source, schema, table string) HasuraPgUntrackTableParams {
	return HasuraPgUntrackTableParams{
		Type: "pg_untrack_table",
		Args: HasuraPgUntrackTableArgs{
			Source: source,
			Table: HasuraQulifiedTable{
				Schema: schema,
				Name:   table,
			},
			Cascade: false,
		},
	}
}
