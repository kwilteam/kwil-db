package sqlschema

import (
	"fmt"

	"github.com/samber/lo"
	"github.com/samber/mo"
)

type diffdb struct {
	Databases     Pair[Database]
	Tables        map[string]Pair[mo.Option[TableID]]
	Columns       map[lo.Tuple2[string, Pair[TableID]]]Pair[mo.Option[ColumnID]]
	ColumnChanges map[Pair[ColumnID]]ColumnChanges
}

func newDiffDb(prev, next Database, flavor SqlDiffFlavor) *diffdb {
	databases := MakePair(prev, next)
	tables := make(map[string]Pair[mo.Option[TableID]])
	columns := make(map[lo.Tuple2[string, Pair[TableID]]]Pair[mo.Option[ColumnID]])
	columnChanges := make(map[Pair[ColumnID]]ColumnChanges)

	for _, table := range prev.WalkTables() {
		tables[table.Name()] = MakePair(mo.Some(table.ID), mo.None[TableID]())
	}

	for _, table := range next.WalkTables() {
		entry := tables[table.Name()]
		entry.Next = mo.Some(table.ID)
		tables[table.Name()] = entry

		// Deal with tables that are both in the previous and the next
		// schema: we are going to look at heir columns.
		if entry.Prev.IsPresent() {
			tpair := MakePair(entry.Prev.MustGet(), entry.Next.MustGet())
			pt, nt := prev.WalkTable(entry.Prev.MustGet()), next.WalkTable(table.ID)
			colcache := make(map[string]Pair[mo.Option[ColumnID]])

			// Same as for tables, walk the previous columns first.
			for _, column := range pt.Columns() {
				colcache[column.Name()] = MakePair(mo.Some(column.ID), mo.None[ColumnID]())
			}

			// Then walk the next columns.
			for _, column := range nt.Columns() {
				entry := colcache[column.Name()]
				entry.Next = mo.Some(column.ID)
				colcache[column.Name()] = entry
			}
			// If the column is both in the previous and the next
			// schema, we are going to look at the changes.
			for name, cids := range colcache {
				columns[lo.T2(name, tpair)] = cids

				if cids.Prev.IsPresent() && cids.Next.IsPresent() {
					pc, nc := prev.WalkColumn(cids.Prev.MustGet()), next.WalkColumn(cids.Next.MustGet())
					var changes ColumnChange
					typeChange := flavor.ColumnTypeChange(pc, nc)
					if pc.Arity() != nc.Arity() {
						changes |= ColumnChangeArity
					}
					if typeChange != ColumnTypeChangeNone {
						changes |= ColumnChangeType
					}
					if pc.Default() != nc.Default() {
						changes |= ColumnChangeDefault
					}
					if pc.Get().AutoIncrement != nc.Get().AutoIncrement {
						changes |= ColumnChangeAutoIncrement
					}

					columnChanges[MakePair(pc.ID, nc.ID)] = ColumnChanges{TypeChange: typeChange, Changes: changes}
				}
			}
		}
	}

	return &diffdb{
		Databases:     databases,
		Tables:        tables,
		Columns:       columns,
		ColumnChanges: columnChanges,
	}
}

func (db *diffdb) CreatedEnums() []EnumWalker {
	var created []EnumWalker
	prevEnums := make(map[string]EnumWalker)

	for _, prev := range db.Databases.Prev.WalkEnums() {
		prevEnums[prev.Name()] = prev
	}

	for _, next := range db.Databases.Next.WalkEnums() {
		if _, ok := prevEnums[next.Name()]; !ok {
			created = append(created, next)
		}
	}
	return created
}

func (db *diffdb) DroppedEnums() []EnumWalker {
	var dropped []EnumWalker
	nextEnums := make(map[string]EnumWalker)

	for _, next := range db.Databases.Next.WalkEnums() {
		nextEnums[next.Name()] = next
	}

	for _, prev := range db.Databases.Prev.WalkEnums() {
		if _, ok := nextEnums[prev.Name()]; !ok {
			dropped = append(dropped, prev)
		}
	}
	return dropped
}

