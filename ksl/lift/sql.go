package lift

import (
	"ksl"
	"ksl/ast"
	"ksl/ast/pdb"
	"ksl/sqlschema"
	"ksl/syntax/nodes"

	"github.com/samber/lo"
)

func Sql(s *ast.SchemaAst, dbName string) sqlschema.Database {
	ctx := newSqlCtx(s, dbName)

	ctx.loadEnums()
	ctx.loadModelTables()
	ctx.loadRelations()

	return ctx.db
}

type sqlctx struct {
	schema *ast.SchemaAst
	db     sqlschema.Database
}

func newSqlCtx(s *ast.SchemaAst, dbName string) *sqlctx {
	return &sqlctx{
		schema: s,
		db:     sqlschema.NewDatabase(dbName),
	}
}

func (ctx *sqlctx) loadEnums() {
	for _, enum := range ctx.schema.Db.WalkEnums() {
		eid := ctx.db.AddEnum(enum.DatabaseName())
		for _, value := range enum.Values() {
			ctx.db.AddEnumVariant(eid, value.DatabaseName())
		}
	}
}

func (ctx *sqlctx) loadModelTables() {
	for _, model := range ctx.schema.Db.WalkModels() {
		tid := ctx.db.AddTable(sqlschema.Table{
			Name:    model.DatabaseName(),
			Comment: model.Documentation(),
		})

		for _, field := range model.ScalarFields() {
			fieldType := field.ScalarFieldType()
			switch typ := fieldType.Type.(type) {
			case pdb.EnumFieldType:
				ctx.loadEnumFieldColumn(field, ctx.schema.Db.WalkEnum(typ.Enum), tid)
			case pdb.BuiltInScalarType:
				ctx.loadScalarFieldColumn(field, typ, tid)
			}
		}

		ctx.loadModelIndexes(model, tid)
	}
}

func (ctx *sqlctx) loadModelIndexes(model pdb.ModelWalker, tableID sqlschema.TableID) {
	if pkw, ok := model.PrimaryKey().Get(); ok {
		constraintName := pkw.ConstraintName()
		pkid := ctx.db.AddPrimaryKey(sqlschema.Index{Table: tableID, Name: constraintName})

		for _, attr := range pkw.ScalarFieldAttributes() {
			colid := ctx.db.WalkTable(tableID).Column(attr.Field().DatabaseName()).MustGet().ID
			ctx.db.AddIndexColumn(sqlschema.IndexColumn{Index: pkid, Column: colid, SortOrder: ParseSortOrderToSqlSchema(attr.SortOrder())})
		}
	}

	for _, index := range model.Indexes() {
		constraintName := index.ConstraintName()

		indexType := sqlschema.NormalIndex
		if index.IsUnique() {
			indexType = sqlschema.UniqueIndex
		}

		idxid := ctx.db.AddIndex(sqlschema.Index{
			Table:     tableID,
			Name:      constraintName,
			Type:      indexType,
			Algorithm: ParseIndexAlgorithmToSqlSchema(index.Algorithm()),
		})

		for _, attr := range index.ScalarFieldAttributes() {
			colid := ctx.db.WalkTable(tableID).Column(attr.Field().DatabaseName()).MustGet().ID
			ctx.db.AddIndexColumn(sqlschema.IndexColumn{
				Index:     idxid,
				Column:    colid,
				SortOrder: ParseSortOrderToSqlSchema(attr.SortOrder()),
			})
		}
	}
}

func (ctx *sqlctx) loadEnumFieldColumn(field pdb.ScalarFieldWalker, enum pdb.EnumWalker, tableID sqlschema.TableID) {
	name := enum.DatabaseName()
	typ := sqlschema.EnumType{Name: name, ID: ctx.db.FindEnum(name).MustGet().ID}

	ctx.db.AddColumn(sqlschema.Column{
		Table: tableID,
		Name:  field.DatabaseName(),
		Type: sqlschema.ColumnType{
			Type:  typ,
			Raw:   typ.Name,
			Arity: AstFieldArityToSqlSchema(field.Arity()),
		},
		AutoIncrement: false,
		Default:       nil,
		Comment:       field.Documentation(),
	})
}

