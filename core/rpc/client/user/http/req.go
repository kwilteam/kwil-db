package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"

	"github.com/kwilteam/kwil-db/core/crypto/auth"
	"github.com/kwilteam/kwil-db/core/rpc/client"
	"github.com/kwilteam/kwil-db/core/rpc/conversion"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// NOTE: a lot of boilerplate code, and part of the logic is kind of duplicated
// from the grpc client
// TODO: refactor

func NewGetRequest(server string, path string) (*http.Request, error) {
	target, err := url.JoinPath(server, path)
	if err != nil {
		return nil, err
	}

	return http.NewRequest(http.MethodGet, target, nil)
}

func NewJsonPostRequest(server string, path string, body io.Reader) (*http.Request, error) {
	target, err := url.JoinPath(server, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, target, body)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func newPingRequest(server string) (*http.Request, error) {
	return NewGetRequest(server, "/api/v1/ping")
}

func parsePingResponse(resp *http.Response) (string, error) {
	if resp.StatusCode != http.StatusOK {
		return "", parseErrorResponse(resp)
	}

	var res txpb.PingResponse
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return "", fmt.Errorf("parsePingResponse: %w", err)
	}

	return res.Message, nil
}

func newEstimateCostRequest(server string, tx *transactions.Transaction) (*http.Request, error) {
	pbTx := conversion.ConvertToPBTx(tx)
	var bodyReader io.Reader
	buf, err := json.Marshal(txpb.EstimatePriceRequest{Tx: pbTx})
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)

	return NewJsonPostRequest(server, "/api/v1/estimate_price", bodyReader)
}

func parseErrorResponse(resp *http.Response) error {
	// NOTE: here directly use status.Status from googleapis/rpc/status
	var res status.Status
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return err
	}

	msg := res.GetMessage()
	if msg == "" {
		msg = resp.Status
	}

	return errors.New(msg)
}

func parseEstimateCostResponse(resp *http.Response) (*big.Int, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var res txpb.EstimatePriceResponse
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return nil, fmt.Errorf("parseEstimateCostResponse: %w", err)
	}

	bigCost, ok := new(big.Int).SetString(res.Price, 10)
	if !ok {
		return nil, fmt.Errorf("parsePrice failed")
	}

	return bigCost, nil
}

func newBroadcastRequest(server string, tx *transactions.Transaction) (*http.Request, error) {
	pbTx := conversion.ConvertToPBTx(tx)
	var bodyReader io.Reader
	buf, err := json.Marshal(txpb.BroadcastRequest{Tx: pbTx})
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)

	return NewJsonPostRequest(server, "/api/v1/broadcast", bodyReader)
}

func parseBroadcastResponse(resp *http.Response) ([]byte, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var res txpb.BroadcastResponse
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return nil, fmt.Errorf("parseBroadcastResponse: %w", err)
	}

	return res.TxHash, nil
}

func newGetAccountRequest(server string, publicKey []byte, _ types.AccountStatus) (*http.Request, error) {
	// TODO: change proto HTTP option to add a query parameter `status`
	pk := url.PathEscape(base64.URLEncoding.EncodeToString(publicKey))
	return NewGetRequest(server, fmt.Sprintf("/api/v1/accounts/%s", pk))
}

func parseGetAccountResponse(resp *http.Response) (*types.Account, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var res getAccountResponse // weird, server respond with a `string` field nonce
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return nil, fmt.Errorf("parseGetAccountResponse: %w", err)
	}

	bigBalance, ok := new(big.Int).SetString(res.Account.Balance, 10)
	if !ok {
		return nil, fmt.Errorf("parseBalance failed")

	}

	acc := types.Account{
		Identifier: res.Account.Identifier,
		Balance:    bigBalance,
		Nonce:      res.Account.Nonce,
	}

	return &acc, nil
}

func newTxQueryRequest(server string, txHash []byte) (*http.Request, error) {
	txQueryReq := txpb.TxQueryRequest{
		TxHash: txHash,
	}

	var bodyReader io.Reader
	buf, err := json.Marshal(&txQueryReq)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)

	return NewJsonPostRequest(server, "/api/v1/tx_query", bodyReader)

	// NOTE: should make tx_query a GET request?
	//h := url.PathEscape(base64.URLEncoding.EncodeToString(txHash))
	//return NewGetRequest(server, fmt.Sprintf("/api/v1/tx_query/%s", h))
}

// parseTxQueryResponse parses the response from tx_query endpoint
// All returned fields from resp are preserved
func parseTxQueryResponse(resp *http.Response) (*transactions.TcTxQueryResponse, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var res txQueryResponse
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return nil, fmt.Errorf("parseTxQueryResponse: %w", err)
	}

	return conversion.ConvertFromPBTxQueryResp((*txpb.TxQueryResponse)(&res))
}

func newListDatabasesRequest(server string, publicKey []byte) (*http.Request, error) {
	pk := url.PathEscape(base64.URLEncoding.EncodeToString(publicKey))
	return NewGetRequest(server, fmt.Sprintf("/api/v1/%s/databases", pk))
}

func parseListDatabasesResponse(resp *http.Response) ([]string, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var res txpb.ListDatabasesResponse
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return nil, fmt.Errorf("parseListDatabasesResponse: %w", err)
	}

	return res.Databases, nil
}

