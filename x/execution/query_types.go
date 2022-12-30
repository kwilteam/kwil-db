package execution

import "fmt"

type QueryType int

// Queries
const (
	INVALID_QUERY QueryType = iota
	SELECT
	INSERT
	UPDATE
	DELETE
	END_QUERY
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
	case SELECT:
		return "select", nil
	case DELETE:
		return "delete", nil
	}
	return "", fmt.Errorf("unknown query type")
}

func (q *QueryType) IsValid() bool {
	return *q > INVALID_QUERY && *q < END_QUERY
}
