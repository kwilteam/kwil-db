package sqlspec

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"ksl/sqlutil"
)

type diff struct{}

func NewDiffer() Differ {
	return &Diff{DiffDriver: &diff{}}
}

func (d *diff) SchemaAttrDiff(_, _ *Schema) []SchemaChange {
	return nil
}

func (d *diff) EnumDiff(from, to *Enum) ([]SchemaChange, error) {
	var changes []SchemaChange

	fromValues := map[string]struct{}{}
	toValues := map[string]struct{}{}
	for _, v := range from.Values {
		fromValues[v] = struct{}{}
	}
	for _, v := range to.Values {
		toValues[v] = struct{}{}
	}

	for ef := range fromValues {
		if _, ok := toValues[ef]; !ok {
			return nil, fmt.Errorf("cannot drop enum value %q from %q", ef, from.Name)
		}
	}

	for et := range toValues {
		if _, ok := fromValues[et]; !ok {
			changes = append(changes, &AddEnumValue{E: to, V: et})
		}
	}

	return changes, nil
}

func (d *diff) TableAttrDiff(from, to *Table) ([]SchemaChange, error) {
	var changes []SchemaChange
	if change := CommentDiff(from.Attrs, to.Attrs); change != nil {
		changes = append(changes, change)
	}
	if err := d.partitionChanged(from, to); err != nil {
		return nil, err
	}
	return append(changes, CheckDiff(from, to, func(c1, c2 *Check) bool {
		return has(c1.Attrs, &NoInherit{}) == has(c2.Attrs, &NoInherit{})
	})...), nil
}

func (d *diff) ColumnChange(_ *Table, from, to *Column) (ChangeKind, error) {
	change := CommentChange(from.Attrs, to.Attrs)
	if from.Type.Nullable != to.Type.Nullable {
		change |= ChangeNullability
	}
	changed, err := d.typeChanged(from, to)
	if err != nil {
		return NoChange, err
	}
	if changed {
		change |= ChangeType
	}
	if changed, err = d.defaultChanged(from, to); err != nil {
		return NoChange, err
	}
	if changed {
		change |= ChangeDefault
	}
	if identityChanged(from.Attrs, to.Attrs) {
		change |= ChangeAttr
	}
	if changed, err = d.generatedChanged(from, to); err != nil {
		return NoChange, err
	}
	if changed {
		change |= ChangeGenerated
	}
	return change, nil
}

func (d *diff) defaultChanged(from, to *Column) (bool, error) {
	d1, ok1 := DefaultValue(from)
	d2, ok2 := DefaultValue(to)

	if ok1 != ok2 {
		return true, nil
	}
	if sqlutil.TrimCast(d1) == sqlutil.TrimCast(d2) {
		return false, nil
	}
	return true, nil
}

func (*diff) generatedChanged(from, to *Column) (bool, error) {
	var fromX, toX GeneratedExpr
	switch fromHas, toHas := has(from.Attrs, &fromX), has(to.Attrs, &toX); {
	case fromHas && toHas && sqlutil.MayWrap(fromX.Expr) != sqlutil.MayWrap(toX.Expr):
		return false, fmt.Errorf("changing the generation expression for a column %q is not supported", from.Name)
	case !fromHas && toHas:
		return false, fmt.Errorf("changing column %q to generated column is not supported (drop and add is required)", from.Name)
	default:
		return fromHas && !toHas, nil
	}
}

func (*diff) partitionChanged(from, to *Table) error {
	var fromP, toP Partition
	switch fromHas, toHas := has(from.Attrs, &fromP), has(to.Attrs, &toP); {
	case fromHas && !toHas:
		return fmt.Errorf("partition key cannot be dropped from %q (drop and add is required)", from.Name)
	case !fromHas && toHas:
		return fmt.Errorf("partition key cannot be added to %q (drop and add is required)", to.Name)
	case fromHas && toHas:
		s1, err := formatPartition(fromP)
		if err != nil {
			return err
		}
		s2, err := formatPartition(toP)
		if err != nil {
			return err
		}
		if s1 != s2 {
			return fmt.Errorf("partition key of table %q cannot be changed from %s to %s (drop and add is required)", to.Name, s1, s2)
		}
	}
	return nil
}

