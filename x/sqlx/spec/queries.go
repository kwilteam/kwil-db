package types

import "fmt"

const (
	MAX_PARAM_COUNT = 50
	MAX_WHERE_COUNT = 3
)

type QueryType int

// Queries
const (
	INVALID_QUERY QueryType = iota
	INSERT
	UPDATE
	SELECT
	DELETE
)

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

// ConvertQuery converts a string to a QueryType
func (v *validation) ConvertQueryType(s string) (QueryType, error) {
	switch s {
	case "insert":
		return INSERT, nil
	case "update":
		return UPDATE, nil
	case "select":
		return SELECT, nil
	case "delete":
		return DELETE, nil
	}
	return INVALID_QUERY, fmt.Errorf("unknown query type")
}
