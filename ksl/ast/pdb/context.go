package pdb

import (
	"ksl"
	"ksl/spec"
)

type context struct {
	Ast  *ParseTree
	Spec spec.SchemaSpec

	Config      *Config
	Context     *ksl.Context
	Names       *NamesContext
	Types       *TypesContext
	Relations   *RelationsContext
	Diagnostics ksl.Diagnostics
}

func newContext(s *ParseTree, spec spec.SchemaSpec, ctx *ksl.Context, diags ksl.Diagnostics) *context {
	return &context{
		Ast: s,
		Names: &NamesContext{
			ModelEnums:  map[string]NodeID{},
			Blocks:      map[string]map[string]NodeID{},
			ModelFields: map[string]map[string]NodeID{},
			EnumFields:  map[string]map[string]NodeID{},
		},
		Types: &TypesContext{
			ScalarFields:     map[ModelFieldID]*ScalarField{},
			RelationFields:   map[ModelFieldID]*RelationField{},
			EnumAnnotations:  map[EnumID]EnumAnnotations{},
			ModelAnnotations: map[ModelID]ModelAnnotations{},
		},
		Relations: &RelationsContext{
			Forward:  make(map[Rel]struct{}),
			Backward: make(map[Rel]struct{}),
		},
		Context:     ctx,
		Spec:        spec,
		Diagnostics: diags,
		Config:      &Config{},
	}
}

func (c *context) diag(diags ...*ksl.Diagnostic) bool {
	c.Diagnostics = append(c.Diagnostics, diags...)
	return ksl.Diagnostics(diags).HasErrors()
}
