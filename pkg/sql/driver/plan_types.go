package driver

import (
	"fmt"
	poly "github.com/kwilteam/kwil-db/pkg/utils/numbers/polynomial"
	"strings"

	"github.com/cstockton/go-conv"
)

type QueryAction uint8

const (
	UNDETECTED QueryAction = iota
	SCAN
	SCAN_INDEX
	SEARCH_INDEX
	SCALAR_SUBQUERY
	MATERIALIZE
	CO_ROUTINE
	OTHER
)

func (q *QueryAction) String() string {
	switch *q {
	case UNDETECTED:
		return "UNDETECTED"
	case SCAN:
		return "SCAN"
	case SCAN_INDEX:
		return "SCAN_INDEX"
	case SEARCH_INDEX:
		return "SEARCH_INDEX"
	case OTHER:
		return "OTHER"
	default:
		return "UNKNOWN"
	}
}

type Action interface {
	Type() QueryAction
	Polynomial() poly.Expression
}

func parseAction(detail string) (Action, error) {
	acc := detectType(detail)
	switch acc {
	case SCAN:
		relation, err := parseScanRelation(detail)
		if err != nil {
			return nil, err
		}

		return &Scan{
			Table: relation,
		}, nil
	case SCAN_INDEX:
		relation, err := parseScanRelation(detail)
		if err != nil {
			return nil, err
		}

		return &ScanIndex{
			Table: relation,
		}, nil
	case SEARCH_INDEX:
		relation, err := parseSearchRelation(detail)
		if err != nil {
			return nil, err
		}

		return &SearchIndex{
			Table: relation,
		}, nil
	case SCALAR_SUBQUERY:
		id, err := parseScalarSubquery(detail)
		if err != nil {
			return nil, err
		}

		return &ScalarSubquery{
			SubqueryId: id,
		}, nil
	case MATERIALIZE:
		name, err := parseMaterializeRelation(detail)
		if err != nil {
			return nil, err
		}

		return &Materialize{
			Name: name,
		}, nil
	case CO_ROUTINE:
		name, anon, err := parseCoRoutineRelation(detail)
		if err != nil {
			return nil, err
		}

		return &CoRoutine{
			Name:      name,
			Anonymous: anon,
		}, nil
	}

	return &Other{}, nil
}

type Scan struct {
	Table string
}

func (s *Scan) Type() QueryAction {
	return SCAN
}

func (s *Scan) Polynomial() poly.Expression {
	return poly.WeightedVar(s.Table, poly.NewFloat(SCAN_WEIGHT))
}

type ScanIndex struct {
	Table string
}

func (s *ScanIndex) Type() QueryAction {
	return SCAN_INDEX
}

func (s *ScanIndex) Polynomial() poly.Expression {
	return poly.Log(poly.WeightedVar(s.Table, poly.NewFloat(SCAN_INDEX_WEIGHT)))
}

type SearchIndex struct {
	Table string
}

func (s *SearchIndex) Type() QueryAction {
	return SEARCH_INDEX
}

func (s *SearchIndex) Polynomial() poly.Expression {
	return poly.Log(poly.WeightedVar(s.Table, poly.NewFloat(SEARCH_WEIGHT)))
}

type ScalarSubquery struct {
	SubqueryId int64
}

func (s *ScalarSubquery) Type() QueryAction {
	return SCALAR_SUBQUERY
}

func (s *ScalarSubquery) Polynomial() poly.Expression {
	return poly.NewWeight(1)
}

type Materialize struct {
	Name string
}

func (m *Materialize) Type() QueryAction {
	return MATERIALIZE
}

func (m *Materialize) Polynomial() poly.Expression {
	return poly.NewWeight(1)
}

type CoRoutine struct {
	// the name of the co-routine, usually either a CTE name or a subquery
	Name string

	// the type of the co-routine.  subquery is usually anonymous
	Anonymous bool
}

func (c *CoRoutine) Type() QueryAction {
	return CO_ROUTINE
}

func (c *CoRoutine) Polynomial() poly.Expression {
	return poly.NewWeight(1)
}

type Other struct {
}

func (o *Other) Type() QueryAction {
	return OTHER
}

func (o *Other) Polynomial() poly.Expression {
	return poly.NewWeight(1)
}

func detectType(s string) QueryAction {
	tokens := strings.Split(s, " ")
	switch tokens[0] {
	case "SCAN":
		if len(tokens) >= 3 && tokens[2] == "USING" {
			return SCAN_INDEX
		}
		return SCAN
	case "SEARCH":
		return SEARCH_INDEX
	case "SCALAR":
		return SCALAR_SUBQUERY
	case "CORRELATED":
		return SCALAR_SUBQUERY // not quite right, but we can't actually determine correlation from query plan
	case "MATERIALIZE":
		return MATERIALIZE
	case "CO-ROUTINE":
		return CO_ROUTINE
	}

	return OTHER
}

// parses the relation being scanned from the query plan
func parseScanRelation(s string) (string, error) {
	tokens := strings.Split(s, " ")
	if len(tokens) < 2 {
		return "", fmt.Errorf("invalid scan relation: %s", s)
	}
	return tokens[1], nil
}

// parses the relation being searched from the query plan
func parseSearchRelation(s string) (string, error) {
	tokens := strings.Split(s, " ")
	if len(tokens) < 2 {
		return "", fmt.Errorf("invalid search relation: %s", s)
	}
	return tokens[1], nil
}

// parses the anonymous subquery id number from the query plan
func parseScalarSubquery(s string) (int64, error) {
	tokens := strings.Split(s, " ")
	if len(tokens) == 0 {
		return 0, fmt.Errorf("invalid scalar subquery: %s", s)
	}

	return conv.Int64(tokens[len(tokens)-1])
}

// parses the name of the materialized relation from the query plan
func parseMaterializeRelation(s string) (string, error) {
	tokens := strings.Split(s, " ")
	if len(tokens) < 2 {
		return "", fmt.Errorf("invalid materialize relation: %s", s)
	}
	return tokens[1], nil
}

func parseCoRoutineRelation(s string) (name string, anonymous bool, err error) {
	tokens := strings.Split(s, " ")
	if len(tokens) < 2 {
		return "", false, fmt.Errorf("invalid co-routine relation: %s", s)
	}

	name = tokens[1]
	// check if string starts with parenthesis
	if strings.HasPrefix(name, "(") {
		anonymous = true
		name = strings.TrimPrefix(name, "(")
		name = strings.TrimSuffix(name, ")")
	}

	return name, anonymous, nil
}
