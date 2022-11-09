package sqlspec

import (
	"database/sql"
	"fmt"
	"sort"

	"ksl/sqlutil"
)

type (
	Driver interface {
		Differ
		ExecQuerier
		Inspector
		Planner
		PlanApplier
	}

	Differ interface {
		RealmDiff(from, to *Realm) ([]SchemaChange, error)
		SchemaDiff(from, to *Schema) ([]SchemaChange, error)
		TableDiff(from, to *Table) ([]SchemaChange, error)
		RoleDiff(from, to *Role) ([]SchemaChange, error)
		QueryDiff(from, to *Query) ([]SchemaChange, error)
	}

	Diff struct {
		DiffDriver
	}

	DiffDriver interface {
		EnumDiff(from, to *Enum) ([]SchemaChange, error)
		SchemaAttrDiff(from, to *Schema) []SchemaChange
		TableAttrDiff(from, to *Table) ([]SchemaChange, error)
		ColumnChange(fromT *Table, from, to *Column) (ChangeKind, error)
		IndexAttrChanged(from, to []Attr) bool
		IndexPartAttrChanged(from, to *IndexPart) bool
		IsGeneratedIndexName(*Table, *Index) bool
		ReferenceChanged(from, to string) bool
	}

	DiffNormalizer interface {
		Normalize(from, to *Table) error
	}
)

type (
	SchemaChange interface{ change() }
	SchemaClause interface{ clause() }

	AddEnum struct {
		E *Enum
	}

	ModifyEnum struct {
		E       *Enum
		Changes []SchemaChange
	}

	DropEnum struct {
		E *Enum
	}

	AddRole struct {
		R *Role
	}

	ModifyRole struct {
		From, To *Role
		Change   RoleChangeKind
	}

	DropRole struct {
		R *Role
	}

	AddQuery struct {
		Q *Query
	}

	ModifyQuery struct {
		From, To *Query
		Change   QueryChangeKind
	}

	DropQuery struct {
		Q *Query
	}

	AddQueryToRole struct {
		R *Role
		Q *Query
	}

	DropQueryFromRole struct {
		R *Role
		Q *Query
	}

	AddSchema struct {
		S     *Schema
		Extra []SchemaClause
	}

	DropSchema struct {
		S     *Schema
		Extra []SchemaClause
	}

	ModifySchema struct {
		S       *Schema
		Changes []SchemaChange
	}

	AddTable struct {
		T     *Table
		Extra []SchemaClause
	}

	DropTable struct {
		T     *Table
		Extra []SchemaClause
	}

	ModifyTable struct {
		T       *Table
		Changes []SchemaChange
	}

	RenameTable struct {
		From, To *Table
	}

	AddColumn struct {
		C *Column
	}

	DropColumn struct {
		C *Column
	}

	ModifyColumn struct {
		From, To *Column
		Change   ChangeKind
	}

	RenameColumn struct {
		From, To *Column
	}

	AddIndex struct {
		I *Index
	}

	DropIndex struct {
		I *Index
	}

	ModifyIndex struct {
		From, To *Index
		Change   ChangeKind
	}

	RenameIndex struct {
		From, To *Index
	}

	AddForeignKey struct {
		F *ForeignKey
	}

	DropForeignKey struct {
		F *ForeignKey
	}

	ModifyForeignKey struct {
		From, To *ForeignKey
		Change   ChangeKind
	}

	AddCheck struct {
		C *Check
	}

	DropCheck struct {
		C *Check
	}

	ModifyCheck struct {
		From, To *Check
		Change   ChangeKind
	}

	AddAttr struct {
		A Attr
	}

	DropAttr struct {
		A Attr
	}

	ModifyAttr struct {
		From, To Attr
	}

	AddEnumValue struct {
		E *Enum
		V string
	}

	IfExists    struct{}
	IfNotExists struct{}
)

type ChangeKind uint

const (
	NoChange ChangeKind = 0

	ChangeAttr ChangeKind = 1 << (iota - 1)
	ChangeCharset
	ChangeCollate
	ChangeComment

	ChangeNullability
	ChangeType
	ChangeDefault
	ChangeGenerated
	ChangeUnique
	ChangeParts
	ChangeColumn
	ChangeRefColumn
	ChangeRefTable
	ChangeUpdateAction
	ChangeDeleteAction
)

func (k ChangeKind) Is(c ChangeKind) bool {
	return k == c || k&c != 0
}

type Changes []SchemaChange

