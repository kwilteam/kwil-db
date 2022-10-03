package postgres

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"kwil/x/schemadef/hcl"
	"kwil/x/schemadef/hclspec"
	"kwil/x/schemadef/schema"

	"kwil/x/sql/catalog"
	"kwil/x/sql/core"
	"kwil/x/sql/engine"
	sqlx "kwil/x/sql/x"

	"github.com/hashicorp/hcl/v2/hclparse"
)

// evalSpec evaluates a DDL document into v using the input.
func evalSpec(p *hclparse.Parser, v any, input map[string]string) error {
	var d hclspec.Realm
	if err := hclState.Eval(p, &d, input); err != nil {
		return err
	}

	c := NewCatalog()
	up := catalog.NewUpdater(c)
	e := engine.NewEngine(NewParser(), c)
	conv := &specConverter{}

	switch v := v.(type) {
	case *schema.Realm:
		if err := hclspec.Scan(v, &d, conv); err != nil {
			return fmt.Errorf("spec: failed converting to *schema.Realm: %w", err)
		}
		for _, s := range v.Schemas {
			if err := up.UpdateSchema(s, &catalogConverter{}); err != nil {
				return fmt.Errorf("spec: failed adding schema %q: %w", s.Name, err)
			}
		}
		if err := validateQueries(e, v); err != nil {
			return err
		}

	case *schema.Schema:
		if len(d.Schemas) != 1 {
			return fmt.Errorf("spec: expecting document to contain a single schema, got %d", len(d.Schemas))
		}
		r := &schema.Realm{}
		if err := hclspec.Scan(r, &d, conv); err != nil {
			return err
		}
		*v = *r.Schemas[0]
		if err := up.UpdateSchema(v, &catalogConverter{}); err != nil {
			return fmt.Errorf("spec: failed adding schema %q: %w", v.Name, err)
		}
		if err := validateQueries(e, r); err != nil {
			return err
		}

	default:
		return fmt.Errorf("spec: failed unmarshaling spec. %T is not supported", v)
	}

	return nil
}

func validateQueries(e *engine.Engine, r *schema.Realm) error {
	for _, q := range r.Queries {
		_, err := e.ParseStatement(q.Statement)
		if err != nil {
			if !errors.Is(err, core.ErrUnsupportedOS) {
				return fmt.Errorf("spec: failed parsing query %q: %w", q.Name, err)
			}
		}
	}
	return nil
}

// MarshalSpec marshals v into a DDL document using a hcl.Marshaler.
func MarshalSpec(v any, marshaler hcl.Marshaler) ([]byte, error) {
	var d hclspec.Realm
	switch v := v.(type) {
	case *schema.Schema:
		r := &schema.Realm{Schemas: []*schema.Schema{v}}

		var err error
		doc, err := realmSpec(r)
		if err != nil {
			return nil, fmt.Errorf("spec: failed converting schema to spec: %w", err)
		}
		d = *doc
	case *schema.Realm:
		doc, err := realmSpec(v)
		if err != nil {
			return nil, fmt.Errorf("spec: failed converting realm to spec: %w", err)
		}
		d = *doc

		if err := hclspec.QualifyDuplicates(d.Tables); err != nil {
			return nil, err
		}
		if err := hclspec.QualifyReferences(d.Tables, v); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("spec: failed marshaling spec. %T is not supported", v)
	}
	return marshaler.MarshalSpec(&d)
}

var (
	hclState = hcl.New(
		hcl.WithTypes(TypeRegistry.Specs()),
		hcl.WithScopedEnums("table.index.type", IndexTypeBTree, IndexTypeHash, IndexTypeGIN, IndexTypeGiST, IndexTypeBRIN),
		hcl.WithScopedEnums("table.partition.type", PartitionTypeRange, PartitionTypeList, PartitionTypeHash),
		hcl.WithScopedEnums("table.column.identity.generated", GeneratedTypeAlways, GeneratedTypeByDefault),
		hcl.WithScopedEnums("table.column.as.type", "STORED"),
		hcl.WithScopedEnums("table.foreign_key.on_update", hclspec.ReferenceVars...),
		hcl.WithScopedEnums("table.foreign_key.on_delete", hclspec.ReferenceVars...),
	)
	// MarshalHCL marshals v into an HCL DDL document.
	MarshalHCL = hcl.MarshalerFunc(func(v any) ([]byte, error) {
		return MarshalSpec(v, hclState)
	})
	// EvalHCL implements the hcl.Evaluator interface.
	EvalHCL = hcl.EvalFunc(evalSpec)

	// EvalHCLBytes is a helper that evaluates an HCL document from a byte slice instead
	// of from an hclparse.Parser instance.
	EvalHCLBytes = hclspec.HCLBytesFunc(EvalHCL)
)

