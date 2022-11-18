package apisvc

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"kwil/x/crypto"
	"kwil/x/execution"
	"kwil/x/proto/apipb"
)

func (s *Service) Cud(ctx context.Context, req *apipb.CUDRequest) (*apipb.CUDResponse, error) {
	p, err := s.p.GetPrice(ctx)
	if err != nil {
		return nil, err
	}

	// parse fee
	fee, ok := parseBigInt(req.Fee)
	if !ok {
		return nil, fmt.Errorf("invalid fee")
	}

	// check price is enough
	if fee.Cmp(p) < 0 {
		return nil, fmt.Errorf("price is not enough")
	}

	// generate id
	id := cudID(req)

	if id != req.Id {
		return nil, fmt.Errorf("invalid id")
	}

	// check signature
	valid, err := crypto.CheckSignature(req.From, req.Signature, []byte(id))
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, fmt.Errorf("invalid signature")
	}

	// spend funds andthen write data!
	err = s.ds.Spend(req.From, req.Fee)
	if err != nil {
		return nil, err
	}

	return &apipb.CUDResponse{
		TraceId: "",
	}, nil
}

func (s *Service) Read(ctx context.Context, req *apipb.ReadRequest) (*apipb.ReadResponse, error) {
	bi, err := json.Marshal(req.Inputs)
	if err != nil {
		return nil, err
	}

	var ins []execution.Input
	err = json.Unmarshal(bi, &ins)
	if err != nil {
		return nil, err
	}

	res, err := s.e.Read(ctx, req.Owner, req.Database, req.Query, ins)
	if err != nil {
		return nil, err
	}

	qRes := convertResult(res)

	return &apipb.ReadResponse{
		Result: qRes,
	}, nil

}

func convertResultColumn(c *execution.Column) *apipb.ColumnResult {
	return &apipb.ColumnResult{
		Name:  c.Name,
		Value: c.Value.String,
		Type:  convertType(c.Type),
	}
}

func convertResultRow(r *execution.Row) *apipb.Row {
	var cols []*apipb.ColumnResult
	for _, c := range r.Columns {
		cols = append(cols, convertResultColumn(&c))
	}
	return &apipb.Row{
		Columns: cols,
	}
}

func convertResult(r *execution.Result) *apipb.QueryResult {
	var rows []*apipb.Row
	for _, row := range r.Rows {
		rows = append(rows, convertResultRow(&row))
	}
	return &apipb.QueryResult{
		Rows: rows,
	}
}

func parseBigInt(s string) (*big.Int, bool) {
	return new(big.Int).SetString(s, 10)
}