func (*AddAttr) change()          {}
func (*DropAttr) change()         {}
func (*ModifyAttr) change()       {}
func (*AddSchema) change()        {}
func (*DropSchema) change()       {}
func (*ModifySchema) change()     {}
func (*AddTable) change()         {}
func (*DropTable) change()        {}
func (*ModifyTable) change()      {}
func (*RenameTable) change()      {}
func (*AddIndex) change()         {}
func (*DropIndex) change()        {}
func (*ModifyIndex) change()      {}
func (*RenameIndex) change()      {}
func (*AddColumn) change()        {}
func (*DropColumn) change()       {}
func (*ModifyColumn) change()     {}
func (*RenameColumn) change()     {}
func (*AddForeignKey) change()    {}
func (*DropForeignKey) change()   {}
func (*ModifyForeignKey) change() {}
func (*AddCheck) change()         {}
func (*DropCheck) change()        {}
func (*ModifyCheck) change()      {}
func (*AddEnum) change()          {}
func (*ModifyEnum) change()       {}
func (*DropEnum) change()         {}
func (*AddEnumValue) change()     {}

func (*AddRole) change()           {}
func (*ModifyRole) change()        {}
func (*DropRole) change()          {}
func (*AddQuery) change()          {}
func (*ModifyQuery) change()       {}
func (*DropQuery) change()         {}
func (*AddQueryToRole) change()    {}
func (*DropQueryFromRole) change() {}

func (*IfExists) clause()    {}
func (*IfNotExists) clause() {}

type RoleChangeKind uint

const (
	NoRoleChange       RoleChangeKind = 0
	RoleDefaultChanged                = 1 << (iota - 1)
	RoleQueriesChanged
)

type QueryChangeKind uint

const (
	NoQueryChange         QueryChangeKind = 0
	QueryStatementChanged                 = 1 << (iota - 1)
)

func (d *Diff) RealmDiff(from, to *Realm) ([]SchemaChange, error) {
	var changes []SchemaChange
	// Drop or modify
	for _, s1 := range from.Schemas {
		s2, ok := to.Schema(s1.Name)
		if !ok {
			changes = append(changes, &DropSchema{S: s1})
			continue
		}
		change, err := d.SchemaDiff(s1, s2)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change...)
	}
	// Add schemas.
	for _, s1 := range to.Schemas {
		if _, ok := from.Schema(s1.Name); ok {
			continue
		}
		changes = append(changes, &AddSchema{S: s1})
		for _, t := range s1.Tables {
			changes = append(changes, &AddTable{T: t})
		}
	}

	// Add roles.
	for _, r1 := range to.Roles {
		if _, ok := from.Role(r1.Name); !ok {
			changes = append(changes, &AddRole{R: r1})
		}
	}
	// Drop or modify roles.
	for _, r1 := range from.Roles {
		r2, ok := to.Role(r1.Name)
		if !ok {
			changes = append(changes, &DropRole{R: r1})
			continue
		}
		change, err := d.RoleDiff(r1, r2)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change...)
	}

	// Add queries.
	for _, q1 := range to.Queries {
		if _, ok := from.Role(q1.Name); !ok {
			changes = append(changes, &AddQuery{Q: q1})
		}
	}

	// Drop or modify queries.
	for _, q1 := range from.Queries {
		q2, ok := to.Query(q1.Name)
		if !ok {
			changes = append(changes, &DropQuery{Q: q1})
			continue
		}
		change, err := d.QueryDiff(q1, q2)
		if err != nil {
			return nil, err
		}

		changes = append(changes, change...)
	}

	return changes, nil
}

func (d *Diff) RoleDiff(from, to *Role) ([]SchemaChange, error) {
	change := NoRoleChange
	var changes []SchemaChange
	if to.Default != from.Default {
		change |= RoleDefaultChanged
	}

	fromValues := map[string]struct{}{}
	toValues := map[string]struct{}{}
	for _, q := range from.Queries {
		fromValues[q.Name] = struct{}{}
	}
	for _, q := range to.Queries {
		toValues[q.Name] = struct{}{}
	}

	for qf := range fromValues {
		if _, ok := toValues[qf]; !ok {
			query, ok := from.Realm.Query(qf)
			if !ok {
				return nil, fmt.Errorf("query %q not found", qf)
			}
			changes = append(changes, &DropQueryFromRole{R: to, Q: query})
		}
	}

	for qt := range toValues {
		if _, ok := fromValues[qt]; !ok {
			query, ok := to.Realm.Query(qt)
			if !ok {
				return nil, fmt.Errorf("query %q not found", qt)
			}
			changes = append(changes, &AddQueryToRole{R: to, Q: query})
		}
	}

	if change != NoRoleChange {
		changes = append(changes, &ModifyRole{From: from, To: to, Change: change})
	}

	return changes, nil
}

