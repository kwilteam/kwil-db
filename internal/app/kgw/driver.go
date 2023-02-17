package kgw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kwil/internal/app/kwild"
	"net/http"
)

// KgwDriver is a driver for the gw client for integration tests
type KgwDriver struct {
	kwild.KwildDriver

	gatewayAddr string // to ignore the gatewayAddr returned by the config.service
}

func NewKgwDriver(gatewayAddr string, kwildDriver *kwild.KwildDriver) *KgwDriver {
	return &KgwDriver{
		KwildDriver: *kwildDriver,
		gatewayAddr: gatewayAddr,
	}
}

func (d *KgwDriver) QueryDatabase(ctx context.Context, query string) ([]byte, error) {
	payload := fmt.Sprintf(`{"query":"%s"}`, query)
	bodyReader := bytes.NewReader([]byte(payload))
	// @yaiba TODO: better url composition
	url := fmt.Sprintf("http://%s/graphql", d.gatewayAddr)
	req, err := http.NewRequest(http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create graphql query request failed: %w", err)
	}

	return d.callGraphql(ctx, req)
}

func (d *KgwDriver) callGraphql(ctx context.Context, req *http.Request) ([]byte, error) {
	req.Header.Set("Content-Type", "application/json")
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