func (ctx *sqlctx) loadScalarFieldColumn(field pdb.ScalarFieldWalker, typ pdb.BuiltInScalarType, tableID sqlschema.TableID) {
	var scalarType ksl.Type
	var raw string

	if nativeType, ok := field.NativeType().Get(); ok {
		raw = nativeType.Name
		scalarType = lo.Must(ctx.schema.Backend.ParseNativeType(nativeType.Name, nativeType.Args...))
	} else {
		raw = typ.Deref().Name()
		scalarType = ctx.schema.Backend.DefaultNativeTypeForScalar(typ.Deref())
	}
	ctx.db.AddColumn(sqlschema.Column{
		Table: tableID,
		Name:  field.DatabaseName(),
		Type: sqlschema.ColumnType{
			Type:  scalarType,
			Raw:   raw,
			Arity: AstFieldArityToSqlSchema(field.Arity()),
		},
		AutoIncrement: false,
		Default:       nil,
		Comment:       field.Documentation(),
	})
}

func (ctx *sqlctx) loadRelations() {
	for _, relation := range ctx.schema.Db.WalkRelations() {
		switch refined := relation.Refine().(type) {
		case pdb.InlineRelationWalker:
			relField := refined.ForwardRelationField().MustGet()
			referencingTable := ctx.db.FindTable(relField.Model().DatabaseName()).MustGet()
			referencedTable := ctx.db.FindTable(refined.ReferencedModel().DatabaseName()).MustGet()

			var onDelete, onUpdate sqlschema.ForeignKeyAction

			if oda, ok := relField.OnDelete().Get(); ok {
				onDelete = ParseReferentialActionToSqlSchema(oda)
			} else {
				switch relField.ReferentialArity() {
				case nodes.Required:
					onDelete = sqlschema.Restrict
				default:
					onDelete = sqlschema.SetNull
				}
			}

			if oua, ok := relField.OnUpdate().Get(); ok {
				onUpdate = ParseReferentialActionToSqlSchema(oua)
			} else {
				onUpdate = sqlschema.Cascade
			}

			fkid := ctx.db.AddForeignKey(sqlschema.ForeignKey{
				ConstraintName:   refined.ConstraintName(),
				ConstrainedTable: referencingTable.ID,
				ReferencedTable:  referencedTable.ID,
				OnDeleteAction:   onDelete,
				OnUpdateAction:   onUpdate,
			})

			for _, col := range lo.Zip2(relField.Fields(), relField.ReferencedFields()) {
				ctx.db.AddForeignKeyColumn(sqlschema.ForeignKeyColumn{
					ForeignKey:        fkid,
					ConstrainedColumn: referencingTable.Column(col.A.DatabaseName()).MustGet().ID,
					ReferencedColumn:  referencedTable.Column(col.B.DatabaseName()).MustGet().ID,
				})
			}

		case pdb.ImplicitManyToManyRelationWalker:
			m2mTableName := "_" + refined.RelationName().String()
			m2mTable := ctx.db.AddTable(sqlschema.Table{Name: m2mTableName})

			modelA, modelB := refined.ModelA(), refined.ModelB()
			tableA := ctx.db.FindTable(modelA.DatabaseName()).MustGet()
			tableB := ctx.db.FindTable(modelB.DatabaseName()).MustGet()
			modelApk, modelBpk := modelA.MustPrimaryKey(), modelB.MustPrimaryKey()

			columnAtype := tableA.PrimaryKey().MustGet().Columns()[0].Column().Get().Type
			columnBtype := tableB.PrimaryKey().MustGet().Columns()[0].Column().Get().Type
			columnA := ctx.db.AddColumn(sqlschema.Column{Table: m2mTable, Name: "A", Type: columnAtype})
			columnB := ctx.db.AddColumn(sqlschema.Column{Table: m2mTable, Name: "B", Type: columnBtype})

			// Unique index on AB
			abIndex := ctx.db.AddUniqueIndex(sqlschema.Index{Table: m2mTable, Name: m2mTableName + "_AB_unique"})
			ctx.db.AddIndexColumn(sqlschema.IndexColumn{Index: abIndex, Column: columnA})
			ctx.db.AddIndexColumn(sqlschema.IndexColumn{Index: abIndex, Column: columnB})

			// sqlschema.Index on B
			bIndex := ctx.db.AddIndex(sqlschema.Index{Table: m2mTable, Name: m2mTableName + "_B_index"})
			ctx.db.AddIndexColumn(sqlschema.IndexColumn{Index: bIndex, Column: columnB})

			fkid := ctx.db.AddForeignKey(sqlschema.ForeignKey{
				ConstraintName:   m2mTableName + "_A_fkey",
				ConstrainedTable: m2mTable,
				ReferencedTable:  tableA.ID,
				OnDeleteAction:   sqlschema.Cascade,
				OnUpdateAction:   sqlschema.Cascade,
			})

			ctx.db.AddForeignKeyColumn(sqlschema.ForeignKeyColumn{
				ForeignKey:        fkid,
				ConstrainedColumn: columnA,
				ReferencedColumn:  tableA.Column(modelApk.FirstField().DatabaseName()).MustGet().ID,
			})

			fkid = ctx.db.AddForeignKey(sqlschema.ForeignKey{
				ConstraintName:   m2mTableName + "_B_fkey",
				ConstrainedTable: m2mTable,
				ReferencedTable:  tableB.ID,
				OnDeleteAction:   sqlschema.Cascade,
				OnUpdateAction:   sqlschema.Cascade,
			})

			ctx.db.AddForeignKeyColumn(sqlschema.ForeignKeyColumn{
				ForeignKey:        fkid,
				ConstrainedColumn: columnB,
				ReferencedColumn:  tableB.Column(modelBpk.FirstField().DatabaseName()).MustGet().ID,
			})
		}
	}
}