type specConverter struct{}

// ConvertRole converts a hclspec.Role to a schema.Role.
func (sc *specConverter) ConvertRole(r *hclspec.Role, db *schema.Realm) (*schema.Role, error) {
	return hclspec.ToRole(r, db)
}

// ConvertTable converts a spec.Table to a schema.Table. Table conversion is done without converting
// ForeignKeySpecs into ForeignKeys, as the target tables do not necessarily exist in the schema
// at this point. Instead, the linking is done by the convertSchema function.
func (sc *specConverter) ConvertTable(tab *hclspec.Table, parent *schema.Schema) (*schema.Table, error) {
	t := &schema.Table{
		Name:   tab.Name,
		Schema: parent,
	}
	for _, csp := range tab.Columns {
		col, err := sc.convertColumn(csp, t)
		if err != nil {
			return nil, err
		}
		t.Columns = append(t.Columns, col)
	}
	if tab.PrimaryKey != nil {
		pk, err := sc.convertPrimaryKey(tab.PrimaryKey, t)
		if err != nil {
			return nil, err
		}
		t.PrimaryKey = pk
	}

	for _, idx := range tab.Indexes {
		i, err := sc.convertIndex(idx, t)
		if err != nil {
			return nil, err
		}
		t.Indexes = append(t.Indexes, i)
	}
	for _, c := range tab.Checks {
		c, err := sc.convertCheck(c)
		if err != nil {
			return nil, err
		}
		t.AddChecks(c)
	}
	if err := hclspec.ConvertCommentFromSpec(tab, &t.Attrs); err != nil {
		return nil, err
	}
	if err := sc.convertPartition(tab.Extra, t); err != nil {
		return nil, err
	}
	return t, nil
}

// ConvertQuery converts a spec.Query to a schema.Query.
func (sc *specConverter) ConvertQuery(q *hclspec.Query, r *schema.Realm) (*schema.Query, error) {
	return hclspec.ToQuery(q, r)
}

// ConvertType converts a spec.Type to a schema.Type.
func (sc *specConverter) ConvertType(htyp *hcl.Type, attrs ...*hcl.Attr) (schema.Type, error) {
	typ, err := TypeRegistry.Type(htyp, attrs)
	if err != nil {
		return nil, err
	}

	// Handle default values for time precision types.
	if t, ok := typ.(*schema.TimeType); ok && strings.HasPrefix(t.T, "time") {
		if _, ok := attr(htyp, "precision"); !ok {
			p := defaultTimePrecision
			t.Precision = &p
		}
	}
	return typ, nil
}

// ConvertEnums converts possibly referenced column types (like enums) to
// an actual schema.Type and sets it on the correct schema.Column.
func (sc *specConverter) ConvertEnums(tables []*hclspec.Table, enums []*hclspec.Enum, r *schema.Realm) error {
	var (
		used   = make(map[*hclspec.Enum]struct{})
		byName = make(map[string]*hclspec.Enum)
	)
	for _, e := range enums {
		byName[e.Name] = e
	}
	for _, t := range tables {
		for _, c := range t.Columns {
			var enum *hclspec.Enum
			switch {
			case c.Type.IsRef:
				n, err := enumName(c.Type)
				if err != nil {
					return err
				}
				e, ok := byName[n]
				if !ok {
					return fmt.Errorf("enum %q was not found", n)
				}
				enum = e
			default:
				n, ok := arrayType(c.Type.T)
				if !ok || byName[n] == nil {
					continue
				}
				enum = byName[n]
			}
			used[enum] = struct{}{}
			schemaE, err := hclspec.SchemaName(enum.Schema)
			if err != nil {
				return fmt.Errorf("extract schema name from enum refrence: %w", err)
			}
			es, ok := r.Schema(schemaE)
			if !ok {
				return fmt.Errorf("schema %q not found in realm for table %q", schemaE, t.Name)
			}
			schemaT, err := hclspec.SchemaName(t.Schema)
			if err != nil {
				return fmt.Errorf("extract schema name from table refrence: %w", err)
			}
			ts, ok := r.Schema(schemaT)
			if !ok {
				return fmt.Errorf("schema %q not found in realm for table %q", schemaT, t.Name)
			}
			tt, ok := ts.Table(t.Name)
			if !ok {
				return fmt.Errorf("table %q not found in schema %q", t.Name, ts.Name)
			}
			cc, ok := tt.Column(c.Name)
			if !ok {
				return fmt.Errorf("column %q not found in table %q", c.Name, t.Name)
			}
			e := &schema.EnumType{T: enum.Name, Schema: es, Values: enum.Values}
			switch t := cc.Type.Type.(type) {
			case *ArrayType:
				t.Type = e
			default:
				cc.Type.Type = e
			}
		}
	}
	for _, e := range enums {
		if _, ok := used[e]; !ok {
			return fmt.Errorf("enum %q declared but not used", e.Name)
		}
	}
	return nil
}