func (d *diff) IsGeneratedIndexName(t *Table, idx *Index) bool {
	columns := make([]*Column, len(idx.Parts))
	for i, p := range idx.Parts {
		if p.Column == nil {
			return false
		}
		columns[i] = p.Column
	}
	p := DefaultIndexName(t, columns...)
	if idx.Name == p {
		return true
	}
	i, err := strconv.ParseInt(strings.TrimPrefix(idx.Name, p), 10, 64)
	return err == nil && i > 0
}

// IndexAttrChanged reports if the index attributes were changed.
// The default type is BTREE if no type was specified.
func (*diff) IndexAttrChanged(from, to []Attr) bool {
	t1 := &IndexType{T: IndexTypeBTree}
	if has(from, t1) {
		t1.T = strings.ToUpper(t1.T)
	}
	t2 := &IndexType{T: IndexTypeBTree}
	if has(to, t2) {
		t2.T = strings.ToUpper(t2.T)
	}
	if t1.T != t2.T {
		return true
	}
	var p1, p2 IndexPredicate
	if has(from, &p1) != has(to, &p2) || (p1.Predicate != p2.Predicate && p1.Predicate != sqlutil.MayWrap(p2.Predicate)) {
		return true
	}
	if indexIncludeChanged(from, to) {
		return true
	}
	s1, ok1 := indexStorageParams(from)
	s2, ok2 := indexStorageParams(to)
	return ok1 != ok2 || ok1 && *s1 != *s2
}

// IndexPartAttrChanged reports if the index-part attributes were changed.
func (*diff) IndexPartAttrChanged(from, to *IndexPart) bool {
	p1 := &IndexColumnProperty{NullsFirst: from.Descending, NullsLast: !from.Descending}
	has(from.Attrs, p1)
	p2 := &IndexColumnProperty{NullsFirst: to.Descending, NullsLast: !to.Descending}
	has(to.Attrs, p2)
	return p1.NullsFirst != p2.NullsFirst || p1.NullsLast != p2.NullsLast
}

// ReferenceChanged reports if the foreign key referential action was changed.
func (*diff) ReferenceChanged(from, to string) bool {
	// According to PostgreSQL, the NO ACTION rule is set
	// if no referential action was defined in foreign key.
	if from == "" {
		from = NoAction
	}
	if to == "" {
		to = NoAction
	}
	return from != to
}

func (d *diff) typeChanged(from, to *Column) (bool, error) {
	fromT, toT := from.Type.Type, to.Type.Type
	if fromT == nil || toT == nil {
		return false, fmt.Errorf("postgres: missing type information for column %q", from.Name)
	}
	if reflect.TypeOf(fromT) != reflect.TypeOf(toT) {
		return true, nil
	}
	var changed bool
	switch fromT := fromT.(type) {
	case *BinaryType, *BitType, *BoolType, *DecimalType, *FloatType,
		*IntervalType, *IntegerType, *JSONType, *SerialType, *SpatialType,
		*StringType, *TimeType, *NetworkType, *UserDefinedType:
		t1, err := FormatType(toT)
		if err != nil {
			return false, err
		}
		t2, err := FormatType(fromT)
		if err != nil {
			return false, err
		}
		changed = t1 != t2
	case *EnumType:
		toT := toT.(*EnumType)
		// Column type was changed if the underlying enum type was changed or values are not equal.
		changed = !sqlutil.ValuesEqual(fromT.Values, toT.Values) || fromT.T != toT.T ||
			(toT.Schema != nil && fromT.Schema != nil && fromT.Schema.Name != toT.Schema.Name)
	case *CurrencyType:
		toT := toT.(*CurrencyType)
		changed = fromT.T != toT.T
	case *UUIDType:
		toT := toT.(*UUIDType)
		changed = fromT.T != toT.T
	case *XMLType:
		toT := toT.(*XMLType)
		changed = fromT.T != toT.T
	case *ArrayType:
		toT := toT.(*ArrayType)
		if changed = fromT.T != toT.T; !changed {
			fromE, ok1 := fromT.Type.(*EnumType)
			toE, ok2 := toT.Type.(*EnumType)
			changed = ok1 && ok2 && !sqlutil.ValuesEqual(fromE.Values, toE.Values)
			break
		}
		// In case the desired schema is not normalized, the string type can look different even
		// if the two strings represent the same array type (varchar(1), character varying (1)).
		// Therefore, we try by comparing the underlying types if they were defined.
		if fromT.Type != nil && toT.Type != nil {
			t1, err := FormatType(fromT.Type)
			if err != nil {
				return false, err
			}
			t2, err := FormatType(toT.Type)
			if err != nil {
				return false, err
			}
			// Same underlying type.
			changed = t1 != t2
		}
	default:
		return false, &UnsupportedTypeError{Type: fromT}
	}
	return changed, nil
}

