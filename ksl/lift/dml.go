package lift

import (
	"ksl"
	"ksl/ast"
	"ksl/ast/pdb"
	"ksl/dml"
	"ksl/syntax/nodes"
	"ksl/ty"

	"github.com/samber/mo"
)

func Ast(schema *ast.SchemaAst) *dml.Datamodel {
	return newDmlCtx(schema).lift()
}

type astdmlctx struct {
	schema *ast.SchemaAst
}

func newDmlCtx(schema *ast.SchemaAst) *astdmlctx {
	return &astdmlctx{schema: schema}
}

func (l *astdmlctx) lift() *dml.Datamodel {
	dm := &dml.Datamodel{}

	for _, model := range l.schema.Db.WalkModels() {
		dm.Models = append(dm.Models, l.liftModel(model))
	}
	for _, enum := range l.schema.Db.WalkEnums() {
		dm.Enums = append(dm.Enums, l.liftEnum(enum))
	}

	l.liftRelations(dm)

	return dm
}

func (l *astdmlctx) liftRelations(dm *dml.Datamodel) {
	for _, relation := range l.schema.Db.WalkRelations() {
		switch relation := relation.Refine().(type) {
		case pdb.InlineRelationWalker:
			forwardModel := dm.FindModel(relation.ReferencingModel().Name()).MustGet()
			forwardField := relation.ForwardRelationField().MustGet()
			forwardModel.Fields = append(forwardModel.Fields, &dml.RelationField{
				Name:             forwardField.Name(),
				Arity:            liftArity(forwardField.Arity()),
				ReferentialArity: liftArity(forwardField.ReferentialArity()),
				Documentation:    forwardField.Documentation(),
				IsIgnored:        forwardField.IsIgnored(),
				RelationInfo: &dml.RelationInfo{
					Name:            relation.RelationName().String(),
					ReferencedModel: relation.ReferencedModel().Name(),
					References:      relation.ReferencedFieldNames(),
					Fields:          relation.ReferencingFieldNames(),
					OnDelete:        ty.MapOption(forwardField.OnDelete(), referentialAction),
					OnUpdate:        ty.MapOption(forwardField.OnUpdate(), referentialAction),
					ForeignKeyName:  relation.ForeignKeyName(),
				},
			})

			backModel := dm.FindModel(relation.ReferencedModel().Name()).MustGet()
			if backField, ok := relation.BackRelationField().Get(); ok {
				backModel.Fields = append(backModel.Fields, &dml.RelationField{
					Name:             backField.Name(),
					Arity:            liftArity(backField.Arity()),
					ReferentialArity: liftArity(backField.ReferentialArity()),
					Documentation:    backField.Documentation(),
					IsIgnored:        backField.IsIgnored(),
					RelationInfo: &dml.RelationInfo{
						Name:            relation.RelationName().String(),
						ReferencedModel: relation.ReferencingModel().Name(),
						OnDelete:        ty.MapOption(backField.OnDelete(), referentialAction),
						OnUpdate:        ty.MapOption(backField.OnUpdate(), referentialAction),
					},
				})
			} else {
				backModel.Fields = append(backModel.Fields, &dml.RelationField{
					Name:             relation.ReferencingModel().Name(),
					Arity:            dml.List,
					ReferentialArity: dml.List,
					IsIgnored:        relation.ReferencingModel().IsIgnored(),
					RelationInfo: &dml.RelationInfo{
						Name:            relation.RelationName().String(),
						ReferencedModel: relation.ReferencingModel().Name(),
					},
				})
			}

		case pdb.ImplicitManyToManyRelationWalker:
			for _, relationField := range []pdb.RelationFieldWalker{relation.FieldA(), relation.FieldB()} {
				model := dm.FindModel(relationField.Model().Name()).MustGet()
				model.Fields = append(model.Fields, &dml.RelationField{
					Name:             relationField.Name(),
					Arity:            liftArity(relationField.Arity()),
					ReferentialArity: liftArity(relationField.ReferentialArity()),
					Documentation:    relationField.Documentation(),
					IsIgnored:        relationField.IsIgnored(),
					RelationInfo: &dml.RelationInfo{
						Name:            relation.RelationName().String(),
						ReferencedModel: relationField.RelatedModel().Name(),
						References:      relationField.RelatedModel().MustPrimaryKey().FieldNames(),
						Fields:          relationField.FieldNames(),
						OnDelete:        ty.MapOption(relationField.OnDelete(), referentialAction),
						OnUpdate:        ty.MapOption(relationField.OnUpdate(), referentialAction),
					},
				})
			}
		}
	}
}

