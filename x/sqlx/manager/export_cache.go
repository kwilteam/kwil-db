package manager

func (c *CachedDB) ExportTables() []*ExportedTable {
	var tables []*ExportedTable
	for tableName := range c.Tables {
		tables = append(tables, &ExportedTable{
			Name:    tableName,
			Columns: c.ExportColumns(tableName),
		})
	}
	return tables
}

func (c *CachedDB) ExportColumns(tableName string) []*ExportedColumn {
	var columns []*ExportedColumn
	for columnName, column := range c.Tables[tableName].Columns {
		var attributes []*ExportedAttribute
		// get all the attributes
		for attributeName, value := range column.Attributes {
			attributes = append(attributes, &ExportedAttribute{
				Name:  attributeName.String(),
				Value: value,
			})
		}
		columns = append(columns, &ExportedColumn{
			Name:       columnName,
			Type:       column.Type.String(),
			Attributes: attributes,
		})
	}
	return columns
}

func (c *CachedDB) ExportQueries() []*ExportedQuery {
	var queries []*ExportedQuery
	for queryName, query := range c.Queries {
		var inputs []*ExportedInput
		var defaults []*ExportedDefault

		for i, arg := range query.Args {
			if arg.Fillable {
				inputs = append(inputs, &ExportedInput{
					Name:    arg.Column,
					Type:    arg.Type,
					Ordinal: i,
				})
			} else {
				defaults = append(defaults, &ExportedDefault{
					Name:    arg.Column,
					Type:    arg.Type,
					Value:   arg.Default,
					Ordinal: i,
				})
			}
		}

		queries = append(queries, &ExportedQuery{
			Name:      queryName,
			Statement: c.Queries[queryName].Statement,
			Inputs:    inputs,
			Defaults:  defaults,
		})
	}
	return queries
}

func (c *CachedDB) ExportRoles() []*ExportedRole {
	var roles []*ExportedRole

	for roleName, permissionedQueries := range c.Roles {
		var queries []string
		for queryName, allowed := range permissionedQueries {
			if allowed {
				queries = append(queries, queryName)
			}
		}
		roles = append(roles, &ExportedRole{
			Name:    roleName,
			Queries: queries,
		})
	}
	return roles
}

func (c *CachedDB) ExportIndexes() []*ExportedIndex {
	var indexes []*ExportedIndex

	for indexName, index := range c.Indexes {
		indexes = append(indexes, &ExportedIndex{
			Name:   indexName,
			Table:  index.Table,
			Column: index.Column,
			Using:  index.Using.String(),
		})
	}
	return indexes
}

func (c *CachedDB) Export(name string) (*ExportedDB, error) {

	db := ExportedDB{
		Name:        name,
		Owner:       c.Owner,
		DefaultRole: c.DefaultRole,
		Tables:      c.ExportTables(),
		Queries:     c.ExportQueries(),
		Roles:       c.ExportRoles(),
		Indexes:     c.ExportIndexes(),
	}

	return &db, nil
}
