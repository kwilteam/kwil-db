package client

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/cstockton/go-conv"
	txpb "github.com/kwilteam/kwil-db/api/protobuf/tx/v1"
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/tx"
)

func (c *Client) Call(ctx context.Context, req *tx.CallActionMessage) ([]map[string]any, error) {

	scalarMap, err := paramsToScalar(req.Payload.Params)
	if err != nil {
		return nil, fmt.Errorf("failed to convert params to scalar: %w", err)
	}

	grpcMsg := &txpb.CallRequest{
		Payload: &txpb.CallPayload{
			Dbid:   req.Payload.DBID,
			Action: req.Payload.Action,
			Args:   scalarMap,
		},
		Signature: convertActionSignature(req.Signature),
		Sender:    req.Sender,
	}

	res, err := c.txClient.Call(ctx, grpcMsg)

	if err != nil {
		return nil, fmt.Errorf("failed to call: %w", err)
	}

	var result []map[string]any
	err = json.Unmarshal(res.Result, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return result, nil
}

func paramsToScalar(oldParams map[string]any) (map[string]*txpb.ScalarValue, error) {
	newParams := make(map[string]*txpb.ScalarValue)
	for k, v := range oldParams {

		var scalarVal *txpb.ScalarValue
		switch concreteVal := v.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool:
			int64Val, err := conv.Int64(concreteVal)
			if err != nil {
				return nil, fmt.Errorf("failed to convert int value: %w", err)
			}

			scalarVal = &txpb.ScalarValue{
				Value: &txpb.ScalarValue_IntValue{
					IntValue: int64Val,
				},
			}
		case string:
			scalarVal = &txpb.ScalarValue{
				Value: &txpb.ScalarValue_StringValue{
					StringValue: concreteVal,
				},
			}
		default:
			return nil, fmt.Errorf("unknown value type '%s' for param %s", reflect.TypeOf(v).String(), k)
		}

		newParams[k] = scalarVal
	}

	return newParams, nil
}

func convertActionSignature(oldSig *crypto.Signature) *txpb.Signature {
	if oldSig == nil {
		return &txpb.Signature{}
	}

	newSig := &txpb.Signature{
		SignatureBytes: oldSig.Signature,
		SignatureType:  oldSig.Type.Int32(),
	}

	return newSig
}
