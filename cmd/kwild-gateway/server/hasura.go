package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	DefaultHasuraSource = "default"
	DefaultHasuraSchema = "public"
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

type HasuraPgTrackTableParams struct {
	Type string                 `json:"type"`
	Args HasuraPgTrackTableArgs `json:"args"`
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

func snakeCase(name string) string {
	return strings.ToLower(strings.Replace(name, " ", "_", -1))
}

// customHasuraTableName return "schema_table".
func customHasuraTableName(schema, table string) string {
	names := []string{snakeCase(schema), snakeCase(table)}
	return strings.Join(names, "_")
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

type HasuraEngine struct {
	endpoint string
}

func NewHasuraEngine(endpoint string) *HasuraEngine {
	return &HasuraEngine{
		endpoint: endpoint,
	}
}

func (h *HasuraEngine) metadataUrl() string {
	s, _ := url.JoinPath(h.endpoint, "v1/metadata")
	return s
}

func (h *HasuraEngine) graphqlUrl() string {
	s, _ := url.JoinPath(h.endpoint, "v1/graphql")
	return s
}

func (h *HasuraEngine) queryUrl() string {
	s, _ := url.JoinPath(h.endpoint, "v2/query")
	return s
}

func (h *HasuraEngine) explainUrl() string {
	s, _ := url.JoinPath(h.endpoint, "v1/graphql/explain")
	return s
}

func (h *HasuraEngine) call(req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
	// uncomment if Hasura admin secret is enabled
	// req.Header.Set("X-Hasura-Role", "admin")
	// req.Header.Set("X-Hasura-Admin-Secret", viper.GetString("hasuraadminsecret"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode != http.StatusOK {
		var hasuraErr HasuraErrorResp
		json.Unmarshal(bodyBytes, &hasuraErr)
		return bodyBytes, fmt.Errorf("code: %s, error: %s", hasuraErr.Code, hasuraErr.Error)
	}

	return bodyBytes, nil
}

// TrackTable call Hasura API to expose a table under 'source.schema'.
func (h *HasuraEngine) TrackTable(source, schema, table string) error {
	// no space is allowed in schema
	if strings.Contains(table, " ") {
		return fmt.Errorf("invalid table name: space is not allowed, '%s'", table)
	}
	trackTableParams := newHasuraPgTrackTableParams(source, schema, table)
	jsonBody, err := json.Marshal(trackTableParams)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest(http.MethodPost, h.metadataUrl(), bodyReader)
	if err != nil {
		return err
	}

	_, err = h.call(req)
	return err
}

// UntrackTable call Hasura API to unexpose a able uder 'source.schema'.
func (h *HasuraEngine) UntrackTable(source, schema, table string) error {
	untrackTableParams := newHasuraPgUntrackTableParams(source, schema, table)
	jsonBody, err := json.Marshal(untrackTableParams)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, h.metadataUrl(), bodyReader)
	if err != nil {
		return err
	}

	_, err = h.call(req)
	return err
}

// UpdateTable first untrack table from 'source.schema', then track it again.
func (h *HasuraEngine) UpdateTable(source, schema, table string) error {
	// if table is already tracked, need to untrack and then track
	if err := h.UntrackTable(source, schema, table); err != nil {
		return err
	}

	if err := h.TrackTable(source, schema, table); err != nil {
		return err
	}
	return nil
}

// AddDefaultSourceAndSchema add 'default' source and 'public' schema
// from db url configured in ENV.
func (h *HasuraEngine) AddDefaultSourceAndSchema() error {
	addSource := fmt.Sprintf(
		`{"type":"pg_add_source",
	 	  "args":{"name":"default",
		  "configuration":{"connection_info":{"database_url":{"from_env":"%s"},
						                      "use_prepared_statements":false,
						       				  "isolation_level":"read-committed"},
						   "read_replicas":null,
						   "extensions_schema":"public"},
		  "replace_configuration":false,
		  "customization":{}}}`,
		// configured in Hasura container
		"PG_DATABASE_URL")
	addSourceBody := []byte(addSource)
	bodyReader := bytes.NewReader(addSourceBody)
	req, err := http.NewRequest(http.MethodPost, h.metadataUrl(), bodyReader)
	if err != nil {
		return err
	}

	_, err = h.call(req)
	return err
}

// AddSchema add schema to default source 'default'.
func (h *HasuraEngine) AddSchema(schema string) error {
	addSchemaBody := fmt.Sprintf(`{"type":"run_sql",
			           "args":{"source":"default",
				               "sql":"create schema %s;",
						       "cascade":false,
						       "read_only":false}}`, schema)
	bodyReader := bytes.NewReader([]byte(addSchemaBody))
	req, err := http.NewRequest(http.MethodPost, h.queryUrl(), bodyReader)
	if err != nil {
		return err
	}

	_, err = h.call(req)
	return err
}

// DeleteSchema delete schema to default source 'default'.
// Set cascade to true to delete all dependent tables.
func (h *HasuraEngine) DeleteSchema(schema string, cascade bool) error {
	cascadeValue := ""
	if cascade {
		cascadeValue = "cascade"
	}
	addSchemaBody := fmt.Sprintf(`{"type":"run_sql",
			           "args":{"source":"default",
				               "sql":"drop schema %s %s;",
						       "cascade":true,
						       "read_only":false}}`, schema, cascadeValue)
	bodyReader := bytes.NewReader([]byte(addSchemaBody))
	req, err := http.NewRequest(http.MethodPost, h.queryUrl(), bodyReader)
	if err != nil {
		return err
	}

	_, err = h.call(req)
	return err
}

// HasInitialized return true if there is a source and a schema configured,
// otherwise return false.
func (h *HasuraEngine) HasInitialized() (bool, error) {
	body := `{"type":"export_metadata","version":2,"args":{}}`
	bodyReader := bytes.NewReader([]byte(body))
	req, err := http.NewRequest(http.MethodPost, h.metadataUrl(), bodyReader)
	if err != nil {
		return false, err
	}

	respBody, err := h.call(req)
	if err != nil {
		return false, err
	}

	var resp HasuraExportMetadataResp
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return false, err
	}

	if len(resp.Metadata.Sources) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

// queryToExplain return a query body for explain API
// Does not support Directives yet.
func queryToExplain(query string) string {
	queryHead, queryBody, _ := strings.Cut(query, "{")
	queryHead = strings.Trim(queryHead, " ")
	s := strings.Split(queryHead, " ")
	if len(s) <= 1 {
		return fmt.Sprintf(`{"query": {"query": "{%s"}}`, queryBody)
	} else {
		operationName := s[1]
		return fmt.Sprintf(`{"query": {"query": "%s", "operationName": "%s"}}`, query, operationName)
	}
}

// ExplainQuery return compiled sql from query.
// Right now only support one query.
func (h *HasuraEngine) ExplainQuery(query string) (string, error) {
	body := queryToExplain(query)
	bodyReader := bytes.NewReader([]byte(body))
	req, err := http.NewRequest(http.MethodPost, h.explainUrl(), bodyReader)
	if err != nil {
		return "", err
	}

	respBody, err := h.call(req)
	if err != nil {
		return "", err
	}

	var resp []HasuraExplainResp
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", err
	}

	return resp[0].Sql, nil
}