func (db *diffdb) EnumPairs() []EnumDiffer {
	var pairs []EnumDiffer
	prevEnums := make(map[string]EnumWalker)

	for _, prev := range db.Databases.Prev.WalkEnums() {
		prevEnums[prev.Name()] = prev
	}

	for _, next := range db.Databases.Next.WalkEnums() {
		if prev, ok := prevEnums[next.Name()]; ok {
			pairs = append(pairs, EnumDiffer{enums: MakePair(prev, next), db: db})
		}
	}

	return pairs
}

func (db *diffdb) CreatedTables() []TableWalker {
	var tables []TableWalker
	for _, entry := range db.Tables {
		if entry.Prev.IsAbsent() && entry.Next.IsPresent() {
			tables = append(tables, db.Databases.Next.WalkTable(entry.Next.MustGet()))
		}
	}
	return tables
}

func (db *diffdb) DroppedTables() []TableWalker {
	var tables []TableWalker
	for _, entry := range db.Tables {
		if entry.Prev.IsPresent() && entry.Next.IsAbsent() {
			tables = append(tables, db.Databases.Next.WalkTable(entry.Prev.MustGet()))
		}
	}
	return tables
}

func (db *diffdb) DroppedColumns(table Pair[TableID]) []ColumnWalker {
	var columns []ColumnWalker
	for key, entry := range db.Columns {
		if key.B.Prev == table.Prev && key.B.Next == table.Next && entry.Prev.IsPresent() && entry.Next.IsAbsent() {
			columns = append(columns, db.Databases.Prev.WalkColumn(entry.Prev.MustGet()))
		}
	}
	return columns
}

func (db *diffdb) AddedColumns(table Pair[TableID]) []ColumnWalker {
	var columns []ColumnWalker
	for key, entry := range db.Columns {
		if key.B.Prev == table.Prev && key.B.Next == table.Next && entry.Prev.IsAbsent() && entry.Next.IsPresent() {
			columns = append(columns, db.Databases.Next.WalkColumn(entry.Next.MustGet()))
		}
	}
	return columns
}

func (db *diffdb) TablePairs() []TableDiffer {
	var tables []TableDiffer
	for _, table := range db.Tables {
		if table.Prev.IsPresent() && table.Next.IsPresent() {
			tables = append(tables, TableDiffer{
				tables: MakePair(
					db.Databases.Prev.WalkTable(table.Prev.MustGet()),
					db.Databases.Next.WalkTable(table.Next.MustGet()),
				),
				db: db,
			})
		}
	}
	return tables
}

func (db *diffdb) ColumnPairs(table Pair[TableID]) []Pair[ColumnWalker] {
	var columns []Pair[ColumnWalker]
	for _, col := range db.Columns {
		if col.Prev.IsPresent() && col.Next.IsPresent() {
			pc := db.Databases.Prev.WalkColumn(col.Prev.MustGet())
			nc := db.Databases.Next.WalkColumn(col.Next.MustGet())
			if pc.Table().ID == table.Prev && nc.Table().ID == table.Next {
				columns = append(columns, MakePair(pc, nc))
			}
		}
	}
	return columns
}

func (db *diffdb) foreignKeysMatch(a, b ForeignKeyWalker) bool {
	if a.ReferencedTable().Name() != b.ReferencedTable().Name() {
		return false
	}

	constrainedAcols, constrainedBcols := a.ConstrainedColumns(), b.ConstrainedColumns()
	referencedAcols, referencedBcols := a.ReferencedColumns(), b.ReferencedColumns()

	if len(constrainedAcols) != len(constrainedBcols) {
		return false
	}

	if len(referencedAcols) != len(referencedBcols) {
		return false
	}

	for _, cols := range lo.Zip2(constrainedAcols, constrainedBcols) {
		if cols.A.Name() != cols.B.Name() {
			return false
		}

		changes := db.ColumnChanges[MakePair(cols.A.ID, cols.B.ID)]
		if changes.TypeChanged() {
			return false
		}
	}

	for _, cols := range lo.Zip2(referencedAcols, referencedBcols) {
		if cols.A.Name() != cols.B.Name() {
			return false
		}
	}

	return a.OnDeleteAction() == b.OnDeleteAction() && a.OnUpdateAction() == b.OnUpdateAction()
}