// convertColumn converts a spec.Column into a schema.Column.
func (sc *specConverter) convertColumn(s *hclspec.Column, _ *schema.Table) (*schema.Column, error) {
	if err := fixDefaultQuotes(s.Default); err != nil {
		return nil, err
	}
	c, err := hclspec.ToColumn(s, sc)
	if err != nil {
		return nil, err
	}
	if r, ok := s.Extra.Resource("identity"); ok {
		id, err := sc.convertIdentity(r)
		if err != nil {
			return nil, err
		}
		c.Attrs = append(c.Attrs, id)
	}
	if err := hclspec.ConvertGenExpr(s.Remain(), c, generatedType); err != nil {
		return nil, err
	}
	return c, nil
}

// convertPrimaryKey converts a hclspec.PrimaryKey into a schema.PrimaryKey.
func (sc *specConverter) convertPrimaryKey(s *hclspec.PrimaryKey, parent *schema.Table) (*schema.Index, error) {
	parts := make([]*schema.IndexPart, 0, len(s.Columns))
	for seqno, c := range s.Columns {
		c, err := hclspec.ColumnByRef(parent, c)
		if err != nil {
			return nil, nil
		}
		parts = append(parts, &schema.IndexPart{
			Seq:    seqno,
			Column: c,
		})
	}
	return &schema.Index{
		Table: parent,
		Parts: parts,
	}, nil
}

// convertIndex converts a hclspec.Index into a schema.Index.
func (sc *specConverter) convertIndex(s *hclspec.Index, t *schema.Table) (*schema.Index, error) {
	idx, err := hclspec.ToIndex(s, t)
	if err != nil {
		return nil, err
	}
	if attr, ok := s.Attr("type"); ok {
		t, err := attr.String()
		if err != nil {
			return nil, err
		}
		idx.Attrs = append(idx.Attrs, &IndexType{T: t})
	}
	if attr, ok := s.Attr("where"); ok {
		p, err := attr.String()
		if err != nil {
			return nil, err
		}
		idx.Attrs = append(idx.Attrs, &IndexPredicate{Predicate: p})
	}
	if attr, ok := s.Attr("page_per_range"); ok {
		p, err := attr.Int64()
		if err != nil {
			return nil, err
		}
		idx.Attrs = append(idx.Attrs, &IndexStorageParams{PagesPerRange: p})
	}
	if attr, ok := s.Attr("include"); ok {
		refs, err := attr.Refs()
		if err != nil {
			return nil, err
		}
		if len(refs) == 0 {
			return nil, fmt.Errorf("unexpected empty INCLUDE in index %q definition", s.Name)
		}
		include := make([]string, len(refs))
		for i, r := range refs {
			col, err := hclspec.ColumnByRef(t, r)
			if err != nil {
				return nil, err
			}
			include[i] = col.Name
		}
		idx.Attrs = append(idx.Attrs, &IndexInclude{Columns: include})
	}
	return idx, nil
}

// convertCheck converts a hclspec.Check into a schema.Check.
func (sc *specConverter) convertCheck(c *hclspec.Check) (*schema.Check, error) {
	return &schema.Check{
		Name: c.Name,
		Expr: c.Expr,
	}, nil
}

