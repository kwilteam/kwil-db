package graphql

import (
	"encoding/json"
	"net/http"
	"strings"
)

func isMutation(query string) bool {
	// NOTE: enough to correctly block most mutations
	var operations []string
	rightBracket := -1
	opens := 0
	for i, c := range query {
		if c == '}' {
			opens -= 1
			if opens == 0 {
				rightBracket = i
			}
		}

		if c == '{' {
			if opens == 0 {
				operations = append(operations, query[rightBracket+1:i])
			}
			opens += 1
		}
	}

	for _, op := range operations {
		if strings.Contains(op, "mutation") {
			return true
		}
	}
	return false
}

func jsonError(w http.ResponseWriter, err error, code int) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(err)
}
