package sqlschema

import (
	"ksl"
	"ksl/backend"
	"ksl/pdb"
	"ksl/schema"
	"ksl/syntax/ast"

	"github.com/samber/lo"
)

type calculate struct {
	schema   *schema.KwilSchema
	db       Database
	provider backend.Connector
}

func CalculateSqlSchema(s *schema.KwilSchema, dbName string) Database {
	ctx := &calculate{
		schema:   s,
		db:       NewDatabase(dbName),
		provider: s.Backend,
	}

	ctx.loadEnums()
	ctx.loadModelTables()
	ctx.loadRelations()

	return ctx.db
}
func (ctx *calculate) loadEnums() {
	for _, enum := range ctx.schema.Db.WalkEnums() {
		eid := ctx.db.AddEnum(enum.DatabaseName())
		for _, value := range enum.Values() {
			ctx.db.AddEnumVariant(eid, value.DatabaseName())
		}
	}
}

func (ctx *calculate) loadModelTables() {
	for _, model := range ctx.schema.Db.WalkModels() {
		tid := ctx.db.AddTable(Table{
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

func (ctx *calculate) loadModelIndexes(model pdb.ModelWalker, tableID TableID) {
	if pkw, ok := model.PrimaryKey(); ok {
		constraintName := pkw.ConstraintName()
		pkid := ctx.db.AddPrimaryKey(Index{Table: tableID, Name: constraintName})

		for _, attr := range pkw.ScalarFieldAttributes() {
			colid := ctx.db.WalkTable(tableID).Column(attr.Field().DatabaseName()).MustGet().ID
			ctx.db.AddIndexColumn(IndexColumn{Index: pkid, Column: colid, SortOrder: sortOrder(attr.SortOrder())})
		}
	}

	for _, index := range model.Indexes() {
		constraintName := index.ConstraintName()

		indexType := NormalIndex
		if index.IsUnique() {
			indexType = UniqueIndex
		}

		idxid := ctx.db.AddIndex(Index{
			Table:     tableID,
			Name:      constraintName,
			Type:      indexType,
			Algorithm: indexAlgorithm(index.Algorithm()),
		})

		for _, attr := range index.ScalarFieldAttributes() {
			colid := ctx.db.WalkTable(tableID).Column(attr.Field().DatabaseName()).MustGet().ID
			ctx.db.AddIndexColumn(IndexColumn{
				Index:     idxid,
				Column:    colid,
				SortOrder: sortOrder(attr.SortOrder()),
			})
		}
	}
}

func (ctx *calculate) loadEnumFieldColumn(field pdb.ScalarFieldWalker, enum pdb.EnumWalker, tableID TableID) {
	name := enum.DatabaseName()
	typ := EnumType{Name: name, ID: ctx.db.FindEnum(name).MustGet().ID}

	ctx.db.AddColumn(Column{
		Table: tableID,
		Name:  field.DatabaseName(),
		Type: ColumnType{
			Type:  typ,
			Raw:   typ.Name,
			Arity: columnArity(field.Arity()),
		},
		AutoIncrement: false,
		Default:       nil,
		Comment:       field.Documentation(),
	})
}

func (ctx *calculate) loadScalarFieldColumn(field pdb.ScalarFieldWalker, typ pdb.BuiltInScalarType, tableID TableID) {
	var scalarType ksl.Type
	var raw string

	if nativeType, ok := field.NativeType(); ok {
		raw = nativeType.Name
		scalarType = lo.Must(ctx.provider.ParseNativeType(nativeType.Name, nativeType.Args...))
	} else {
		raw = typ.Deref().Name()
		scalarType = ctx.provider.DefaultNativeTypeForScalar(typ.Deref())
	}
	ctx.db.AddColumn(Column{
		Table: tableID,
		Name:  field.DatabaseName(),
		Type: ColumnType{
			Type:  scalarType,
			Raw:   raw,
			Arity: columnArity(field.Arity()),
		},
		AutoIncrement: false,
		Default:       nil,
		Comment:       field.Documentation(),
	})
}

func (ctx *calculate) loadRelations() {
	for _, relation := range ctx.schema.Db.WalkRelations() {
		switch refined := relation.Refine().(type) {
		case pdb.InlineRelationWalker:
			relField := refined.ForwardRelationField().MustGet()
			referencingTable := ctx.db.FindTable(relField.Model().DatabaseName()).MustGet()
			referencedTable := ctx.db.FindTable(refined.ReferencedModel().DatabaseName()).MustGet()

			var onDelete, onUpdate ForeignKeyAction

			if oda, ok := relField.OnDelete(); ok {
				onDelete = referentialAction(oda.Action)
			} else {
				switch relField.ReferentialArity() {
				case ast.Required:
					onDelete = Restrict
				default:
					onDelete = SetNull
				}
			}

			if oua, ok := relField.OnUpdate(); ok {
				onUpdate = referentialAction(oua.Action)
			} else {
				onUpdate = Cascade
			}

			fkid := ctx.db.AddForeignKey(ForeignKey{
				ConstraintName:   refined.ConstraintName(),
				ConstrainedTable: referencingTable.ID,
				ReferencedTable:  referencedTable.ID,
				OnDeleteAction:   onDelete,
				OnUpdateAction:   onUpdate,
			})

			for _, col := range lo.Zip2(relField.Fields(), relField.ReferencedFields()) {
				ctx.db.AddForeignKeyColumn(ForeignKeyColumn{
					ForeignKey:        fkid,
					ConstrainedColumn: referencingTable.Column(col.A.DatabaseName()).MustGet().ID,
					ReferencedColumn:  referencedTable.Column(col.B.DatabaseName()).MustGet().ID,
				})
			}

		case pdb.ImplicitManyToManyRelationWalker:
			m2mTableName := "_" + refined.RelationName().String()
			m2mTable := ctx.db.AddTable(Table{Name: m2mTableName})

			modelA, modelB := refined.ModelA(), refined.ModelB()
			tableA := ctx.db.FindTable(modelA.DatabaseName()).MustGet()
			tableB := ctx.db.FindTable(modelB.DatabaseName()).MustGet()
			modelApk, modelBpk := modelA.MustPrimaryKey(), modelB.MustPrimaryKey()

			columnAtype := tableA.PrimaryKey().MustGet().Columns()[0].Column().Get().Type
			columnBtype := tableB.PrimaryKey().MustGet().Columns()[0].Column().Get().Type
			columnA := ctx.db.AddColumn(Column{Table: m2mTable, Name: "A", Type: columnAtype})
			columnB := ctx.db.AddColumn(Column{Table: m2mTable, Name: "B", Type: columnBtype})

			// Unique index on AB
			abIndex := ctx.db.AddUniqueIndex(Index{Table: m2mTable, Name: m2mTableName + "_AB_unique"})
			ctx.db.AddIndexColumn(IndexColumn{Index: abIndex, Column: columnA})
			ctx.db.AddIndexColumn(IndexColumn{Index: abIndex, Column: columnB})

			// Index on B
			bIndex := ctx.db.AddIndex(Index{Table: m2mTable, Name: m2mTableName + "_B_index"})
			ctx.db.AddIndexColumn(IndexColumn{Index: bIndex, Column: columnB})

			fkid := ctx.db.AddForeignKey(ForeignKey{
				ConstraintName:   m2mTableName + "_A_fkey",
				ConstrainedTable: m2mTable,
				ReferencedTable:  tableA.ID,
				OnDeleteAction:   Cascade,
				OnUpdateAction:   Cascade,
			})

			ctx.db.AddForeignKeyColumn(ForeignKeyColumn{
				ForeignKey:        fkid,
				ConstrainedColumn: columnA,
				ReferencedColumn:  tableA.Column(modelApk.FirstField().DatabaseName()).MustGet().ID,
			})

			fkid = ctx.db.AddForeignKey(ForeignKey{
				ConstraintName:   m2mTableName + "_B_fkey",
				ConstrainedTable: m2mTable,
				ReferencedTable:  tableB.ID,
				OnDeleteAction:   Cascade,
				OnUpdateAction:   Cascade,
			})

			ctx.db.AddForeignKeyColumn(ForeignKeyColumn{
				ForeignKey:        fkid,
				ConstrainedColumn: columnB,
				ReferencedColumn:  tableB.Column(modelBpk.FirstField().DatabaseName()).MustGet().ID,
			})
		}
	}
}

func columnArity(arity ast.FieldArity) ColumnArity {
	switch arity {
	case ast.Optional:
		return Nullable
	case ast.Repeated:
		return List
	case ast.Required:
		return Required
	default:
		panic("unreachable")
	}
}

func referentialAction(action pdb.ReferentialAction) ForeignKeyAction {
	switch action {
	case pdb.NoAction:
		return NoAction
	case pdb.Cascade:
		return Cascade
	case pdb.Restrict:
		return Restrict
	case pdb.SetDefault:
		return SetDefault
	case pdb.SetNull:
		return SetNull
	default:
		panic("unreachable")
	}
}

func sortOrder(order string) SortOrder {
	switch order {
	case pdb.Ascending:
		return Ascending
	case pdb.Descending:
		return Descending
	default:
		return Ascending
	}
}

func indexAlgorithm(algo pdb.IndexAlgorithm) IndexAlgorithm {
	switch algo {
	case pdb.BTree:
		return BTreeAlgo
	case pdb.Hash:
		return HashAlgo
	case pdb.Gist:
		return GistAlgo
	case pdb.Gin:
		return GinAlgo
	case pdb.SpGist:
		return SpGistAlgo
	case pdb.Brin:
		return BrinAlgo
	}
	return BTreeAlgo
}
