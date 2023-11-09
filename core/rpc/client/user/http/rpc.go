package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/big"

	"github.com/kwilteam/kwil-db/core/crypto/auth"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
)

func (c *Client) ChainInfo(ctx context.Context) (*types.ChainInfo, error) {
	req, err := newChainInfoRequest(c.target)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parseChainInfoResponse(resp)
}

func (c *Client) Call(ctx context.Context, msg *transactions.CallMessage) ([]map[string]any, error) {
	req, err := newActionCallRequest(c.target, msg)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// NOTE: probably should start to use status.FromError
	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parseActionCallResponse(resp)
}

func (c *Client) TxQuery(ctx context.Context, txHash []byte) (*transactions.TcTxQueryResponse, error) {
	req, err := newTxQueryRequest(c.target, txHash)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parseTxQueryResponse(resp)
}

func (c *Client) GetSchema(ctx context.Context, dbid string) (*transactions.Schema, error) {
	req, err := newGetSchemaRequest(c.target, dbid)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parseGetSchemaResponse(resp)
}

func (c *Client) Query(ctx context.Context, dbid string, query string) ([]map[string]any, error) {
	req, err := newDBQueryRequest(c.target, dbid, query)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parseDBQueryResponse(resp)
}

func (c *Client) ListDatabases(ctx context.Context, ownerPubKey []byte) ([]string, error) {
	req, err := newListDatabasesRequest(c.target, ownerPubKey)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parseListDatabasesResponse(resp)
}

func (c *Client) GetAccount(ctx context.Context, pubKey []byte) (*types.Account, error) {
	req, err := newGetAccountRequest(c.target, pubKey)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parseGetAccountResponse(resp)
}

func (c *Client) Broadcast(ctx context.Context, tx *transactions.Transaction) ([]byte, error) {
	req, err := newBroadcastRequest(c.target, tx)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parseBroadcastResponse(resp)
}

// unmarshalResponse unmarshal the response into v, which should be a pointer.
func unmarshalResponse(resp io.Reader, v any) error {
	decoder := json.NewDecoder(resp)
	return decoder.Decode(v)
}

// Ping sends a ping request to the target and returns the response body
func (c *Client) Ping(ctx context.Context) (string, error) {
	req, err := newPingRequest(c.target)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parsePingResponse(resp)
}

func (c *Client) EstimateCost(ctx context.Context, tx *transactions.Transaction) (*big.Int, error) {
	req, err := newEstimateCostRequest(c.target, tx)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parseEstimateCostResponse(resp)
}

func (c *Client) ValidatorJoinStatus(ctx context.Context, pubKey []byte) (*types.JoinRequest, error) {
	req, err := newValidatorJoinStatusRequest(c.target, pubKey)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parseValidatorJoinStatusResponse(resp, pubKey)
}

func (c *Client) CurrentValidators(ctx context.Context) ([]*types.Validator, error) {
	req, err := newCurrentValidatorsRequest(c.target)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parseCurrentValidatorsResponse(resp)
}

func (c *Client) VerifySignature(ctx context.Context, sender []byte,
	signature *auth.Signature, message []byte) (bool, error) {
	req, err := newVerifySignatureRequest(c.target, sender, signature, message)
	if err != nil {
		return false, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.makeRequest(req.WithContext(ctx))
	if err != nil {
		return false, fmt.Errorf("make request: %w", err)
	}
	defer resp.Body.Close()

	return parseVerifySignatureResponse(resp)
}
