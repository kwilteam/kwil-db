package pricer

import (
	"context"
	"fmt"
	"kwil/internal/usecases/executor"
	"kwil/pkg/accounts"
	"kwil/pkg/databases/executables"
	querytype "kwil/pkg/databases/spec"
	"kwil/pkg/pricing"
	"kwil/pkg/utils/serialize"
	"math"
	"strconv"
)

type QueryPriceEstimator interface {
	Estimate() (string, error)
}

type qpestimator struct {
	params *pricing.Params
}

const (
	POSTGRES_BLOCKSIZE = 8192
)

var W_I = [...]float64{20000000000000.0, 20000000000000.0, 20000000000000.0}
var W_U = [...]float64{20000000000000.0, 20000000000000.0, 20000000000000.0}

func NewQueryPriceEstimator(ctx context.Context, tx *accounts.Transaction, exec executor.Executor) (QueryPriceEstimator, error) {
	// get executionBody
	executionBody, err := serialize.Deserialize[*executables.ExecutionBody](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload of type ExecutionBody: %w", err)
	}

	// Execute the equivalent select query or explain query to get the cost estimation info
	pricingparams, err := exec.GetQueryCostEstimationInfo(ctx, executionBody, tx.Sender)
	if err != nil {
		return nil, fmt.Errorf("failed to get query cost estimation info: %w", err)
	}

	return &qpestimator{
		params: pricingparams,
	}, nil
}

func (q *qpestimator) Estimate() (string, error) {
	switch q.params.Q {
	default:
		return "", fmt.Errorf(`invalid query type "%d"`, q.params.Q)
	case querytype.INSERT:
		return strconv.Itoa(int(q.estimateInsert(q.params))), nil
	case querytype.UPDATE:
		return strconv.Itoa(int(q.estimateUpdate(q.params))), nil
	case querytype.DELETE:
		return strconv.Itoa(int(q.estimateDelete(q.params))), nil
	}

}

func (q *qpestimator) estimateInsert(p *pricing.Params) float64 {
	return float64(p.I)*q.f0(float64(p.T))*W_I[0] + float64(q.f1(p.T))*W_I[0] + float64(p.S)*W_I[0]
}

func (q *qpestimator) estimateUpdate(p *pricing.Params) float64 {
	w := q.getWhereCost(p)
	return float64(w)*W_U[0] + float64(p.I)*math.Log(float64(p.T))*W_U[1] + float64(p.U*p.S)*W_U[2]
}

func (q *qpestimator) estimateDelete(p *pricing.Params) float64 {
	w := q.getWhereCost(p)
	return float64(w)*W_U[0] + float64(p.I)*math.Log(float64(p.T))*W_U[1] + float64(p.U*p.S)*W_U[2]
}

func (q *qpestimator) getWhereCost(p *pricing.Params) float64 {
	return float64(len(p.W)) * q.f0(float64(p.T))
}

func (q *qpestimator) f0(t float64) float64 {
	return math.Log(t)
}

func (q *qpestimator) f1(t int64) float64 {
	//return t
	return float64(t) / float64(POSTGRES_BLOCKSIZE)
}
