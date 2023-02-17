package pricer

import (
	"context"
	"fmt"
	querytype "kwil/x/execution"
	"kwil/x/execution/executor"
	"kwil/x/pricing"
	"kwil/x/types/databases/clean"
	"kwil/x/types/execution"
	"kwil/x/types/execution/convert"
	txTypes "kwil/x/types/transactions"
	"kwil/x/utils/serialize"
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

func NewQueryPriceEstimator(ctx context.Context, tx *txTypes.Transaction, exec executor.Executor) (QueryPriceEstimator, error) {
	// get executionBody
	executionBody, err := serialize.Deserialize[*execution.ExecutionBody[[]byte]](tx.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload of type ExecutionBody: %w", err)
	}

	clean.Clean(&executionBody)

	convExecutionBody, err := convert.Bytes.BodyToKwilAny(executionBody)
	if err != nil {
		return nil, fmt.Errorf("failed to convert execution body to kwil any: %w", err)
	}

	// Execute the equivalent select query or explain query to get the cost estimation info
	pricingparams, err := exec.GetQueryCostEstimationInfo(ctx, convExecutionBody, tx.Sender)
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
	return float64(p.I)*q.f0(float64(p.T)) + float64(q.f1(p.T)) + float64(p.S)
}

func (q *qpestimator) estimateUpdate(p *pricing.Params) float64 {
	w := q.getWhereCost(p)
	return float64(w) + float64(p.I)*math.Log(float64(p.T)) + float64(p.T*p.S)
}

func (q *qpestimator) estimateDelete(p *pricing.Params) float64 {
	w := q.getWhereCost(p)
	return float64(w) + float64(p.I)*math.Log(float64(p.T)) + float64(p.T*p.S)
}

func (q *qpestimator) getWhereCost(p *pricing.Params) float64 {
	return float64(len(p.W)) * q.f0(float64(p.T))
}

func (q *qpestimator) f0(t float64) float64 {
	return math.Log(t)
}

func (q *qpestimator) f1(t int64) int64 {
	//return t
	return t / POSTGRES_BLOCKSIZE
}