func newActionCallRequest(server string, msg *transactions.CallMessage) (*http.Request, error) {
	var sender []byte
	if msg.Sender != nil {
		sender = msg.Sender
	}

	callReq := &txpb.CallRequest{
		Body: &txpb.CallRequest_Body{
			Description: msg.Body.Description,
			Payload:     msg.Body.Payload,
		},
		Signature:     conversion.ConvertToPBCryptoSignature(msg.Signature),
		Sender:        sender,
		Serialization: msg.Serialization.String(),
	}

	var bodyReader io.Reader
	buf, err := json.Marshal(callReq)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)

	return NewJsonPostRequest(server, "/api/v1/call", bodyReader)
}

func parseActionCallResponse(resp *http.Response) ([]map[string]any, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var res txpb.CallResponse
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return nil, fmt.Errorf("parseActionCallResponse: %w", err)
	}

	var result []map[string]any
	err = json.Unmarshal(res.Result, &result)
	if err != nil {
		return nil, fmt.Errorf("parseActionCallResponse: %w", err)
	}

	return result, nil
}

func newDBQueryRequest(server string, dbid string, query string) (*http.Request, error) {
	queryReq := &txpb.QueryRequest{
		Dbid:  dbid,
		Query: query,
	}

	var bodyReader io.Reader
	buf, err := json.Marshal(queryReq)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)

	return NewJsonPostRequest(server, "/api/v1/query", bodyReader)
}

func parseDBQueryResponse(resp *http.Response) ([]map[string]any, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var res txpb.QueryResponse
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return nil, fmt.Errorf("parseDBQueryResponse: %w", err)
	}

	var result []map[string]any
	err = json.Unmarshal(res.Result, &result)
	if err != nil {
		return nil, fmt.Errorf("parseDBQueryResponse: %w", err)
	}

	return result, nil
}

func newGetSchemaRequest(server string, dbid string) (*http.Request, error) {
	dbid = url.PathEscape(dbid)
	return NewGetRequest(server, fmt.Sprintf("/api/v1/databases/%s/schema", dbid))
}

func parseGetSchemaResponse(resp *http.Response) (*transactions.Schema, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var res txpb.GetSchemaResponse
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return nil, fmt.Errorf("parseGetSchemaResponse: %w", err)
	}

	return conversion.ConvertFromPBSchema(res.Schema), nil
}

func newValidatorJoinStatusRequest(server string, publicKey []byte) (*http.Request, error) {
	pk := url.PathEscape(base64.URLEncoding.EncodeToString(publicKey))
	return NewGetRequest(server, fmt.Sprintf("/api/v1/validator_join_status/%s", pk))
}

func parseValidatorJoinStatusResponse(resp *http.Response, publicKey []byte) (*types.JoinRequest, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var res validatorJoinsStatusResponse
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return nil, fmt.Errorf("parseValidatorJoinStatusResponse: %w", err)
	}

	jq := conversion.ConvertFromPBJoinRequest((*txpb.ValidatorJoinStatusResponse)(&res))
	jq.Candidate = publicKey
	return jq, nil
}

func newCurrentValidatorsRequest(server string) (*http.Request, error) {
	return NewGetRequest(server, "/api/v1/current_validators")
}

func parseCurrentValidatorsResponse(resp *http.Response) ([]*types.Validator, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var res currentValidatorsResponse
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return nil, fmt.Errorf("parseCurrentValidatorsResponse: %w", err)
	}

	vals := make([]*types.Validator, len(res.Validators))
	for i, vi := range res.Validators {
		vals[i] = &types.Validator{
			PubKey: vi.Pubkey,
			Power:  vi.Power,
		}
	}
	return vals, nil
}

func newChainInfoRequest(server string) (*http.Request, error) {
	return NewGetRequest(server, "/api/v1/chain_info")
}

func parseChainInfoResponse(resp *http.Response) (*types.ChainInfo, error) {
	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp)
	}

	var res chainInfoResponse
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return nil, fmt.Errorf("parseChainInfoResponse: %w", err)
	}

	info := types.ChainInfo{
		ChainID:     res.ChainId,
		BlockHeight: res.Height,
		BlockHash:   res.Hash,
	}

	return &info, nil
}

func newVerifySignatureRequest(server string, sender []byte, sig *auth.Signature,
	msg []byte) (*http.Request, error) {
	req := &txpb.VerifySignatureRequest{
		Signature: &txpb.Signature{
			SignatureBytes: sig.Signature,
			SignatureType:  sig.Type,
		},
		Sender: sender,
		Msg:    msg,
	}

	var bodyReader io.Reader
	buf, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	bodyReader = bytes.NewReader(buf)

	return NewJsonPostRequest(server, "/api/v1/verify_signature", bodyReader)
}

// parseVerifySignatureResponse parses the response from verify_signature endpoint.
// An ErrInvalidSignature is returned if the signature is invalid.
func parseVerifySignatureResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp)
	}

	var res txpb.VerifySignatureResponse
	err := unmarshalResponse(resp.Body, &res)
	if err != nil {
		return fmt.Errorf("parseVerifySignatureResponse: %w", err)
	}

	// caller can tell if signature is valid
	if !res.Valid {
		return fmt.Errorf("%w: %s", client.ErrInvalidSignature, res.Error)
	}

	return nil
}
