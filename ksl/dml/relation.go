package dml

import "github.com/samber/mo"

type RelationInfo struct {
	ReferencedModel string
	Fields          []string
	References      []string
	Name            string
	ForeignKeyName  mo.Option[string]
	OnDelete        mo.Option[ReferentialAction]
	OnUpdate        mo.Option[ReferentialAction]
}
