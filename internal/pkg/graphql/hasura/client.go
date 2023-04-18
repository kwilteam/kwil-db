package hasura

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"io"
	"kwil/pkg/log"
	"net/http"
	"net/url"
	"strings"
)

type client struct {
	endpoint string
	logger   log.Logger
}

func NewClient(endpoint string, logger log.Logger) *client {
	return &client{
		endpoint: endpoint,
		logger:   *logger.Named("hasura"),
	}
}

func (c *client) metadataURL() string {
	s, _ := url.JoinPath(c.endpoint, "v1/metadata")
	return s
}

func (c *client) queryURL() string {
	s, _ := url.JoinPath(c.endpoint, "v2/query")
	return s
}

func (c *client) explainURL() string {
	s, _ := url.JoinPath(c.endpoint, "v1/graphql/explain")
	return s
}

func (c *client) call(req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
	// uncomment if Hasura admin secret is enabled
	// req.Header.Set("X-Hasura-Role", "admin")
	// req.Header.Set("X-Hasura-Admin-Secret", viper.GetString("hasuraadminsecret"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("call Graphql failed: %s", err.Error())
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode != http.StatusOK {
		var hasuraErr ErrorResp
		if err := json.Unmarshal(bodyBytes, &hasuraErr); err != nil {
			return []byte{}, fmt.Errorf("parse Graphql response failed: %s", err.Error())
		}
		return bodyBytes, fmt.Errorf("code: %s, error: %s", hasuraErr.Code, hasuraErr.Error)
	}

	return bodyBytes, nil
}

// TrackTable call Hasura API to expose a table under 'source.schema'.
func (c *client) TrackTable(source, schema, table string) error {
	fields := zap.Fields(zap.String("source", source), zap.String("schema", schema), zap.String("table", table))
	logger := c.logger.WithOptions(fields)
	logger.Info("track table")
	// no space is allowed in schema
	if strings.Contains(table, " ") {
		return fmt.Errorf("track table failed: invalid table name, '%s'", table)
	}
	trackTableParams := newHasuraPgTrackTableParams(source, schema, table)
	jsonBody, err := json.Marshal(trackTableParams)
	if err != nil {
		logger.Error("track table failed", zap.Error(err))
		return fmt.Errorf("track table failed: %w", err)
	}
	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest(http.MethodPost, c.metadataURL(), bodyReader)
	if err != nil {
		logger.Error("track table failed", zap.Error(err))
		return fmt.Errorf("track table failed: %w", err)
	}

	_, err = c.call(req)
	return err
}

// UntrackTable call Hasura API to un-expose a able under 'source.schema'.
func (c *client) UntrackTable(source, schema, table string) error {
	fields := zap.Fields(zap.String("source", source), zap.String("schema", schema), zap.String("table", table))
	logger := c.logger.WithOptions(fields)
	logger.Info("untrack table")
	untrackTableParams := newHasuraPgUntrackTableParams(source, schema, table)
	jsonBody, err := json.Marshal(untrackTableParams)
	if err != nil {
		logger.Error("untrack table failed", zap.Error(err))
		return fmt.Errorf("untrack table failed: %s", err.Error())
	}
	bodyReader := bytes.NewReader(jsonBody)

	req, err := http.NewRequest(http.MethodPost, c.metadataURL(), bodyReader)
	if err != nil {
		logger.Error("untrack table failed", zap.Error(err))
		return fmt.Errorf("untrack table failed: %s", err.Error())
	}

	_, err = c.call(req)
	return err
}

// UpdateTable first untrack table from 'source.schema', then track it again.
func (c *client) UpdateTable(source, schema, table string) error {
	fields := zap.Fields(zap.String("source", source), zap.String("schema", schema), zap.String("table", table))
	logger := c.logger.WithOptions(fields)
	logger.Info("update table")
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
func (c *client) AddDefaultSourceAndSchema() error {
	c.logger.Info("add default source and schema")
	// PG_DATABASE_URL is configured in Hasura container
	addSource := fmt.Sprintf(
		`{"type":"pg_add_source",
				 "args":
				   {"name":"default",
		  			"configuration":{"connection_info":{"database_url":{"from_env":"%s"},
						                      "use_prepared_statements":false,
						       				  "isolation_level":"read-committed"},
							  	    "read_replicas":null,
						   			"extensions_schema":"public"},
		  		 "replace_configuration":false,
		         "customization":{}}}`, "PG_DATABASE_URL")
	addSourceBody := []byte(addSource)
	bodyReader := bytes.NewReader(addSourceBody)
	req, err := http.NewRequest(http.MethodPost, c.metadataURL(), bodyReader)
	if err != nil {
		c.logger.Error("add default source failed", zap.Error(err))
		return err
	}

	_, err = c.call(req)
	return err
}

// AddSchema add schema to default source 'default'.
func (c *client) AddSchema(schema string) error {
	addSchemaBody := fmt.Sprintf(`{"type":"run_sql",
			           "args":{"source":"default",
				               "sql":"create schema %s;",
						       "cascade":false,
						       "read_only":false}}`, schema)
	bodyReader := bytes.NewReader([]byte(addSchemaBody))
	req, err := http.NewRequest(http.MethodPost, c.queryURL(), bodyReader)
	if err != nil {
		return err
	}

	_, err = c.call(req)
	return err
}

// DeleteSchema delete schema to default source 'default'.
// Set cascade to true to delete all dependent tables.
func (c *client) DeleteSchema(schema string, cascade bool) error {
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
	req, err := http.NewRequest(http.MethodPost, c.queryURL(), bodyReader)
	if err != nil {
		return err
	}

	_, err = c.call(req)
	return err
}

// HasInitialized return true if there is a source and a schema configured,
// otherwise return false.
func (c *client) HasInitialized() (bool, error) {
	body := `{"type":"export_metadata","version":2,"args":{}}`
	bodyReader := bytes.NewReader([]byte(body))
	req, err := http.NewRequest(http.MethodPost, c.metadataURL(), bodyReader)
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
func (c *client) ExplainQuery(query string) (string, error) {
	c.logger.Debug("explain query", zap.String("query", query))
	body := queryToExplain(query)
	bodyReader := bytes.NewReader([]byte(body))
	req, err := http.NewRequest(http.MethodPost, c.explainURL(), bodyReader)
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
