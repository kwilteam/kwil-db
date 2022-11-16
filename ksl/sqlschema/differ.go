package sqlschema

import (
	"sort"

	"golang.org/x/exp/slices"
)

type Differ interface {
	Diff(from, to Database) ([]MigrationStep, error)
}

type MigrationStep interface{ step() }

type differ struct {
	flavor SqlDiffFlavor
}

func NewDiffer(flavor SqlDiffFlavor) Differ { return &differ{flavor} }

func (d *differ) Diff(from, to Database) ([]MigrationStep, error) {
	ctx := newDiffContext(from, to, d.flavor)
	if err := ctx.calculateExtensionSteps(); err != nil {
		return nil, err
	}
	if err := ctx.calculateCreatedTableSteps(); err != nil {
		return nil, err
	}
	if err := ctx.calculateDroppedTableSteps(); err != nil {
		return nil, err
	}
	if err := ctx.calculateDroppedIndexSteps(); err != nil {
		return nil, err
	}
	if err := ctx.calculateCreatedIndexSteps(); err != nil {
		return nil, err
	}
	if err := ctx.calculateAlteredTableSteps(); err != nil {
		return nil, err
	}
	if err := ctx.calculateEnumSteps(); err != nil {
		return nil, err
	}
	sort.Stable(byStepType{ctx.Steps})
	return ctx.Steps, nil
}

type diffctx struct {
	Schemas Pair[Database]
	Db      *diffdb
	Steps   []MigrationStep
}

func newDiffContext(prev, next Database, flavor SqlDiffFlavor) *diffctx {
	return &diffctx{
		Db:      newDiffDb(prev, next, flavor),
		Schemas: MakePair(prev, next),
	}
}

func (ctx *diffctx) step(step MigrationStep) {
	ctx.Steps = append(ctx.Steps, step)
}

func (ctx *diffctx) calculateCreatedTableSteps() error {
	for _, table := range ctx.Db.CreatedTables() {
		ctx.step(CreateTable{Table: table.ID})

		for _, fk := range table.ForeignKeys() {
			ctx.step(AddForeignKey{ForeignKey: fk.ID})
		}

		for _, index := range table.Indexes() {
			if !index.IsPrimaryKey() {
				ctx.step(CreateIndex{Index: index.ID})
			}
		}
	}
	return nil
}

func (ctx *diffctx) calculateDroppedTableSteps() error {
	for _, table := range ctx.Db.DroppedTables() {
		ctx.step(DropTable{Table: table.ID})

		for _, fk := range table.ForeignKeys() {
			ctx.step(DropForeignKey{ForeignKey: fk.ID})
		}
	}
	return nil
}

func (ctx *diffctx) calculateDroppedIndexSteps() error {
	for _, table := range ctx.Db.TablePairs() {
		for _, index := range table.DroppedIndexes() {
			ctx.step(DropIndex{Index: index.ID})
		}
	}
	return nil
}

func (ctx *diffctx) calculateCreatedIndexSteps() error {
	for _, table := range ctx.Db.TablePairs() {
		for _, index := range table.CreatedIndexes() {
			ctx.step(CreateIndex{Index: index.ID})
		}

		var darcols []ColumnID
		for _, col := range table.ColumnPairs() {
			changes := ctx.Db.ColumnChanges[MakePair(col.Prev.ID, col.Next.ID)]
			if changes.TypeChange != ColumnTypeChangeNotCastable {
				continue
			}
			darcols = append(darcols, col.Next.ID)
		}

		for _, index := range table.IndexPairs() {
			for _, col := range index.Next.Columns() {
				if !slices.Contains(darcols, col.Column().ID) {
					continue
				}
				ctx.step(CreateIndex{Index: index.Next.ID})
			}
		}
	}
	return nil
}

func (ctx *diffctx) calculateAlteredTableSteps() error {
	for _, table := range ctx.Db.TablePairs() {
		for _, fk := range table.CreatedForeignKeys() {
			ctx.step(AddForeignKey{ForeignKey: fk.ID})
		}

		for _, fk := range table.DroppedForeignKeys() {
			ctx.step(DropForeignKey{ForeignKey: fk.ID})
		}

		for _, fk := range table.ForeignKeyPairs() {
			if fk.Prev.ConstraintName() != fk.Next.ConstraintName() {
				if fk.Prev.IsImplicitManyToManyFK() && fk.Next.IsImplicitManyToManyFK() {
					continue
				}
				ctx.step(RenameForeignKey{ForeignKeys: MakePair(fk.Prev.ID, fk.Next.ID)})
			}
		}

		for _, index := range table.IndexPairs() {
			if index.Prev.Name() != index.Next.Name() {
				ctx.step(RenameIndex{Index: MakePair(index.Prev.ID, index.Next.ID)})
			}
		}

		var changes []TableChange

		if table.DroppedPrimaryKey().IsPresent() || table.PrimaryKeyChanged() {
			changes = append(changes, DropPrimaryKey{})
		}

		if table.RenamedPrimaryKey() {
			changes = append(changes, RenamePrimaryKey{})
		}

		for _, col := range table.DroppedColumns() {
			changes = append(changes, DropColumn{Column: col.ID})
		}

		for _, col := range table.AddedColumns() {
			changes = append(changes, AddColumn{Column: col.ID})
		}

		var alterColumns []TableChange

		for _, col := range table.ColumnPairs() {
			columnChanges := ctx.Db.ColumnChanges[MakePair(col.Prev.ID, col.Next.ID)]

			if !columnChanges.DiffersInSomething() {
				continue
			}

			colid := MakePair(col.Prev.ID, col.Next.ID)
			switch columnChanges.TypeChange {
			case ColumnTypeChangeNotCastable:
				alterColumns = append(alterColumns, DropAndRecreateColumn{Columns: colid, Changes: columnChanges})
			case ColumnTypeChangeRiskyCast:
				alterColumns = append(alterColumns, AlterColumn{Columns: colid, Changes: columnChanges, TypeChange: columnChanges.TypeChange})
			case ColumnTypeChangeSafeCast:
				alterColumns = append(alterColumns, AlterColumn{Columns: colid, Changes: columnChanges, TypeChange: columnChanges.TypeChange})
			default:
				alterColumns = append(alterColumns, AlterColumn{Columns: colid, Changes: columnChanges})
			}
		}

		// TODO: sort alter columns

		changes = append(changes, alterColumns...)

		if table.CreatedPrimaryKey().IsPresent() || table.PrimaryKeyChanged() {
			changes = append(changes, AddPrimaryKey{})
		}

		if len(changes) == 0 {
			continue
		}

		ctx.step(AlterTable{Tables: MakePair(table.Previous().ID, table.Next().ID), Changes: changes})
	}

	return nil
}

func (ctx *diffctx) calculateEnumSteps() error {
	for _, enum := range ctx.Db.EnumPairs() {
		created := enum.CreatedVariants()
		dropped := enum.DroppedVariants()
		if len(created) == 0 && len(dropped) == 0 {
			continue
		}

		ctx.step(AlterEnum{
			Enums:           enum.IDS(),
			CreatedVariants: created,
			DroppedVariants: dropped,
		})
	}

	for _, enum := range ctx.Db.CreatedEnums() {
		ctx.step(CreateEnum{Enum: enum.ID})
	}

	for _, enum := range ctx.Db.DroppedEnums() {
		ctx.step(DropEnum{Enum: enum.ID})
	}

	return nil
}

func (ctx *diffctx) calculateExtensionSteps() error { return nil }