// Default IDENTITY attributes.
const (
	defaultIdentityGen  = "BY DEFAULT"
	defaultSeqStart     = 1
	defaultSeqIncrement = 1
)

// identityChanged reports if one of the identity attributes was changed.
func identityChanged(from, to []Attr) bool {
	i1, ok1 := identity(from)
	i2, ok2 := identity(to)
	if !ok1 && !ok2 || ok1 != ok2 {
		return ok1 != ok2
	}
	return i1.Generation != i2.Generation || i1.Sequence.Start != i2.Sequence.Start || i1.Sequence.Increment != i2.Sequence.Increment
}

func identity(attrs []Attr) (*Identity, bool) {
	i := &Identity{}
	if !has(attrs, i) {
		return nil, false
	}
	if i.Generation == "" {
		i.Generation = defaultIdentityGen
	}
	if i.Sequence == nil {
		i.Sequence = &Sequence{Start: defaultSeqStart, Increment: defaultSeqIncrement}
		return i, true
	}
	if i.Sequence.Start == 0 {
		i.Sequence.Start = defaultSeqStart
	}
	if i.Sequence.Increment == 0 {
		i.Sequence.Increment = defaultSeqIncrement
	}
	return i, true
}

// formatPartition returns the string representation of the
// partition key according to the PostgreSQL format/grammar.
func formatPartition(p Partition) (string, error) {
	b := &sqlutil.Builder{QuoteChar: '"'}
	b.P("PARTITION BY")
	switch t := strings.ToUpper(p.T); t {
	case PartitionTypeRange, PartitionTypeList, PartitionTypeHash:
		b.P(t)
	default:
		return "", fmt.Errorf("unknown partition type: %q", t)
	}
	if len(p.Parts) == 0 {
		return "", errors.New("missing parts for partition key")
	}
	b.Wrap(func(b *sqlutil.Builder) {
		b.MapComma(p.Parts, func(i int, b *sqlutil.Builder) {
			switch k := p.Parts[i]; {
			case k.Column != "":
				b.Ident(k.Column)
			case k.Expr != nil:
				b.P(sqlutil.MayWrap(k.Expr.(*RawExpr).Expr))
			}
		})
	})
	return b.String(), nil
}

// indexStorageParams returns the index storage parameters from the attributes
// in case it is there, and it is not the default.
func indexStorageParams(attrs []Attr) (*IndexStorageParams, bool) {
	s := &IndexStorageParams{}
	if !has(attrs, s) {
		return nil, false
	}
	if !s.AutoSummarize && (s.PagesPerRange == 0 || s.PagesPerRange == DefaultPagePerRange) {
		return nil, false
	}
	return s, true
}

// indexIncludeChanged reports if the INCLUDE attribute clause was changed.
func indexIncludeChanged(from, to []Attr) bool {
	var fromI, toI IndexInclude
	if has(from, &fromI) != has(to, &toI) || len(fromI.Columns) != len(toI.Columns) {
		return true
	}
	for i := range fromI.Columns {
		if fromI.Columns[i] != toI.Columns[i] {
			return true
		}
	}
	return false
}