func (db *diffdb) indexesMatch(a, b IndexWalker) bool {
	leftColumns := a.Columns()
	rightColumns := b.Columns()

	if len(leftColumns) != len(rightColumns) {
		return false
	}

	for _, cols := range lo.Zip2(leftColumns, rightColumns) {
		if cols.A.Name() != cols.B.Name() {
			return false
		}

		if cols.A.SortOrder() != cols.B.SortOrder() {
			return false
		}
	}

	return a.IndexType() == b.IndexType()
}

type EnumDiffer struct {
	enums Pair[EnumWalker]
	db    *diffdb
}

func (d EnumDiffer) IDS() Pair[EnumID] {
	return MakePair(d.enums.Prev.ID, d.enums.Next.ID)
}

func (d EnumDiffer) CreatedVariants() []string {
	prev := map[string]struct{}{}

	var variants []string

	for _, variant := range d.enums.Prev.Values() {
		prev[variant] = struct{}{}
	}

	for _, variant := range d.enums.Next.Values() {
		if _, ok := prev[variant]; !ok {
			variants = append(variants, variant)
		}
	}

	return variants
}

func (d EnumDiffer) DroppedVariants() []string {
	next := map[string]struct{}{}

	var variants []string

	for _, variant := range d.enums.Next.Values() {
		next[variant] = struct{}{}
	}

	for _, variant := range d.enums.Prev.Values() {
		if _, ok := next[variant]; !ok {
			variants = append(variants, variant)
		}
	}

	return variants
}

type TableDiffer struct {
	tables Pair[TableWalker]
	db     *diffdb
}

func (t TableDiffer) ColumnPairs() []Pair[ColumnWalker] {
	return t.db.ColumnPairs(MakePair(t.tables.Prev.ID, t.tables.Next.ID))
}

func (t TableDiffer) AddedColumns() []ColumnWalker {
	return t.db.AddedColumns(MakePair(t.tables.Prev.ID, t.tables.Next.ID))
}

func (t TableDiffer) DroppedColumns() []ColumnWalker {
	return t.db.DroppedColumns(MakePair(t.tables.Prev.ID, t.tables.Next.ID))
}

func (t TableDiffer) CreatedForeignKeys() []ForeignKeyWalker {
	var fks []ForeignKeyWalker
	for _, nextfk := range t.NextForeignKeys() {
		match := false
		for _, prevfk := range t.PreviousForeignKeys() {
			if t.db.foreignKeysMatch(prevfk, nextfk) {
				match = true
				break
			}
		}
		if !match {
			fks = append(fks, nextfk)
		}
	}
	return fks
}

func (t TableDiffer) DroppedForeignKeys() []ForeignKeyWalker {
	var fks []ForeignKeyWalker
	for _, prevfk := range t.PreviousForeignKeys() {
		match := false
		for _, nextfk := range t.NextForeignKeys() {
			if t.db.foreignKeysMatch(nextfk, prevfk) {
				match = true
				break
			}
		}
		if !match {
			fks = append(fks, prevfk)
		}
	}
	return fks
}

func (t TableDiffer) CreatedIndexes() []IndexWalker {
	var indexes []IndexWalker
	for _, next := range t.NextIndexes() {
		match := false
		for _, prev := range t.PreviousIndexes() {
			if t.db.indexesMatch(prev, next) {
				match = true
				break
			}
		}
		if !match {
			indexes = append(indexes, next)
		}
	}
	return indexes
}

func (t TableDiffer) DroppedIndexes() []IndexWalker {
	var indexes []IndexWalker
	for _, prev := range t.PreviousIndexes() {
		match := false
		for _, next := range t.NextIndexes() {
			if t.db.indexesMatch(next, prev) {
				match = true
				break
			}
		}
		if !match {
			indexes = append(indexes, prev)
		}
	}
	return indexes
}