// convertPartition converts and appends the partition block into the table attributes if exists.
func (sc *specConverter) convertPartition(s hcl.Resource, table *schema.Table) error {
	r, ok := s.Resource("partition")
	if !ok {
		return nil
	}
	var p struct {
		Type    string     `spec:"type"`
		Columns []*hcl.Ref `spec:"columns"`
		Parts   []*struct {
			Expr   string   `spec:"expr"`
			Column *hcl.Ref `spec:"column"`
		} `spec:"by"`
	}
	if err := r.As(&p); err != nil {
		return fmt.Errorf("parsing %s.partition: %w", table.Name, err)
	}
	if p.Type == "" {
		return fmt.Errorf("missing attribute %s.partition.type", table.Name)
	}
	key := &Partition{T: p.Type}
	switch n, m := len(p.Columns), len(p.Parts); {
	case n == 0 && m == 0:
		return fmt.Errorf("missing columns or expressions for %s.partition", table.Name)
	case n > 0 && m > 0:
		return fmt.Errorf(`multiple definitions for %s.partition, use "columns" or "by"`, table.Name)
	case n > 0:
		for _, r := range p.Columns {
			c, err := hclspec.ColumnByRef(table, r)
			if err != nil {
				return err
			}
			key.Parts = append(key.Parts, &PartitionPart{Column: c.Name})
		}
	case m > 0:
		for i, p := range p.Parts {
			switch {
			case p.Column == nil && p.Expr == "":
				return fmt.Errorf("missing column or expression for %s.partition.by at position %d", table.Name, i)
			case p.Column != nil && p.Expr != "":
				return fmt.Errorf("multiple definitions for  %s.partition.by at position %d", table.Name, i)
			case p.Column != nil:
				c, err := hclspec.ColumnByRef(table, p.Column)
				if err != nil {
					return err
				}
				key.Parts = append(key.Parts, &PartitionPart{Column: c.Name})
			case p.Expr != "":
				key.Parts = append(key.Parts, &PartitionPart{Expr: &schema.RawExpr{X: p.Expr}})
			}
		}
	}
	table.AddAttrs(key)
	return nil
}

func (sc *specConverter) convertIdentity(r *hcl.Resource) (*Identity, error) {
	var s struct {
		Generation string `spec:"generated"`
		Start      int64  `spec:"start"`
		Increment  int64  `spec:"increment"`
	}
	if err := r.As(&s); err != nil {
		return nil, err
	}
	id := &Identity{Generation: hclspec.FromVar(s.Generation), Sequence: &Sequence{}}
	if s.Start != 0 {
		id.Sequence.Start = s.Start
	}
	if s.Increment != 0 {
		id.Sequence.Increment = s.Increment
	}
	return id, nil
}

// fixDefaultQuotes fixes the quotes on the Default field to be single quotes
// instead of double quotes.
func fixDefaultQuotes(value hcl.Value) error {
	lv, ok := value.(*hcl.LiteralValue)
	if !ok {
		return nil
	}
	if sqlx.IsQuoted(lv.V, '"') {
		uq, err := strconv.Unquote(lv.V)
		if err != nil {
			return err
		}
		lv.V = "'" + uq + "'"
	}
	return nil
}

const defaultTimePrecision = 6

// fromPartition returns the resource spec for representing the partition block.
func fromPartition(p Partition) *hcl.Resource {
	key := &hcl.Resource{
		Type: "partition",
		Attrs: []*hcl.Attr{
			hclspec.VarAttr("type", strings.ToUpper(hclspec.ToVar(p.T))),
		},
	}
	columns, ok := func() (*hcl.ListValue, bool) {
		parts := make([]hcl.Value, 0, len(p.Parts))
		for _, p := range p.Parts {
			if p.Column == "" {
				return nil, false
			}
			parts = append(parts, hclspec.ColumnRef(p.Column))
		}
		return &hcl.ListValue{V: parts}, true
	}()
	if ok {
		key.Attrs = append(key.Attrs, &hcl.Attr{K: "columns", V: columns})
		return key
	}
	for _, p := range p.Parts {
		part := &hcl.Resource{Type: "by"}
		switch {
		case p.Column != "":
			part.Attrs = append(part.Attrs, hclspec.RefAttr("column", hclspec.ColumnRef(p.Column)))
		case p.Expr != nil:
			part.Attrs = append(part.Attrs, hclspec.StrAttr("expr", p.Expr.(*schema.RawExpr).X))
		}
		key.Children = append(key.Children, part)
	}
	return key
}

