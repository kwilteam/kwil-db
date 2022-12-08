package manager

import "kwil/x/proto/apipb"

/*
	These are made to be an intermediary structure between the cache and GRPC.
	In case GRPC changes, it should be easy to go from this to any new protobuf.
	A better way to do this would likely be to store the db metadata in postgres.
*/

type ExportedDB struct {
	Name        string
	Owner       string
	DefaultRole string
	Tables      []*ExportedTable
	Queries     []*ExportedQuery
	Roles       []*ExportedRole
	Indexes     []*ExportedIndex
}

type ExportedTable struct {
	Name    string
	Columns []*ExportedColumn
}

type ExportedColumn struct {
	Name       string
	Type       string
	Attributes []*ExportedAttribute
}

type ExportedQuery struct {
	Name      string
	Statement string
	Inputs    []*ExportedInput
	Defaults  []*ExportedDefault
}

type ExportedRole struct {
	Name    string
	Queries []string
}

type ExportedIndex struct {
	Name   string
	Table  string
	Column string
	Using  string
}

type ExportedInput struct {
	Name    string
	Type    string
	Ordinal int
}

type ExportedDefault struct {
	Name    string
	Type    string
	Value   string
	Ordinal int
}

type ExportedAttribute struct {
	Name  string
	Value string
}

func (e *ExportedDB) AsProtobuf() *apipb.Metadata {
	db := apipb.Metadata{
		Name:        e.Name,
		Owner:       e.Owner,
		DefaultRole: e.DefaultRole,
		Tables:      []*apipb.Table{},
		Queries:     []*apipb.Query{},
		Roles:       []*apipb.Role{},
		Indexes:     []*apipb.Index{},
	}
	for _, table := range e.Tables {
		db.Tables = append(db.Tables, table.AsProtobuf())
	}

	for _, query := range e.Queries {
		db.Queries = append(db.Queries, query.AsProtobuf())
	}

	for _, role := range e.Roles {
		db.Roles = append(db.Roles, role.AsProtobuf())
	}

	for _, index := range e.Indexes {
		db.Indexes = append(db.Indexes, index.AsProtobuf())
	}

	return &db
}

func (e *ExportedTable) AsProtobuf() *apipb.Table {
	table := &apipb.Table{
		Name:    e.Name,
		Columns: []*apipb.Column{},
	}
	for _, column := range e.Columns {
		table.Columns = append(table.Columns, column.AsProtobuf())
	}
	return table
}

func (e *ExportedColumn) AsProtobuf() *apipb.Column {
	var attributes []*apipb.Attribute
	for _, attr := range e.Attributes {
		attributes = append(attributes, &apipb.Attribute{
			Name:  attr.Name,
			Value: attr.Value,
		})
	}

	column := &apipb.Column{
		Name:       e.Name,
		Type:       e.Type,
		Attributes: attributes,
	}
	return column
}

func (e *ExportedQuery) AsProtobuf() *apipb.Query {
	query := &apipb.Query{
		Name:      e.Name,
		Statement: e.Statement,
		Inputs:    []*apipb.Input{},
		Defaults:  []*apipb.DefaultInput{},
	}
	for _, input := range e.Inputs {
		query.Inputs = append(query.Inputs, input.AsProtobuf())
	}
	for _, def := range e.Defaults {
		query.Defaults = append(query.Defaults, def.AsProtobuf())
	}
	return query
}

func (e *ExportedInput) AsProtobuf() *apipb.Input {
	input := &apipb.Input{
		Name:    e.Name,
		Type:    e.Type,
		Ordinal: int32(e.Ordinal),
	}
	return input
}

func (e *ExportedDefault) AsProtobuf() *apipb.DefaultInput {
	def := &apipb.DefaultInput{
		Name:    e.Name,
		Type:    e.Type,
		Value:   e.Value,
		Ordinal: int32(e.Ordinal),
	}
	return def
}

func (e *ExportedRole) AsProtobuf() *apipb.Role {
	role := &apipb.Role{
		Name:    e.Name,
		Queries: e.Queries,
	}
	return role
}

func (e *ExportedIndex) AsProtobuf() *apipb.Index {
	index := &apipb.Index{
		Name:   e.Name,
		Table:  e.Table,
		Column: e.Column,
		Using:  e.Using,
	}
	return index
}