func (d *Diff) QueryDiff(from, to *Query) ([]SchemaChange, error) {
	var changes []SchemaChange
	change := NoQueryChange
	if from.Statement != to.Statement {
		change |= QueryStatementChanged
	}

	if change != NoQueryChange {
		changes = append(changes, &ModifyQuery{From: from, To: to, Change: change})
	}
	return changes, nil
}

func (d *Diff) SchemaDiff(from, to *Schema) ([]SchemaChange, error) {
	if from.Name != to.Name {
		return nil, fmt.Errorf("mismatched schema names: %q != %q", from.Name, to.Name)
	}
	var changes []SchemaChange

	// Add enums.
	for _, e1 := range to.Enums {
		if _, ok := from.Enum(e1.Name); !ok {
			changes = append(changes, &AddEnum{E: e1})
		}
	}

	// Drop or modify attributes (collations, charset, etc).
	if change := d.SchemaAttrDiff(from, to); len(change) > 0 {
		changes = append(changes, &ModifySchema{
			S:       to,
			Changes: change,
		})
	}

	// Drop or modify tables.
	for _, t1 := range from.Tables {
		t2, ok := to.Table(t1.Name)
		if !ok {
			changes = append(changes, &DropTable{T: t1})
			continue
		}
		change, err := d.TableDiff(t1, t2)
		if err != nil {
			return nil, err
		}
		if len(change) > 0 {
			changes = append(changes, &ModifyTable{
				T:       t2,
				Changes: change,
			})
		}
	}
	// Add tables.
	for _, t1 := range to.Tables {
		if _, ok := from.Table(t1.Name); !ok {
			changes = append(changes, &AddTable{T: t1})
		}
	}

	// Drop or modify enums.
	for _, e1 := range from.Enums {
		e2, ok := to.Enum(e1.Name)
		if !ok {
			changes = append(changes, &DropEnum{E: e1})
			continue
		}
		change, err := d.EnumDiff(e1, e2)
		if err != nil {
			return nil, err
		}

		if len(change) > 0 {
			changes = append(changes, &ModifyEnum{
				E:       e2,
				Changes: change,
			})
		}
	}

	return changes, nil
}

func (d *Diff) TableDiff(from, to *Table) ([]SchemaChange, error) {
	if from.Name != to.Name {
		return nil, fmt.Errorf("mismatched table names: %q != %q", from.Name, to.Name)
	}
	// Normalizing tables before starting the diff process.
	if n, ok := d.DiffDriver.(DiffNormalizer); ok {
		if err := n.Normalize(from, to); err != nil {
			return nil, err
		}
	}
	var changes []SchemaChange
	if from.Name != to.Name {
		return nil, fmt.Errorf("mismatched table names: %q != %q", from.Name, to.Name)
	}
	// PK modification is not supported.
	if pk1, pk2 := from.PrimaryKey, to.PrimaryKey; (pk1 != nil) != (pk2 != nil) || (pk1 != nil) && d.pkChange(pk1, pk2) != NoChange {
		return nil, fmt.Errorf("changing %q table primary key is not supported", to.Name)
	}

	// Drop or modify attributes (collations, checks, etc).
	change, err := d.TableAttrDiff(from, to)
	if err != nil {
		return nil, err
	}
	changes = append(changes, change...)

	// Drop or modify columns.
	for _, c1 := range from.Columns {
		c2, ok := to.Column(c1.Name)
		if !ok {
			changes = append(changes, &DropColumn{C: c1})
			continue
		}
		change, err := d.ColumnChange(from, c1, c2)
		if err != nil {
			return nil, err
		}
		if change != NoChange {
			changes = append(changes, &ModifyColumn{
				From:   c1,
				To:     c2,
				Change: change,
			})
		}
	}
	// Add columns.
	for _, c1 := range to.Columns {
		if _, ok := from.Column(c1.Name); !ok {
			changes = append(changes, &AddColumn{C: c1})
		}
	}

	// Index changes.
	changes = append(changes, d.indexDiff(from, to)...)

	// Drop or modify foreign-keys.
	for _, fk1 := range from.ForeignKeys {
		fk2, ok := to.ForeignKey(fk1.Name)
		if !ok {
			changes = append(changes, &DropForeignKey{F: fk1})
			continue
		}
		if change := d.fkChange(fk1, fk2); change != NoChange {
			changes = append(changes, &ModifyForeignKey{
				From:   fk1,
				To:     fk2,
				Change: change,
			})
		}
	}
	// Add foreign-keys.
	for _, fk1 := range to.ForeignKeys {
		if _, ok := from.ForeignKey(fk1.Name); !ok {
			changes = append(changes, &AddForeignKey{F: fk1})
		}
	}
	return changes, nil
}

