package dml

import "github.com/samber/mo"

type PrimaryKey struct {
	Name         string
	DatabaseName string
	Fields       []*IndexField
}

type Index struct {
	Name         string
	DatabaseName string
	Fields       []*IndexField
	Type         IndexType
	Algorithm    mo.Option[IndexAlgorithm]
}

type IndexField struct {
	Name      string
	SortOrder mo.Option[SortOrder]
}
