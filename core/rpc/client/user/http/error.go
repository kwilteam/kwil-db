package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/kwilteam/kwil-db/core/rpc/client"
	httpTx "github.com/kwilteam/kwil-db/core/rpc/http/tx"
	txpb "github.com/kwilteam/kwil-db/core/rpc/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/core/types/transactions"

	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

type gatewayErrResponse struct {
	Error string `json:"error"`
}

func parseErrorResponse(respTxt []byte) error {
	// NOTE: here directly use status.Status from googleapis/rpc/status
	var res status.Status
	err := json.Unmarshal(respTxt, &res)
	if err != nil {
		return err
	}

	// for kgw error response
	if res.GetCode() == 0 { // if it's grpc status, it should have a code, code 0 is not error code
		// NOTE: this could be removed once #623 is merged
		// try to parse kgw error first
		var kgwErr gatewayErrResponse
		err := json.Unmarshal(respTxt, &kgwErr)
		if err == nil {
			return errors.Join(errors.New(kgwErr.Error))
		}
	}

	rpcErr := &client.RPCError{
		Msg:  res.GetMessage(),
		Code: res.GetCode(),
	}

	switch res.Code {
	case int32(codes.NotFound):
		return errors.Join(client.ErrNotFound, rpcErr)
	case int32(codes.PermissionDenied), int32(codes.Unauthenticated): // these have different meaning, but are handled via auth
		return errors.Join(client.ErrUnauthorized, rpcErr)
	default:
	}

	return rpcErr
}

func wrapResponseError(err error, res *http.Response) error {
	if res != nil {
		// Wrap certain errors in our own types.
		switch res.StatusCode {
		case http.StatusUnauthorized:
			err = errors.Join(err, client.ErrUnauthorized)
		case http.StatusNotFound:
			err = errors.Join(err, client.ErrNotFound)
		}
		// Continue to attempt decoding swagger error's response body.
	}

	// NOTE: for kgw error response, it happens to have `error` field in the response body
	// it's also recognized as a httpTx.GenericSwaggerError
	if swaggerErr, ok := err.(httpTx.GenericSwaggerError); ok {
		body := swaggerErr.Body()
		if body != nil {
			err = errors.Join(err, parseErrorResponse(body))
		}
	}

	return err
}

// parseBroadcastError parses the response body from a broadcast error.
// It returns true if the error was parsed successfully, false otherwise.
func parseBroadcastError(respTxt []byte) (bool, error) {
	var protoStatus status.Status
	err := protojson.Unmarshal(respTxt, &protoStatus) // jsonpb is deprecated, otherwise we could use the resp.Body directly
	if err != nil {
		if err = json.Unmarshal(respTxt, &protoStatus); err != nil {
			return false, err
		}
	}

	// for kgw error response
	if protoStatus.GetCode() == 0 { // if it's grpc status, it should have a code, code 0 is not error code
		// NOTE: this could be removed once #623 is merged
		// try to parse kgw error first
		var kgwErr gatewayErrResponse
		err := json.Unmarshal(respTxt, &kgwErr)
		if err == nil {
			return true, errors.Join(errors.New(kgwErr.Error))
		}
	}

	stat := grpcStatus.FromProto(&protoStatus)
	code, message := stat.Code(), stat.Message()
	rpcErr := &client.RPCError{
		Msg:  message,
		Code: int32(code),
	}
	err = rpcErr

	for _, detail := range stat.Details() {
		if bcastErr, ok := detail.(*txpb.BroadcastErrorDetails); ok {
			txCode := transactions.TxCode(bcastErr.Code)
			switch txCode {
			case transactions.CodeWrongChain:
				err = errors.Join(err, transactions.ErrWrongChain)
			case transactions.CodeInvalidNonce:
				err = errors.Join(err, transactions.ErrInvalidNonce)
			case transactions.CodeInvalidAmount:
				err = errors.Join(err, transactions.ErrInvalidAmount)
			case transactions.CodeInsufficientBalance:
				err = errors.Join(err, transactions.ErrInsufficientBalance)
			}

			// Reset the generic code and message in the RPCError with the
			// broadcast-specific details. NOTE: this will overwrite if there
			// are more than one details object, which is not expected.
			rpcErr.Code = int32(txCode)
			rpcErr.Msg = bcastErr.Message
			if bcastErr.Hash != "" { // if there is a tx hash, include it (possibly just executed it)
				rpcErr.Msg += "\nTxHash: " + bcastErr.Hash
			}
		} else { // else unknown details type
			err = errors.Join(err, fmt.Errorf("unrecognized status error detail type %T", detail))
		}
	}

	return true, err
}