func (t TableDiffer) ForeignKeyPairs() []Pair[ForeignKeyWalker] {
	var fks []Pair[ForeignKeyWalker]

	for _, nextfk := range t.NextForeignKeys() {
		for _, prevfk := range t.PreviousForeignKeys() {
			if t.db.foreignKeysMatch(prevfk, nextfk) {
				fks = append(fks, MakePair(prevfk, nextfk))
			}
		}
	}

	return fks
}

func (t TableDiffer) IndexPairs() []Pair[IndexWalker] {
	var indexes []Pair[IndexWalker]

	for _, next := range t.NextIndexes() {
		for _, prev := range t.PreviousIndexes() {
			if t.db.indexesMatch(prev, next) {
				indexes = append(indexes, MakePair(prev, next))
			}
		}
	}

	return indexes
}

func (t TableDiffer) CreatedPrimaryKey() mo.Option[IndexWalker] {
	prev := t.db.Databases.Prev.WalkTable(t.tables.Prev.ID).PrimaryKey()
	next := t.db.Databases.Next.WalkTable(t.tables.Next.ID).PrimaryKey()
	if prev.IsAbsent() && next.IsPresent() {
		return mo.Some(next.MustGet())
	}

	return mo.None[IndexWalker]()
}

func (t TableDiffer) DroppedPrimaryKey() mo.Option[IndexWalker] {
	prev := t.db.Databases.Prev.WalkTable(t.tables.Prev.ID).PrimaryKey()
	next := t.db.Databases.Next.WalkTable(t.tables.Next.ID).PrimaryKey()
	if prev.IsPresent() && next.IsAbsent() {
		return mo.Some(next.MustGet())
	}
	return mo.None[IndexWalker]()
}

func (t TableDiffer) PrimaryKeyChanged() bool {
	prev, next := t.Previous().PrimaryKey(), t.Next().PrimaryKey()
	if prev.IsAbsent() || next.IsAbsent() {
		return false
	}

	prevCols, nextCols := prev.MustGet().Columns(), next.MustGet().Columns()
	if len(prevCols) != len(nextCols) {
		return true
	}

	for _, cols := range lo.Zip2(prevCols, nextCols) {
		if cols.A.Name() != cols.B.Name() {
			return true
		}

		if cols.A.SortOrder() != cols.B.SortOrder() {
			return true
		}
	}

	return false
}

func (t TableDiffer) RenamedPrimaryKey() bool {
	prev, next := t.Previous().PrimaryKey(), t.Next().PrimaryKey()
	if prev.IsAbsent() || next.IsAbsent() {
		return false
	}

	return prev.MustGet().Name() != next.MustGet().Name()
}

func (t TableDiffer) PreviousForeignKeys() []ForeignKeyWalker { return t.Previous().ForeignKeys() }
func (t TableDiffer) NextForeignKeys() []ForeignKeyWalker     { return t.Next().ForeignKeys() }
func (t TableDiffer) PreviousIndexes() []IndexWalker          { return t.Previous().Indexes() }
func (t TableDiffer) NextIndexes() []IndexWalker              { return t.Next().Indexes() }
func (t TableDiffer) Previous() TableWalker                   { return t.db.Databases.Prev.WalkTable(t.tables.Prev.ID) }
func (t TableDiffer) Next() TableWalker                       { return t.db.Databases.Next.WalkTable(t.tables.Next.ID) }

type (
	DropExtension   struct{ Extension ExtensionID }
	CreateExtension struct{ Extension ExtensionID }
	AlterExtension  struct {
		Extensions Pair[ExtensionID]
		Changes    ExtensionChanges
	}
	CreateEnum struct{ Enum EnumID }
	AlterEnum  struct {
		Enums           Pair[EnumID]
		CreatedVariants []string
		DroppedVariants []string
	}
	DropForeignKey struct{ ForeignKey ForeignKeyID }
	DropIndex      struct{ Index IndexID }
	AlterTable     struct {
		Tables  Pair[TableID]
		Changes []TableChange
	}

	DropTable struct{ Table TableID }

	DropEnum    struct{ Enum EnumID }
	CreateTable struct{ Table TableID }

	CreateIndex      struct{ Index IndexID }
	RenameForeignKey struct{ ForeignKeys Pair[ForeignKeyID] }

	AddForeignKey struct{ ForeignKey ForeignKeyID }
	RenameIndex   struct{ Index Pair[IndexID] }
)

