package schema

import (
	"fmt"
	"sort"
)

type DefinedQuery interface {
	Type() QueryType
	Prepare(*Database) (*ExecutableQuery, error)
}

type DefinedQueries struct {
	Inserts map[string]*InsertDef
	Updates map[string]*UpdateDef
	Deletes map[string]*DeleteDef

	statements map[string]PreparedStatement

	all []string
}

func (q *DefinedQueries) ListAll() []string {
	if q.all != nil {
		return q.all
	}

	var queries []string
	for k := range q.Inserts {
		queries = append(queries, k)
	}
	for k := range q.Updates {
		queries = append(queries, k)
	}
	for k := range q.Deletes {
		queries = append(queries, k)
	}

	sort.Strings(queries)
	q.all = queries

	return queries
}

// GetAll is like ListAll, but returns the queries themselves
func (q *DefinedQueries) GetAll() map[string]DefinedQuery {
	queries := make(map[string]DefinedQuery)
	for name, query := range q.Inserts {
		queries[name] = query
	}
	for name, query := range q.Updates {
		queries[name] = query
	}
	for name, query := range q.Deletes {
		queries[name] = query
	}
	return queries
}

func (q *DefinedQueries) Find(name string) (DefinedQuery, error) {
	i, ok := q.Inserts[name]
	if ok {
		return i, nil
	}

	u, ok := q.Updates[name]
	if ok {
		return u, nil
	}

	d, ok := q.Deletes[name]
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

	if q.Inserts != nil || len(q.Inserts) == 0 {
		m = make(map[string]defined_query_marshalled)
		for name, query := range q.Inserts {
			m[name] = defined_query_marshalled{
				Type:    "create",
				Columns: query.Columns,
			}
		}
	}

	if q.Updates != nil || len(q.Updates) == 0 {
		if m == nil {
			m = make(map[string]defined_query_marshalled)
		}
		for name, query := range q.Updates {
			m[name] = defined_query_marshalled{
				Type:    "update",
				Columns: query.Columns,
				Where:   query.Where,
			}
		}
	}

	if q.Deletes != nil || len(q.Deletes) == 0 {
		if m == nil {
			m = make(map[string]defined_query_marshalled)
		}
		for name, query := range q.Deletes {
			m[name] = defined_query_marshalled{
				Type:  "delete",
				Where: query.Where,
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

	if q.Inserts == nil {
		q.Inserts = make(map[string]*InsertDef)
		q.Updates = make(map[string]*UpdateDef)
		q.Deletes = make(map[string]*DeleteDef)
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
	q.Inserts[name] = &InsertDef{
		Name:    name,
		Table:   table,
		Columns: columns,
	}
}

func (q *DefinedQueries) addUpdate(name, table string, columns ColumnMap, where []where_predicate) {
	q.Updates[name] = &UpdateDef{
		Name:    name,
		Table:   table,
		Columns: columns,
		Where:   where,
	}
}

func (q *DefinedQueries) addDelete(name, table string, where []where_predicate) {
	q.Deletes[name] = &DeleteDef{
		Name:  name,
		Table: table,
		Where: where,
	}
}
