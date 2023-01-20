package repository

import (
	"context"
	"kwil/kwil/repository/gen"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
	"strings"
)

type DatabaseBuilder interface {
	CreateDatabase(ctx context.Context, db *databases.DatabaseIdentifier) error
	DropDatabase(ctx context.Context, db *databases.DatabaseIdentifier) error
	GetDatabaseId(ctx context.Context, db *databases.DatabaseIdentifier) (int32, error)
	CreateTable(ctx context.Context, dbId int32, table string) error
	GetTableId(ctx context.Context, dbId int32, table string) (int32, error)
	CreateColumn(ctx context.Context, tableId int32, columnName string, columnType int32) error
	GetColumnId(ctx context.Context, tableId int32, columnName string) (int32, error)
	CreateAttribute(ctx context.Context, columnId int32, attributeType int32, attributeValue anytype.KwilAny) error
	CreateQuery(ctx context.Context, queryName string, tableId int32, queryData []byte) error
	CreateRole(ctx context.Context, dbId int32, roleName string, isDefault bool) error
	ApplyPermissionToRole(ctx context.Context, dbId int32, roleName string, queryName string) error
	CreateIndex(ctx context.Context, tableId int32, indexName string, indexType int32, columns []string) error
}

func (q *queries) CreateDatabase(ctx context.Context, db *databases.DatabaseIdentifier) error {
	return q.gen.CreateDatabase(ctx, &gen.CreateDatabaseParams{
		DbName:         strings.ToLower(db.Name),
		AccountAddress: strings.ToLower(db.Owner),
	})
}

func (q *queries) DropDatabase(ctx context.Context, db *databases.DatabaseIdentifier) error {
	return q.gen.DropDatabase(ctx, &gen.DropDatabaseParams{
		DbName:         strings.ToLower(db.Name),
		AccountAddress: strings.ToLower(db.Owner),
	})
}

func (q *queries) GetDatabaseId(ctx context.Context, db *databases.DatabaseIdentifier) (int32, error) {
	return q.gen.GetDatabaseId(ctx, &gen.GetDatabaseIdParams{
		DbName:         strings.ToLower(db.Name),
		AccountAddress: strings.ToLower(db.Owner),
	})
}

func (q *queries) CreateTable(ctx context.Context, dbId int32, table string) error {
	return q.gen.CreateTable(ctx, &gen.CreateTableParams{
		DbID:      dbId,
		TableName: strings.ToLower(table),
	})
}

func (q *queries) GetTableId(ctx context.Context, dbId int32, table string) (int32, error) {
	return q.gen.GetTableId(ctx, &gen.GetTableIdParams{
		DbID:      dbId,
		TableName: strings.ToLower(table),
	})
}

func (q *queries) CreateColumn(ctx context.Context, tableId int32, columnName string, columnType int32) error {
	return q.gen.CreateColumn(ctx, &gen.CreateColumnParams{
		TableID:    tableId,
		ColumnName: strings.ToLower(columnName),
		ColumnType: columnType,
	})
}

func (q *queries) GetColumnId(ctx context.Context, tableId int32, columnName string) (int32, error) {
	return q.gen.GetColumnId(ctx, &gen.GetColumnIdParams{
		TableID:    tableId,
		ColumnName: strings.ToLower(columnName),
	})
}

func (q *queries) CreateAttribute(ctx context.Context, columnId int32, attributeType int32, value anytype.KwilAny) error {
	// marshal attribute value

	return q.gen.CreateAttribute(ctx, &gen.CreateAttributeParams{
		ColumnID:       columnId,
		AttributeType:  attributeType,
		AttributeValue: value.Bytes(),
	})
}

func (q *queries) CreateQuery(ctx context.Context, queryName string, dbId int32, queryData []byte) error {
	return q.gen.CreateQuery(ctx, &gen.CreateQueryParams{
		QueryName: strings.ToLower(queryName),
		DbID:      dbId,
		Query:     queryData,
	})
}

func (q *queries) CreateRole(ctx context.Context, dbId int32, roleName string, isDefault bool) error {
	return q.gen.CreateRole(ctx, &gen.CreateRoleParams{
		DbID:      dbId,
		RoleName:  strings.ToLower(roleName),
		IsDefault: isDefault,
	})
}

func (q *queries) ApplyPermissionToRole(ctx context.Context, dbId int32, roleName string, queryName string) error {
	return q.gen.ApplyPermissionToRole(ctx, &gen.ApplyPermissionToRoleParams{
		DbID:      dbId,
		RoleName:  strings.ToLower(roleName),
		QueryName: strings.ToLower(queryName),
	})
}

func (q *queries) CreateIndex(ctx context.Context, tableId int32, indexName string, indexType int32, columns []string) error {
	cols := make([]string, len(columns))
	for i, col := range columns {
		cols[i] = strings.ToLower(col)
	}

	return q.gen.CreateIndex(ctx, &gen.CreateIndexParams{
		TableID:   tableId,
		IndexName: strings.ToLower(indexName),
		IndexType: indexType,
		Columns:   cols,
	})
}
