package query_planner

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

// RelationName is the name of a relation in the catalog.
// It can be an unqualified name (table) or a fully qualified name (schema.table).
type RelationName string

func (t RelationName) String() string {
	return string(t)
}

func (t RelationName) Segments() []string {
	return strings.Split(string(t), ".") // schema.relation
}

func (t RelationName) IsQualified() bool {
	return strings.ContainsRune(string(t), '.') // len(t.Segments()) > 1
}

func (t RelationName) Parse() (*datatypes.TableRef, error) {
	segments := t.Segments()
	switch len(segments) {
	case 1:
		return &datatypes.TableRef{Table: segments[0]}, nil
	case 2:
		return &datatypes.TableRef{Namespace: segments[0], Table: segments[1]}, nil
	default:
		return nil, fmt.Errorf("invalid relation name: %s", t)
	}
}
