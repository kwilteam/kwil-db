package pdb

import (
	"ksl"
	"ksl/syntax/nodes"
)

type Db struct {
	Ast       *ParseTree
	Names     *NamesContext
	Types     *TypesContext
	Relations *RelationsContext
	Context   *ksl.Context
	Config    *Config
}

func New(fs []*nodes.File, ctx *ksl.Context, diags ksl.Diagnostics) (*Db, ksl.Diagnostics) {
	ptree := NewParseTree(fs...)
	c := newContext(ptree, Spec, ctx, diags)

	// First pass: resolve names.
	c.ResolveNames()

	// Return early on name resolution errors
	if c.Diagnostics.HasErrors() {
		return &Db{Ast: ptree, Context: ctx, Config: c.Config, Types: c.Types, Names: c.Names, Relations: c.Relations}, c.Diagnostics
	}

	// Second pass: resolve top-level items and field types.
	c.ResolveTypes()

	// Return early on type resolution errors
	if c.Diagnostics.HasErrors() {
		return &Db{Ast: ptree, Context: ctx, Config: c.Config, Types: c.Types, Names: c.Names, Relations: c.Relations}, c.Diagnostics
	}

	// Third pass: validate model and field annotations.
	c.ResolveAnnotations()

	// Fourth step: relation inference
	c.InferRelations()

	return &Db{Ast: ptree, Context: ctx, Config: c.Config, Types: c.Types, Names: c.Names, Relations: c.Relations}, c.Diagnostics
}

func (db *Db) Eval(expr nodes.Expression, target any) ksl.Diagnostics {
	return Eval(expr, db.Context, target)
}
