package execution

import "fmt"

type QueryType int

// Queries
const (
	INVALID_QUERY_TYPE QueryType = iota
	INSERT
	UPDATE
	DELETE
	SELECT
	END_QUERY_TYPE
)

func (q *QueryType) Int() int {
	return int(*q)
}

func (q *QueryType) String() (string, error) {
	switch *q {
	case INSERT:
		return "insert", nil
	case UPDATE:
		return "update", nil
	case DELETE:
		return "delete", nil
	}
	return "", fmt.Errorf("unknown query type")
}

func (q *QueryType) IsValid() bool {
	return *q > INVALID_QUERY_TYPE && *q < END_QUERY_TYPE
}
