package dml

import "github.com/samber/mo"

type Model struct {
	Name          string
	Fields        []Field
	Documentation string
	DatabaseName  string
	Indexes       []*Index
	PrimaryKey    mo.Option[PrimaryKey]
	IsGenerated   bool
	IsIgnored     bool
}