func (l *astdmlctx) liftEnum(enum pdb.EnumWalker) *dml.Enum {
	e := &dml.Enum{
		Name:          enum.DatabaseName(),
		Documentation: enum.Documentation(),
		DatabaseName:  enum.DatabaseName(),
	}

	for _, value := range enum.Values() {
		e.Values = append(e.Values, dml.EnumValue{
			Name:          value.Name(),
			DatabaseName:  value.DatabaseName(),
			Documentation: value.Documentation(),
		})
	}

	return nil
}

func (l *astdmlctx) liftModel(model pdb.ModelWalker) *dml.Model {
	astModel := model.AstModel()
	m := &dml.Model{
		Name:          astModel.GetName(),
		Documentation: astModel.Documentation(),
		DatabaseName:  model.DatabaseName(),
		IsIgnored:     model.IsIgnored(),
	}

	if pk, ok := model.PrimaryKey().Get(); ok {
		primaryKey := dml.PrimaryKey{
			Name:         pk.Name(),
			DatabaseName: pk.ConstraintName(),
		}
		for _, attr := range pk.ScalarFieldAttributes() {
			primaryKey.Fields = append(primaryKey.Fields, &dml.IndexField{
				Name:      attr.Field().Name(),
				SortOrder: sortOrder(attr.SortOrder()),
			})
		}
		m.PrimaryKey = mo.Some(primaryKey)
	}

	for _, index := range model.Indexes() {
		idx := &dml.Index{
			Name:         index.Name(),
			DatabaseName: index.ConstraintName(),
		}

		switch index.Type() {
		case pdb.IndexTypeUnique:
			idx.Type = dml.Unique
		case pdb.IndexTypeNormal:
			idx.Type = dml.Normal
		}

		switch index.Algorithm() {
		case pdb.BTree:
			idx.Algorithm = mo.Some(dml.BTree)
		case pdb.Hash:
			idx.Algorithm = mo.Some(dml.Hash)
		case pdb.Gist:
			idx.Algorithm = mo.Some(dml.Gist)
		case pdb.SpGist:
			idx.Algorithm = mo.Some(dml.SpGist)
		case pdb.Gin:
			idx.Algorithm = mo.Some(dml.Gin)
		case pdb.Brin:
			idx.Algorithm = mo.Some(dml.Brin)
		}

		for _, f := range index.ScalarFieldAttributes() {
			idx.Fields = append(idx.Fields, &dml.IndexField{
				Name:      f.Field().Name(),
				SortOrder: sortOrder(f.SortOrder()),
			})
		}
		m.Indexes = append(m.Indexes, idx)
	}

	for _, field := range model.ScalarFields() {
		astField := field.AstField()
		f := &dml.ScalarField{
			Name:          astField.GetName(),
			Arity:         liftArity(astField.Type.Arity),
			Documentation: astField.Documentation(),
			DatabaseName:  field.DatabaseName(),
			IsIgnored:     field.IsIgnored(),
		}

		switch typ := field.ScalarFieldType().Type.(type) {
		case pdb.BuiltInScalarType:
			var nativeType mo.Option[ksl.Type]
			if nt, ok := field.NativeType().Get(); ok {
				if t, err := l.schema.Backend.ParseNativeType(nt.Name, nt.Args...); err == nil {
					nativeType = mo.Some(t)
				}
			}
			f.Type = dml.ScalarFieldType{Type: typ.Deref(), NativeType: nativeType}
		case pdb.EnumFieldType:
			f.Type = dml.EnumFieldType{Enum: l.schema.Db.WalkEnum(typ.Enum).Name()}
		}
		m.Fields = append(m.Fields, f)
	}

	return m
}

func sortOrder(s string) mo.Option[dml.SortOrder] {
	switch s {
	case pdb.Ascending:
		return mo.Some(dml.Ascending)
	case pdb.Descending:
		return mo.Some(dml.Descending)
	}
	return mo.None[dml.SortOrder]()
}

func liftArity(arity nodes.FieldArity) dml.FieldArity {
	switch arity {
	case nodes.Repeated:
		return dml.List
	case nodes.Required:
		return dml.Required
	default:
		return dml.Optional
	}
}

func referentialAction(action pdb.ReferentialAction) mo.Option[dml.ReferentialAction] {
	switch action {
	case pdb.NoAction:
		return mo.Some(dml.NoAction)
	case pdb.Restrict:
		return mo.Some(dml.Restrict)
	case pdb.Cascade:
		return mo.Some(dml.Cascade)
	case pdb.SetNull:
		return mo.Some(dml.SetNull)
	case pdb.SetDefault:
		return mo.Some(dml.SetDefault)
	}
	return mo.None[dml.ReferentialAction]()
}