func AstFieldArityToSqlSchema(arity nodes.FieldArity) sqlschema.ColumnArity {
	switch arity {
	case nodes.Optional:
		return sqlschema.Nullable
	case nodes.Repeated:
		return sqlschema.List
	case nodes.Required:
		return sqlschema.Required
	default:
		panic("unreachable")
	}
}

func ParseReferentialActionToSqlSchema(action pdb.ReferentialAction) sqlschema.ForeignKeyAction {
	switch action {
	case pdb.NoAction:
		return sqlschema.NoAction
	case pdb.Cascade:
		return sqlschema.Cascade
	case pdb.Restrict:
		return sqlschema.Restrict
	case pdb.SetDefault:
		return sqlschema.SetDefault
	case pdb.SetNull:
		return sqlschema.SetNull
	default:
		panic("unreachable")
	}
}

func ParseSortOrderToSqlSchema(order string) sqlschema.SortOrder {
	switch order {
	case pdb.Ascending:
		return sqlschema.Ascending
	case pdb.Descending:
		return sqlschema.Descending
	default:
		return sqlschema.Ascending
	}
}

func ParseIndexAlgorithmToSqlSchema(algo pdb.IndexAlgorithm) sqlschema.IndexAlgorithm {
	switch algo {
	case pdb.BTree:
		return sqlschema.BTreeAlgo
	case pdb.Hash:
		return sqlschema.HashAlgo
	case pdb.Gist:
		return sqlschema.GistAlgo
	case pdb.Gin:
		return sqlschema.GinAlgo
	case pdb.SpGist:
		return sqlschema.SpGistAlgo
	case pdb.Brin:
		return sqlschema.BrinAlgo
	}
	return sqlschema.BTreeAlgo
}
