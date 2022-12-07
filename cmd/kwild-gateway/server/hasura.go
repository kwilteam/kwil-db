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
	Confituration HasuraTableConf     `json:"configuration"`
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
				Name:   convertHasuraTableName(table),
			},
			Confituration: HasuraTableConf{
				CustomName: convertHasuraTableName(table),
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
				Name:   convertHasuraTableName(table),
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

func convertHasuraTableName(name string) string {
	return strings.ToLower(strings.Replace(name, " ", "_", -1))
}

func (h *HasuraEngine) call(req *http.Request) error {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Hasura-Role", "admin")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		var hasuraErr HasuraErrorResp
		json.Unmarshal(bodyBytes, &hasuraErr)
		return fmt.Errorf("code: %s, error: %s", hasuraErr.Code, hasuraErr.Error)
	}

	return nil
}

func (h *HasuraEngine) trackTable(source, schema, table string) error {
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

	return h.call(req)
}

func (h *HasuraEngine) untrackTable(source, schema, table string) error {
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

	return h.call(req)
}

func (h *HasuraEngine) updateTable(source, schema, table string) error {
	// if table is already tracked, need to untrack and then track
	if err := h.untrackTable(source, schema, table); err != nil {
		return err
	}

	if err := h.trackTable(source, schema, table); err != nil {
		return err
	}
	return nil
}

func (h *HasuraEngine) addDefaultSourceAndSchema() error {
	addSource := fmt.Sprintf(`{"type":"pg_add_source",
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

	return h.call(req)
}

func (h *HasuraEngine) addDefaultSchema() error {
	addSchemaBody := `{"type":"run_sql",
			           "args":{"source":"default",
				               "sql":"create schema 'ppppp';",
						       "cascade":false,
						       "read_only":false}}`
	bodyReader := bytes.NewReader([]byte(addSchemaBody))
	req, err := http.NewRequest(http.MethodPost, h.queryUrl(), bodyReader)
	if err != nil {
		return err
	}

	return h.call(req)
}
