package schema

import (
	"fmt"
	"sort"
)

type DefinedQuery interface {
	Name() string
	Type() QueryType
	Prepare(*Database) (*executableQuery, error)
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

// GetAll is like ListAll, but returns the queries themselves
func (q *DefinedQueries) GetAll() map[string]DefinedQuery {
	queries := make(map[string]DefinedQuery)
	for name, query := range q.inserts {
		queries[name] = query
	}
	for name, query := range q.updates {
		queries[name] = query
	}
	for name, query := range q.deletes {
		queries[name] = query
	}
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

	d, ok := q.deletes[name]
	if ok {
		return d, nil
	}

	return nil, fmt.Errorf("query not found: %s", name)
}

type defined_query_marshalled struct {
	Type    string            `yaml:"type"`
	Table   string            `yaml:"table"`
	Columns ColumnMap         `yaml:"columns"`
	Where   []where_predicate `yaml:"where"`
}

type where_predicate struct {
	Column   string `yaml:"column"`
	Operator string `yaml:"operator"`
	Default  string `yaml:"default"`
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
				Where:   query.where,
			}
		}
	}

	if q.deletes != nil || len(q.deletes) == 0 {
		if m == nil {
			m = make(map[string]defined_query_marshalled)
		}
		for name, query := range q.deletes {
			m[name] = defined_query_marshalled{
				Type:  "delete",
				Where: query.where,
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
			q.addCreate(name, query.Table, query.Columns)
		case "update":
			q.addUpdate(name, query.Table, query.Columns, query.Where)
		case "delete":
			q.addDelete(name, query.Table, query.Where)
		default:
			return fmt.Errorf("unknown query type: %s", query.Type)
		}
	}

	return nil
}

func (q *DefinedQueries) addCreate(name, table string, columns ColumnMap) {
	q.inserts[name] = &InsertDef{
		name:    name,
		table:   table,
		columns: columns,
	}
}

func (q *DefinedQueries) addUpdate(name, table string, columns ColumnMap, where []where_predicate) {
	q.updates[name] = &UpdateDef{
		name:    name,
		table:   table,
		columns: columns,
		where:   where,
	}
}

func (q *DefinedQueries) addDelete(name, table string, where []where_predicate) {
	q.deletes[name] = &DeleteDef{
		name:  name,
		table: table,
		where: where,
	}
}
