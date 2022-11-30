package schema

import (
	"fmt"
	"sort"
)

type DefinedQuery interface {
	Name() string
	Type() QueryType
	//Prepare() PreparedStatement
}

type DefinedQueries struct {
	inserts map[string]*InsertDef
	updates map[string]*UpdateDef
	deletes map[string]*DeleteDef

	statements map[string]PreparedStatement

	all []string
}

func (q *DefinedQueries) ListAll() []string {
	if q.all != nil {
		return q.all
	}

	var queries []string
	for k := range q.inserts {
		queries = append(queries, k)
	}
	for k := range q.updates {
		queries = append(queries, k)
	}
	for k := range q.deletes {
		queries = append(queries, k)
	}

	sort.Strings(queries)
	q.all = queries

	return queries
}

func (q *DefinedQueries) Find(name string) (DefinedQuery, error) {
	i, ok := q.inserts[name]
	if ok {
		return i, nil
	}

	u, ok := q.updates[name]
	if ok {
		return u, nil
	}

	d, ok := q.updates[name]
	if ok {
		return d, nil
	}

	return nil, fmt.Errorf("query not found: %s", name)
}

type defined_query_marshalled struct {
	Type    string    `yaml:"type"`
	Table   string    `yaml:"table"`
	Columns ColumnMap `yaml:"columns"`
	IfMatch ColumnMap `yaml:"if-match"`
}

func (q *DefinedQueries) MarshalYAML() (interface{}, error) {
	if q == nil {
		return nil, nil
	}

	var m map[string]defined_query_marshalled

	if q.inserts != nil || len(q.inserts) == 0 {
		m = make(map[string]defined_query_marshalled)
		for name, query := range q.inserts {
			m[name] = defined_query_marshalled{
				Type:    "create",
				Columns: query.columns,
			}
		}
	}

	if q.updates != nil || len(q.updates) == 0 {
		if m == nil {
			m = make(map[string]defined_query_marshalled)
		}
		for name, query := range q.updates {
			m[name] = defined_query_marshalled{
				Type:    "update",
				Columns: query.columns,
				IfMatch: query.ifMatch,
			}
		}
	}

	if q.deletes != nil || len(q.deletes) == 0 {
		if m == nil {
			m = make(map[string]defined_query_marshalled)
		}
		for name, query := range q.deletes {
			m[name] = defined_query_marshalled{
				Type:    "delete",
				IfMatch: query.ifMatch,
			}
		}
	}

	return m, nil
}

func (q *DefinedQueries) UnmarshalYAML(unmarshal func(interface{}) error) error {
	m := make(map[string]defined_query_marshalled)

	if err := unmarshal(&m); err != nil {
		return err
	}

	if len(m) == 0 {
		return nil
	}

	if q == nil {
		*q = DefinedQueries{}
	}

	if q.inserts == nil {
		q.inserts = make(map[string]*InsertDef)
		q.updates = make(map[string]*UpdateDef)
		q.deletes = make(map[string]*DeleteDef)
	}

	for name, query := range m {
		switch query.Type {
		case "create":
			q.addCreate(name, query.Columns)
		case "update":
			q.addUpdate(name, query.Columns, query.IfMatch)
		case "delete":
			q.addDelete(name, query.IfMatch)
		default:
			return fmt.Errorf("unknown query type: %s", query.Type)
		}
	}

	return nil
}

func (q *DefinedQueries) addCreate(name string, columns ColumnMap) {
	q.inserts[name] = &InsertDef{
		name:    name,
		columns: columns,
	}
}

func (q *DefinedQueries) addUpdate(name string, columns ColumnMap, ifMatch ColumnMap) {
	q.updates[name] = &UpdateDef{
		name:    name,
		columns: columns,
		ifMatch: ifMatch,
	}
}

func (q *DefinedQueries) addDelete(name string, ifMatch ColumnMap) {
	q.deletes[name] = &DeleteDef{
		name:    name,
		ifMatch: ifMatch,
	}
}
