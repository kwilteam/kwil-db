/*
Package datasource provides the data source for the cost engine.
There are two types of data sources: SchemaSource and DataSource.

SchemaSource is an interface that provides access to schema info. It's supposed
to be used in logical plan, since it doesn't have the ability to scan the
underlying data.
*/
package datasource