func (d *Diff) indexDiff(from, to *Table) []SchemaChange {
	var (
		changes []SchemaChange
		exists  = make(map[*Index]bool)
	)
	// Drop or modify indexes.
	for _, idx1 := range from.Indexes {
		idx2, ok := to.Index(idx1.Name)
		// Found directly.
		if ok {
			if change := d.indexChange(idx1, idx2); change != NoChange {
				changes = append(changes, &ModifyIndex{
					From:   idx1,
					To:     idx2,
					Change: change,
				})
			}
			exists[idx2] = true
			continue
		}
		// Found indirectly.
		if d.IsGeneratedIndexName(from, idx1) {
			if idx2, ok := d.similarUnnamedIndex(to, idx1); ok {
				exists[idx2] = true
				continue
			}
		}
		// Not found.
		changes = append(changes, &DropIndex{I: idx1})
	}
	// Add indexes.
	for _, idx := range to.Indexes {
		if exists[idx] {
			continue
		}
		if _, ok := from.Index(idx.Name); !ok {
			changes = append(changes, &AddIndex{I: idx})
		}
	}
	return changes
}

func (d *Diff) pkChange(from, to *Index) ChangeKind {
	change := d.indexChange(from, to)
	return change & ^ChangeUnique
}

func (d *Diff) indexChange(from, to *Index) ChangeKind {
	var change ChangeKind
	if from.Unique != to.Unique {
		change |= ChangeUnique
	}
	if d.IndexAttrChanged(from.Attrs, to.Attrs) {
		change |= ChangeAttr
	}
	change |= d.partsChange(from.Parts, to.Parts)
	change |= CommentChange(from.Attrs, to.Attrs)
	return change
}

func (d *Diff) partsChange(from, to []*IndexPart) ChangeKind {
	if len(from) != len(to) {
		return ChangeParts
	}
	sort.Slice(to, func(i, j int) bool { return to[i].Seq < to[j].Seq })
	sort.Slice(from, func(i, j int) bool { return from[i].Seq < from[j].Seq })
	for i := range from {
		switch {
		case from[i].Descending != to[i].Descending || d.IndexPartAttrChanged(from[i], to[i]):
			return ChangeParts
		case from[i].Column != nil && to[i].Column != nil:
			if from[i].Column.Name != to[i].Column.Name {
				return ChangeParts
			}
		case from[i].Expr != nil && to[i].Expr != nil:
			x1, x2 := from[i].Expr.(*RawExpr).Expr, to[i].Expr.(*RawExpr).Expr
			if x1 != x2 && x1 != sqlutil.MayWrap(x2) {
				return ChangeParts
			}
		default: // (C1 != nil) != (C2 != nil) || (X1 != nil) != (X2 != nil).
			return ChangeParts
		}
	}
	return NoChange
}

func (d *Diff) fkChange(from, to *ForeignKey) ChangeKind {
	var change ChangeKind
	switch {
	case from.Table.Name != to.Table.Name:
		change |= ChangeRefTable | ChangeRefColumn
	case len(from.RefColumns) != len(to.RefColumns):
		change |= ChangeRefColumn
	default:
		for i := range from.RefColumns {
			if from.RefColumns[i].Name != to.RefColumns[i].Name {
				change |= ChangeRefColumn
			}
		}
	}
	switch {
	case len(from.Columns) != len(to.Columns):
		change |= ChangeColumn
	default:
		for i := range from.Columns {
			if from.Columns[i].Name != to.Columns[i].Name {
				change |= ChangeColumn
			}
		}
	}
	if d.ReferenceChanged(from.OnUpdate, to.OnUpdate) {
		change |= ChangeUpdateAction
	}
	if d.ReferenceChanged(from.OnDelete, to.OnDelete) {
		change |= ChangeDeleteAction
	}
	return change
}

func (d *Diff) similarUnnamedIndex(t *Table, idx1 *Index) (*Index, bool) {
	for _, idx2 := range t.Indexes {
		if idx2.Name != "" || len(idx2.Parts) != len(idx1.Parts) || idx2.Unique != idx1.Unique {
			continue
		}
		if d.partsChange(idx1.Parts, idx2.Parts) == NoChange {
			return idx2, true
		}
	}
	return nil, false
}

func CommentChange(from, to []Attr) ChangeKind {
	var c1, c2 Comment
	if has(from, &c1) != has(to, &c2) || c1.Text != c2.Text {
		return ChangeComment
	}
	return NoChange
}