// enumName extracts the name of the referenced Enum from the reference string.
func enumName(ref *hcl.Type) (string, error) {
	s := strings.Split(ref.T, "$enum.")
	if len(s) != 2 {
		return "", fmt.Errorf("postgres: failed to extract enum name from %q", ref.T)
	}
	return s[1], nil
}

// enumRef returns a reference string to the given enum name.
func enumRef(n string) *hcl.Ref {
	return &hcl.Ref{
		V: "$enum." + n,
	}
}

func realmSpec(r *schema.Realm) (*hclspec.Realm, error) {
	d := &hclspec.Realm{}

	for _, s := range r.Schemas {
		doc, err := schemaSpec(s)
		if err != nil {
			return nil, fmt.Errorf("spec: failed converting schema to spec: %w", err)
		}
		d.Tables = append(d.Tables, doc.Tables...)
		d.Schemas = append(d.Schemas, doc.Schemas...)
		d.Enums = append(d.Enums, doc.Enums...)
	}

	for _, q := range r.Queries {
		sq, err := querySpec(q)
		if err != nil {
			return nil, fmt.Errorf("spec: failed converting query to spec: %w", err)
		}
		d.Queries = append(d.Queries, sq)
	}

	for _, rol := range r.Roles {
		sr, err := hclspec.FromRole(rol)
		if err != nil {
			return nil, fmt.Errorf("spec: failed converting role to spec: %w", err)
		}
		d.Roles = append(d.Roles, sr)
	}

	return d, nil
}

// schemaSpec converts from a concrete Postgres schema to spec.
func schemaSpec(schem *schema.Schema) (*hclspec.Realm, error) {
	s, tbls, err := hclspec.FromSchema(schem, tableSpec)
	if err != nil {
		return nil, err
	}
	d := &hclspec.Realm{
		Tables:  tbls,
		Schemas: []*hclspec.Schema{s},
	}
	enums := make(map[string]bool)
	for _, t := range schem.Tables {
		for _, c := range t.Columns {
			if e, ok := hasEnumType(c); ok && !enums[e.T] {
				d.Enums = append(d.Enums, &hclspec.Enum{
					Name:   e.T,
					Schema: hclspec.SchemaRef(s.Name),
					Values: e.Values,
				})
				enums[e.T] = true
			}
		}
	}

	return d, nil
}

// tableSpec converts from a concrete Postgres spec.Table to a schema.Table.
func tableSpec(table *schema.Table) (*hclspec.Table, error) {
	spec, err := hclspec.FromTable(
		table,
		columnSpec,
		hclspec.FromPrimaryKey,
		indexSpec,
		hclspec.FromForeignKey,
		hclspec.FromCheck,
	)
	if err != nil {
		return nil, err
	}
	if p := (Partition{}); schema.Has(table.Attrs, &p) {
		spec.Extra.Children = append(spec.Extra.Children, fromPartition(p))
	}
	return spec, nil
}

func indexSpec(idx *schema.Index) (*hclspec.Index, error) {
	s, err := hclspec.FromIndex(idx)
	if err != nil {
		return nil, err
	}
	// Avoid printing the index type if it is the default.
	if i := (IndexType{}); schema.Has(idx.Attrs, &i) && i.T != IndexTypeBTree {
		s.Extra.Attrs = append(s.Extra.Attrs, hclspec.VarAttr("type", strings.ToUpper(i.T)))
	}
	if i := (IndexInclude{}); schema.Has(idx.Attrs, &i) && len(i.Columns) > 0 {
		attr := &hcl.ListValue{}
		for _, c := range i.Columns {
			attr.V = append(attr.V, hclspec.ColumnRef(c))
		}
		s.Extra.Attrs = append(s.Extra.Attrs, &hcl.Attr{
			K: "include",
			V: attr,
		})
	}
	if i := (IndexPredicate{}); schema.Has(idx.Attrs, &i) && i.Predicate != "" {
		s.Extra.Attrs = append(s.Extra.Attrs, hclspec.VarAttr("where", strconv.Quote(i.Predicate)))
	}
	if p, ok := indexStorageParams(idx.Attrs); ok {
		s.Extra.Attrs = append(s.Extra.Attrs, hclspec.Int64Attr("page_per_range", p.PagesPerRange))
	}
	return s, nil
}