const (
	dropExtensionStep = iota
	createExtensionStep
	alterExtensionStep
	createEnumStep
	alterEnumStep
	dropForeignKeyStep
	dropIndexStep
	alterTableStep

	// Order matters: we must drop tables before we create indexes,
	// because on Postgres we may create indexes whose names
	// clash with the names of indexes on the dropped tables.
	dropTableStep

	// Order matters:
	// - We must drop enums before we create tables, because the new tables
	//   might be named the same as the dropped enum, and that conflicts on
	//   postgres.
	// - We must drop enums after we drop tables, or dropping the enum will
	//   fail on postgres because objects (=tables) still depend on them.
	dropEnumStep
	createTableStep

	// Order matters: we must create indexes after ALTER TABLEs because the indexes can be
	// on fields that are dropped/created there.
	createIndexStep
	renameForeignKeyStep

	// Order matters: this needs to come after create_indexes, because the foreign keys can depend on unique
	// indexes created there.
	addForeignKeyStep
	renameIndexStep
)

func stepSortIndex(step MigrationStep) int {
	switch step.(type) {
	case DropExtension:
		return dropExtensionStep
	case CreateExtension:
		return createExtensionStep
	case AlterExtension:
		return alterExtensionStep
	case CreateEnum:
		return createEnumStep
	case AlterEnum:
		return alterEnumStep
	case DropForeignKey:
		return dropForeignKeyStep
	case DropIndex:
		return dropIndexStep
	case AlterTable:
		return alterTableStep
	case DropTable:
		return dropTableStep
	case DropEnum:
		return dropEnumStep
	case CreateTable:
		return createTableStep
	case CreateIndex:
		return createIndexStep
	case RenameForeignKey:
		return renameForeignKeyStep
	case AddForeignKey:
		return addForeignKeyStep
	case RenameIndex:
		return renameIndexStep
	default:
		panic(fmt.Sprintf("unknown step type %T", step))
	}
}

type byStepType struct{ steps []MigrationStep }

func (s byStepType) Len() int      { return len(s.steps) }
func (s byStepType) Swap(i, j int) { s.steps[i], s.steps[j] = s.steps[j], s.steps[i] }
func (s byStepType) Less(i, j int) bool {
	si, sj := stepSortIndex(s.steps[i]), stepSortIndex(s.steps[j])
	if si == sj {
		switch s.steps[i].(type) {
		case DropExtension:
			return s.steps[i].(DropExtension).Extension < s.steps[j].(DropExtension).Extension
		case CreateExtension:
			return s.steps[i].(CreateExtension).Extension < s.steps[j].(CreateExtension).Extension
		case AlterExtension:
			return s.steps[i].(AlterExtension).Extensions.Next < s.steps[j].(AlterExtension).Extensions.Next
		case CreateEnum:
			return s.steps[i].(CreateEnum).Enum < s.steps[j].(CreateEnum).Enum
		case AlterEnum:
			return s.steps[i].(AlterEnum).Enums.Next < s.steps[j].(AlterEnum).Enums.Next
		case DropForeignKey:
			return s.steps[i].(DropForeignKey).ForeignKey < s.steps[j].(DropForeignKey).ForeignKey
		case DropIndex:
			return s.steps[i].(DropIndex).Index < s.steps[j].(DropIndex).Index
		case AlterTable:
			return s.steps[i].(AlterTable).Tables.Next < s.steps[j].(AlterTable).Tables.Next
		case DropTable:
			return s.steps[i].(DropTable).Table < s.steps[j].(DropTable).Table
		case DropEnum:
			return s.steps[i].(DropEnum).Enum < s.steps[j].(DropEnum).Enum
		case CreateTable:
			return s.steps[i].(CreateTable).Table < s.steps[j].(CreateTable).Table
		case CreateIndex:
			return s.steps[i].(CreateIndex).Index < s.steps[j].(CreateIndex).Index
		case RenameForeignKey:
			return s.steps[i].(RenameForeignKey).ForeignKeys.Next < s.steps[j].(RenameForeignKey).ForeignKeys.Next
		case AddForeignKey:
			return s.steps[i].(AddForeignKey).ForeignKey < s.steps[j].(AddForeignKey).ForeignKey
		case RenameIndex:
			return s.steps[i].(RenameIndex).Index.Next < s.steps[j].(RenameIndex).Index.Next
		}
		return false
	}
	return si < sj
}