func SchemaFKs(s *Schema, rows *sql.Rows) error {
	for rows.Next() {
		var name, table, column, tSchema, refTable, refColumn, refSchema, updateRule, deleteRule string
		if err := rows.Scan(&name, &table, &column, &tSchema, &refTable, &refColumn, &refSchema, &updateRule, &deleteRule); err != nil {
			return err
		}
		t, ok := s.Table(table)
		if !ok {
			return fmt.Errorf("table %q was not found in schema", table)
		}
		fk, ok := t.ForeignKey(name)
		if !ok {
			rt := s.Realm.GetOrCreateTable(refSchema, refTable)
			fk = NewForeignKey(name).SetTable(t).SetRefTable(rt).SetOnDelete(deleteRule).SetOnUpdate(updateRule)
			t.AddForeignKeys(fk)
		}
		c, ok := t.Column(column)
		if !ok {
			return fmt.Errorf("column %q was not found for fk %q", column, fk.Name)
		}
		if _, ok := fk.Column(c.Name); !ok {
			fk.AddColumns(c)
		}
		rc := fk.RefTable.GetOrCreateColumn(refColumn)
		if _, ok := fk.RefColumn(rc.Name); !ok {
			fk.AddRefColumns(rc)
		}
	}
	return nil
}

func LinkSchemaTables(schemas []*Schema) {
	byName := make(map[string]map[string]*Table)
	for _, s := range schemas {
		byName[s.Name] = make(map[string]*Table)
		for _, t := range s.Tables {
			t.Schema = s
			byName[s.Name][t.Name] = t
		}
	}
	for _, s := range schemas {
		for _, t := range s.Tables {
			for _, fk := range t.ForeignKeys {
				rs, ok := byName[fk.RefTable.Name]
				if !ok {
					continue
				}
				ref, ok := rs[fk.RefTable.Name]
				if !ok {
					continue
				}
				fk.RefTable = ref
				for i, c := range fk.RefColumns {
					rc, ok := ref.Column(c.Name)
					if ok {
						fk.RefColumns[i] = rc
					}
				}
			}
		}
	}
}

func ReverseChanges(c []SchemaChange) {
	for i, n := 0, len(c); i < n/2; i++ {
		c[i], c[n-i-1] = c[n-i-1], c[i]
	}
}

func CommentDiff(from, to []Attr) SchemaChange {
	var fromC, toC Comment
	switch fromHas, toHas := has(from, &fromC), has(to, &toC); {
	case !fromHas && !toHas:
	case !fromHas && toC.Text != "":
		return &AddAttr{
			A: &toC,
		}
	case !toHas:
		return &ModifyAttr{
			From: &fromC,
			To:   &toC,
		}
	default:
		v1, err1 := sqlutil.Unquote(fromC.Text)
		v2, err2 := sqlutil.Unquote(toC.Text)
		if err1 == nil && err2 == nil && v1 != v2 {
			return &ModifyAttr{
				From: &fromC,
				To:   &toC,
			}
		}
	}
	return nil
}

func CheckDiff(from, to *Table, compare ...func(c1, c2 *Check) bool) []SchemaChange {
	var changes []SchemaChange
	// Drop or modify checks.
	for _, c1 := range checks(from.Attrs) {
		switch c2, ok := similarCheck(to.Attrs, c1); {
		case !ok:
			changes = append(changes, &DropCheck{
				C: c1,
			})
		case len(compare) == 1 && !compare[0](c1, c2):
			changes = append(changes, &ModifyCheck{
				From: c1,
				To:   c2,
			})
		}
	}
	// Add checks.
	for _, c1 := range checks(to.Attrs) {
		if _, ok := similarCheck(from.Attrs, c1); !ok {
			changes = append(changes, &AddCheck{
				C: c1,
			})
		}
	}
	return changes
}

func checks(attr []Attr) (checks []*Check) {
	for i := range attr {
		if c, ok := attr[i].(*Check); ok {
			checks = append(checks, c)
		}
	}
	return checks
}

func similarCheck(attrs []Attr, c *Check) (*Check, bool) {
	var byName, byExpr *Check
	for i := 0; i < len(attrs) && (byName == nil || byExpr == nil); i++ {
		check, ok := attrs[i].(*Check)
		if !ok {
			continue
		}
		if check.Name != "" && check.Name == c.Name {
			byName = check
		}
		if check.Expr == c.Expr {
			byExpr = check
		}
	}

	if byName != nil {
		return byName, true
	}
	if byExpr != nil {
		return byExpr, true
	}
	return nil, false
}
