package execution

import "fmt"

type queries struct{}

var Queries = &queries{}

// ConvertQuery converts a string to a QueryType
func (q *queries) ConvertQueryType(s string) (QueryType, error) {
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
	return INVALID_QUERY_TYPE, fmt.Errorf("unknown query type")
}
