package kgw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kwil/internal/app/kwild"
	"kwil/pkg/databases"
	"kwil/pkg/kclient"
	"net/http"
)

// Driver is a driver for the gw client for integration tests
type Driver struct {
	grpcClt   *kwild.Driver
	graphAddr string
	apiKey    string
}

func NewDriver(cfg *kclient.Config, graphqlAddr string, apiKey string) *Driver {
	return &Driver{
		grpcClt:   kwild.NewDriver(cfg),
		graphAddr: graphqlAddr,
		apiKey:    apiKey,
	}
}

func (d *Driver) DeployDatabase(ctx context.Context, db *databases.Database[[]byte]) error {
	return d.grpcClt.DeployDatabase(ctx, db)
}

func (d *Driver) DatabaseShouldExists(ctx context.Context, dbName string) error {
	return d.grpcClt.DatabaseShouldExists(ctx, dbName)
}

func (d *Driver) ExecuteQuery(ctx context.Context, dbName string, queryName string, queryInputs []any) error {
	return d.grpcClt.ExecuteQuery(ctx, dbName, queryName, queryInputs)
}

func (d *Driver) DropDatabase(ctx context.Context, dbName string) error {
	return d.grpcClt.DropDatabase(ctx, dbName)
}

func (d *Driver) QueryDatabase(ctx context.Context, query string) ([]byte, error) {
	payload := fmt.Sprintf(`{"query":"%s"}`, query)
	bodyReader := bytes.NewReader([]byte(payload))
	// @yaiba TODO: better url composition
	url := fmt.Sprintf("http://%s/graphql", d.graphAddr)
	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create graphql query request failed: %w", err)
	}

	return d.callGraphql(ctx, req)
}

func (d *Driver) callGraphql(ctx context.Context, req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", d.apiKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("call Graphql failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	type ErrorResp struct {
		Code  string `json:"code"`
		Error string `json:"error"`
		Path  string `json:"path"`
	}

	if resp.StatusCode != http.StatusOK {
		var hasuraErr ErrorResp
		if err := json.Unmarshal(bodyBytes, &hasuraErr); err != nil {
			return []byte{}, fmt.Errorf("parse Graphql response failed: %w", err)
		}
		return bodyBytes, fmt.Errorf("code: %s, error: %s", hasuraErr.Code, hasuraErr.Error)
	}

	return bodyBytes, nil
}