// columnSpec converts from a concrete Postgres schema.Column into a spec.Column.
func columnSpec(c *schema.Column, _ *schema.Table) (*hclspec.Column, error) {
	s, err := hclspec.FromColumn(c, columnTypeSpec)
	if err != nil {
		return nil, err
	}
	if i := (&Identity{}); schema.Has(c.Attrs, i) {
		s.Extra.Children = append(s.Extra.Children, fromIdentity(i))
	}
	if x := (schema.GeneratedExpr{}); schema.Has(c.Attrs, &x) {
		s.Extra.Children = append(s.Extra.Children, hclspec.FromGenExpr(x, generatedType))
	}
	return s, nil
}

func querySpec(q *schema.Query) (*hclspec.Query, error) {
	s, err := hclspec.FromQuery(q)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// fromIdentity returns the resource spec for representing the identity attributes.
func fromIdentity(i *Identity) *hcl.Resource {
	id := &hcl.Resource{
		Type: "identity",
		Attrs: []*hcl.Attr{
			hclspec.VarAttr("generated", strings.ToUpper(hclspec.ToVar(i.Generation))),
		},
	}
	if s := i.Sequence; s != nil {
		if s.Start != 1 {
			id.Attrs = append(id.Attrs, hclspec.Int64Attr("start", s.Start))
		}
		if s.Increment != 1 {
			id.Attrs = append(id.Attrs, hclspec.Int64Attr("increment", s.Increment))
		}
	}
	return id
}

// columnTypeSpec converts from a concrete Postgres schema.Type into spec.Column Type.
func columnTypeSpec(t schema.Type) (*hclspec.Column, error) {
	// Handle postgres enum types. They cannot be put into the TypeRegistry since their name is dynamic.
	if e, ok := t.(*schema.EnumType); ok {
		return &hclspec.Column{Type: &hcl.Type{
			T:     enumRef(e.T).V,
			IsRef: true,
		}}, nil
	}
	st, err := TypeRegistry.Convert(t)
	if err != nil {
		return nil, err
	}
	return &hclspec.Column{Type: st}, nil
}

// TypeRegistry contains the supported TypeSpecs for the Postgres driver.
var TypeRegistry = hcl.NewRegistry(
	hcl.WithSpecFunc(typeSpec),
	hcl.WithParser(ParseType),
	hcl.WithSpecs(
		hcl.NewTypeSpec(TypeBit, hcl.WithAttributes(&hcl.TypeAttr{Name: "len", Kind: reflect.Int64})),
		hcl.AliasTypeSpec("bit_varying", TypeBitVar, hcl.WithAttributes(&hcl.TypeAttr{Name: "len", Kind: reflect.Int64})),
		hcl.NewTypeSpec(TypeVarChar, hcl.WithAttributes(hcl.SizeTypeAttr(false))),
		hcl.AliasTypeSpec("character_varying", TypeCharVar, hcl.WithAttributes(hcl.SizeTypeAttr(false))),
		hcl.NewTypeSpec(TypeChar, hcl.WithAttributes(hcl.SizeTypeAttr(false))),
		hcl.NewTypeSpec(TypeCharacter, hcl.WithAttributes(hcl.SizeTypeAttr(false))),
		hcl.NewTypeSpec(TypeInt2),
		hcl.NewTypeSpec(TypeInt4),
		hcl.NewTypeSpec(TypeInt8),
		hcl.NewTypeSpec(TypeInt),
		hcl.NewTypeSpec(TypeInteger),
		hcl.NewTypeSpec(TypeSmallInt),
		hcl.NewTypeSpec(TypeBigInt),
		hcl.NewTypeSpec(TypeText),
		hcl.NewTypeSpec(TypeBoolean),
		hcl.NewTypeSpec(TypeBool),
		hcl.NewTypeSpec(TypeBytea),
		hcl.NewTypeSpec(TypeCIDR),
		hcl.NewTypeSpec(TypeInet),
		hcl.NewTypeSpec(TypeMACAddr),
		hcl.NewTypeSpec(TypeMACAddr8),
		hcl.NewTypeSpec(TypeCircle),
		hcl.NewTypeSpec(TypeLine),
		hcl.NewTypeSpec(TypeLseg),
		hcl.NewTypeSpec(TypeBox),
		hcl.NewTypeSpec(TypePath),
		hcl.NewTypeSpec(TypePoint),
		hcl.NewTypeSpec(TypePolygon),
		hcl.NewTypeSpec(TypeDate),
		hcl.NewTypeSpec(TypeTime, hcl.WithAttributes(precisionTypeAttr()), formatTime()),
		hcl.NewTypeSpec(TypeTimeTZ, hcl.WithAttributes(precisionTypeAttr()), formatTime()),
		hcl.NewTypeSpec(TypeTimestampTZ, hcl.WithAttributes(precisionTypeAttr()), formatTime()),
		hcl.NewTypeSpec(TypeTimestamp, hcl.WithAttributes(precisionTypeAttr()), formatTime()),
		hcl.AliasTypeSpec("double_precision", TypeDouble),
		hcl.NewTypeSpec(TypeReal),
		hcl.NewTypeSpec(TypeFloat, hcl.WithAttributes(precisionTypeAttr())),
		hcl.NewTypeSpec(TypeFloat8),
		hcl.NewTypeSpec(TypeFloat4),
		hcl.NewTypeSpec(TypeNumeric, hcl.WithAttributes(precisionTypeAttr(), &hcl.TypeAttr{Name: "scale", Kind: reflect.Int, Required: false})),
		hcl.NewTypeSpec(TypeDecimal, hcl.WithAttributes(precisionTypeAttr(), &hcl.TypeAttr{Name: "scale", Kind: reflect.Int, Required: false})),
		hcl.NewTypeSpec(TypeSmallSerial),
		hcl.NewTypeSpec(TypeSerial),
		hcl.NewTypeSpec(TypeBigSerial),
		hcl.NewTypeSpec(TypeSerial2),
		hcl.NewTypeSpec(TypeSerial4),
		hcl.NewTypeSpec(TypeSerial8),
		hcl.NewTypeSpec(TypeXML),
		hcl.NewTypeSpec(TypeJSON),
		hcl.NewTypeSpec(TypeJSONB),
		hcl.NewTypeSpec(TypeUUID),
		hcl.NewTypeSpec(TypeMoney),
		hcl.NewTypeSpec("hstore"),
		hcl.NewTypeSpec("sql", hcl.WithAttributes(&hcl.TypeAttr{Name: "def", Required: true, Kind: reflect.String})),
	),
	hcl.WithSpecs(func() (specs []*hcl.TypeSpec) {
		opts := []hcl.TypeSpecOption{
			hcl.WithToSpec(func(t schema.Type) (*hcl.Type, error) {
				i, ok := t.(*IntervalType)
				if !ok {
					return nil, fmt.Errorf("postgres: unexpected interval type %T", t)
				}
				s := &hcl.Type{T: TypeInterval}
				if i.F != "" {
					s.T = hclspec.ToVar(strings.ToLower(i.F))
				}
				if p := i.Precision; p != nil && *p != defaultTimePrecision {
					s.Attrs = []*hcl.Attr{hclspec.IntAttr("precision", *p)}
				}
				return s, nil
			}),
			hcl.WithFromSpec(func(t *hcl.Type) (schema.Type, error) {
				i := &IntervalType{T: TypeInterval}
				if t.T != TypeInterval {
					i.F = hclspec.FromVar(t.T)
				}
				if a, ok := attr(t, "precision"); ok {
					p, err := a.Int()
					if err != nil {
						return nil, fmt.Errorf(`postgres: parsing attribute "precision": %w`, err)
					}
					if p != defaultTimePrecision {
						i.Precision = &p
					}
				}
				return i, nil
			}),
		}
		for _, f := range []string{"interval", "second", "day to second", "hour to second", "minute to second"} {
			specs = append(specs, hcl.NewTypeSpec(hclspec.ToVar(f), append(opts, hcl.WithAttributes(precisionTypeAttr()))...))
		}
		for _, f := range []string{"year", "month", "day", "hour", "minute", "year to month", "day to hour", "day to minute", "hour to minute"} {
			specs = append(specs, hcl.NewTypeSpec(hclspec.ToVar(f), opts...))
		}
		return specs
	}()...),
)
