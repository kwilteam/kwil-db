package query

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Query(ctx context.Context, graphqlURL, query string) ([]byte, error) {
	payload := fmt.Sprintf(`{"query":"%s"}`, query)
	bodyReader := bytes.NewReader([]byte(payload))
	// @yaiba TODO: better url composition
	req, err := http.NewRequest(http.MethodPost, graphqlURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create graphql query request failed: %w", err)
	}

	return callGraphql(ctx, req)
}

func callGraphql(ctx context.Context, req *http.Request) ([]byte, error) {
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