type (
	TableChange interface{ tablechange() }
	AddColumn   struct{ Column ColumnID }
	AlterColumn struct {
		Columns    Pair[ColumnID]
		Changes    ColumnChanges
		TypeChange ColumnTypeChange
	}
	DropColumn            struct{ Column ColumnID }
	DropAndRecreateColumn struct {
		Columns Pair[ColumnID]
		Changes ColumnChanges
	}
	RenameColumn     struct{ Columns Pair[ColumnID] }
	AddPrimaryKey    struct{}
	DropPrimaryKey   struct{}
	RenamePrimaryKey struct{}
)

func (AddColumn) tablechange()             {}
func (AlterColumn) tablechange()           {}
func (DropColumn) tablechange()            {}
func (DropAndRecreateColumn) tablechange() {}
func (AddPrimaryKey) tablechange()         {}
func (RenameColumn) tablechange()          {}
func (DropPrimaryKey) tablechange()        {}
func (RenamePrimaryKey) tablechange()      {}

func (DropExtension) step()    {}
func (CreateExtension) step()  {}
func (AlterExtension) step()   {}
func (CreateEnum) step()       {}
func (AlterEnum) step()        {}
func (DropForeignKey) step()   {}
func (DropIndex) step()        {}
func (AlterTable) step()       {}
func (DropTable) step()        {}
func (DropEnum) step()         {}
func (CreateTable) step()      {}
func (CreateIndex) step()      {}
func (RenameForeignKey) step() {}
func (AddForeignKey) step()    {}
func (RenameIndex) step()      {}

type ColumnChanges struct {
	TypeChange ColumnTypeChange
	Changes    ColumnChange
}

func (c ColumnChanges) DiffersInSomething() bool {
	return c.Changes != ColumnChangeNone
}

func (c ColumnChanges) AutoIncrementChanged() bool {
	return c.Changes&ColumnChangeAutoIncrement != 0
}

func (c ColumnChanges) ArityChanged() bool {
	return c.Changes&ColumnChangeArity != 0
}

func (c ColumnChanges) TypeChanged() bool {
	return c.Changes&ColumnChangeType != 0
}

func (c ColumnChanges) DefaultChanged() bool {
	return c.Changes == ColumnChangeDefault
}

func (c ColumnChanges) OnlyTypeChanged() bool {
	return c.Changes == ColumnChangeDefault
}

type ExtensionChanges int

const (
	ExtensionChangeNone    ExtensionChanges = 0
	ExtensionChangeVersion ExtensionChanges = 1 << (iota - 1)
	ExtensionChangeSchema
)

type ColumnChange int

const (
	ColumnChangeNone    ColumnChange = 0
	ColumnChangeDefault ColumnChange = 1 << (iota - 1)
	ColumnChangeArity
	ColumnChangeType
	ColumnChangeAutoIncrement
)

type ColumnTypeChange int

const (
	ColumnTypeChangeNone ColumnTypeChange = iota
	ColumnTypeChangeSafeCast
	ColumnTypeChangeRiskyCast
	ColumnTypeChangeNotCastable
)

type AlterColumnChanges struct {
	SetDefault  Value
	DropDefault bool
	SetNotNull  bool
	DropNotNull bool
	SetType     bool
	AddSequence bool
}
