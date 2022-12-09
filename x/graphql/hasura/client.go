package hasura

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	endpoint string
}

func NewClient(endpoint string) *Client {
	return &Client{
		endpoint: endpoint,
	}
}

func (c *Client) metadataUrl() string {
	s, _ := url.JoinPath(c.endpoint, "v1/metadata")
	return s
}

func (c *Client) graphqlUrl() string {
	s, _ := url.JoinPath(c.endpoint, "v1/graphql")
	return s
}

func (c *Client) queryUrl() string {
	s, _ := url.JoinPath(c.endpoint, "v2/query")
	return s
}

func (c *Client) explainUrl() string {
	s, _ := url.JoinPath(c.endpoint, "v1/graphql/explain")
	return s
}

func (c *Client) call(req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
	// uncomment if Hasura admin secret is enabled
	// req.Header.Set("X-Hasura-Role", "admin")
	// req.Header.Set("X-Hasura-Admin-Secret", viper.GetString("hasuraadminsecret"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("call Hasura failed: %s", err.Error())
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode != http.StatusOK {
		var hasuraErr ErrorResp
		if err := json.Unmarshal(bodyBytes, &hasuraErr); err != nil {
			return []byte{}, fmt.Errorf("parse Hasura response failed: %s", err.Error())
		}
		return bodyBytes, fmt.Errorf("code: %s, error: %s", hasuraErr.Code, hasuraErr.Error)
	}

	return bodyBytes, nil
}

// TrackTable call Hasura API to expose a table under 'source.schema'.
func (c *Client) TrackTable(source, schema, table string) error {
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
	req, err := http.NewRequest(http.MethodPost, c.metadataUrl(), bodyReader)
	if err != nil {
		return err
	}

	_, err = c.call(req)
	return err
}

// UntrackTable call Hasura API to un-expose a able under 'source.schema'.
func (c *Client) UntrackTable(source, schema, table string) error {
	untrackTableParams := newHasuraPgUntrackTableParams(source, schema, table)
	jsonBody, err := json.Marshal(untrackTableParams)
	if err != nil {
		return err
	}
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, c.metadataUrl(), bodyReader)
	if err != nil {
		return err
	}

	_, err = c.call(req)
	return err
}

// UpdateTable first untrack table from 'source.schema', then track it again.
func (c *Client) UpdateTable(source, schema, table string) error {
	// if table is already tracked, need to untrack and then track
	if err := c.UntrackTable(source, schema, table); err != nil {
		return err
	}

	if err := c.TrackTable(source, schema, table); err != nil {
		return err
	}
	return nil
}

// AddDefaultSourceAndSchema add 'default' source and 'public' schema
// from db url configured in ENV.
func (c *Client) AddDefaultSourceAndSchema() error {
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
	req, err := http.NewRequest(http.MethodPost, c.metadataUrl(), bodyReader)
	if err != nil {
		return err
	}

	_, err = c.call(req)
	return err
}

// AddSchema add schema to default source 'default'.
func (c *Client) AddSchema(schema string) error {
	addSchemaBody := fmt.Sprintf(`{"type":"run_sql",
			           "args":{"source":"default",
				               "sql":"create schema %s;",
						       "cascade":false,
						       "read_only":false}}`, schema)
	bodyReader := bytes.NewReader([]byte(addSchemaBody))
	req, err := http.NewRequest(http.MethodPost, c.queryUrl(), bodyReader)
	if err != nil {
		return err
	}

	_, err = c.call(req)
	return err
}

// DeleteSchema delete schema to default source 'default'.
// Set cascade to true to delete all dependent tables.
func (c *Client) DeleteSchema(schema string, cascade bool) error {
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
	req, err := http.NewRequest(http.MethodPost, c.queryUrl(), bodyReader)
	if err != nil {
		return err
	}

	_, err = c.call(req)
	return err
}

// HasInitialized return true if there is a source and a schema configured,
// otherwise return false.
func (c *Client) HasInitialized() (bool, error) {
	body := `{"type":"export_metadata","version":2,"args":{}}`
	bodyReader := bytes.NewReader([]byte(body))
	req, err := http.NewRequest(http.MethodPost, c.metadataUrl(), bodyReader)
	if err != nil {
		return false, err
	}

	respBody, err := c.call(req)
	if err != nil {
		return false, err
	}

	var resp ExportMetadataResp
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return false, err
	}

	if len(resp.Metadata.Sources) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

// ExplainQuery return compiled sql from query.
// Right now only support one query.
func (c *Client) ExplainQuery(query string) (string, error) {
	body := queryToExplain(query)
	bodyReader := bytes.NewReader([]byte(body))
	req, err := http.NewRequest(http.MethodPost, c.explainUrl(), bodyReader)
	if err != nil {
		return "", err
	}

	respBody, err := c.call(req)
	if err != nil {
		return "", err
	}

	var resp []ExplainResp
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", err
	}

	return resp[0].Sql, nil
}
