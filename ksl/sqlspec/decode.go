package sqlspec

import (
	"fmt"
	"strings"

	"golang.org/x/exp/slices"
	"ksl"
	"ksl/kslspec"
	"ksl/kslsyntax/ast"
)

type sqldecoder struct {
	ctx    *ksl.Context
	schema *kslspec.DocumentSpec
	types  *kslspec.TypeRegistry
}

func newDecoder(schema *kslspec.DocumentSpec, types *kslspec.TypeRegistry) *sqldecoder {
	return &sqldecoder{
		schema: schema,
		types:  types,
		ctx:    &ksl.Context{},
	}
}

func (d *sqldecoder) decodeFileSet(file *kslspec.FileSet) (*Realm, ksl.Diagnostics) {
	var r Realm
	var diags ksl.Diagnostics
	for _, file := range file.Files {
		diags = append(diags, kslspec.Validate(file.Body, d.schema)...)
	}

	if diags.HasErrors() {
		return nil, diags
	}

	var enums ast.Blocks
	var queries ast.Blocks
	var roles ast.Blocks
	var tables ast.Blocks

	for _, file := range file.Files {
		for _, blk := range file.Body.Blocks {
			switch blk.GetType() {
			case "enum":
				enums = append(enums, blk)
			case "query":
				queries = append(queries, blk)
			case "role":
				roles = append(roles, blk)
			case "table":
				tables = append(tables, blk)
			}
		}
	}

	diags = append(diags, d.decodeEnums(&r, enums)...)
	diags = append(diags, d.decodeQueries(&r, queries)...)
	diags = append(diags, d.decodeRoles(&r, roles)...)
	diags = append(diags, d.decodeTables(&r, tables)...)

	return &r, diags
}

func (d *sqldecoder) decode(file *kslspec.File) (*Realm, ksl.Diagnostics) {
	var r Realm
	diags := kslspec.Validate(file.Body, d.schema)
	if diags.HasErrors() {
		return &r, diags
	}

	grouped := file.Body.Blocks.ByType()

	diags = append(diags, d.decodeEnums(&r, grouped["enum"])...)
	diags = append(diags, d.decodeQueries(&r, grouped["query"])...)
	diags = append(diags, d.decodeRoles(&r, grouped["role"])...)
	diags = append(diags, d.decodeTables(&r, grouped["table"])...)

	return &r, diags
}

func (d *sqldecoder) decodeEnums(r *Realm, blocks ast.Blocks) ksl.Diagnostics {
	var diags ksl.Diagnostics

	for _, block := range blocks {
		qualName := getQualifiedTypeName(block.GetName())
		schema := r.GetOrCreateSchema(qualName.Schema)
		enum := schema.GetOrCreateEnum(qualName.Name)
		enum.Values = block.GetEnumValues()
	}

	return diags
}

func (d *sqldecoder) decodeQueries(r *Realm, blocks ast.Blocks) ksl.Diagnostics {
	var diags ksl.Diagnostics

	for _, block := range blocks {
		attrs := block.GetAttributes().ByName()
		query := r.GetOrCreateQuery(block.GetName())
		if attr, ok := attrs["statement"]; ok && attr.Value != nil {
			diags = append(diags, kslspec.DecodeExpression(attr.Value, d.ctx, &query.Statement)...)
		}
	}
	return diags
}

func (d *sqldecoder) decodeRoles(r *Realm, blocks ast.Blocks) ksl.Diagnostics {
	var diags ksl.Diagnostics
	var defaultRole *ast.Block

	for _, block := range blocks {
		role := r.GetOrCreateRole(block.GetName())
		attrs := block.GetAllLabels().ByName()

		if attr, ok := attrs["default"]; ok {
			switch attr.Value {
			case nil:
				role.Default = true
			default:
				diags = append(diags, kslspec.DecodeExpression(attr.Value, d.ctx, &role.Default)...)
			}

			if role.Default {
				if defaultRole != nil {
					diags = append(diags, &ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  DiagDuplicateDefaultRole,
						Detail:   fmt.Sprintf("Default role can only be assigned to a single role, and role %q is already the default role.", defaultRole.GetName()),
						Subject:  block.Range().Ptr(),
					})
				} else {
					defaultRole = block
				}
			}
		}

		if attr, ok := attrs["allow"]; ok {
			listValue := attr.GetValue().(*ast.List)

			for _, v := range listValue.GetValues() {
				var ref string
				diags = append(diags, kslspec.DecodeExpression(v, d.ctx, &ref)...)

				if q, ok := r.Query(ref); ok {
					role.Queries = append(role.Queries, q)
				} else {
					diags = append(diags, &ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  DiagUnknownQueryReference,
						Detail:   fmt.Sprintf("No query exists with name %q.", v),
						Subject:  v.Range().Ptr(),
					})
				}
			}
		}
	}
	if defaultRole != nil {
		r.DefaultRole, _ = r.Role(defaultRole.GetName())
	}

	return diags
}

func (d *sqldecoder) decodeTables(r *Realm, blocks ast.Blocks) ksl.Diagnostics {
	var diags ksl.Diagnostics

	for _, block := range blocks {
		qualName := getQualifiedTypeName(block.GetName())
		schema := r.GetOrCreateSchema(qualName.Schema)
		table := schema.GetOrCreateTable(qualName.Name)
		for _, def := range block.GetDefinitions() {
			table.GetOrCreateColumn(def.GetName())
		}
	}

	for _, block := range blocks {
		qualName := getQualifiedTypeName(block.GetName())
		schema, _ := r.Schema(qualName.Schema)
		table, _ := schema.Table(qualName.Name)

		for _, def := range block.GetDefinitions() {
			column, _ := table.Column(def.GetName())
			diags = append(diags, d.decodeColumn(r, column, def)...)
		}

		for _, annot := range block.GetAnnotations() {
			switch annot.GetName() {
			case "id":
				diags = append(diags, d.decodePrimaryKeyBlockAnnotation(r, table, annot)...)
			case "index":
				diags = append(diags, d.decodeIndexBlockAnnotation(r, table, annot)...)
			case "foreign_key":
				diags = append(diags, d.decodeForeignKeyBlockAnnotation(r, table, annot)...)
			}
		}
	}

	return diags
}

func (d *sqldecoder) decodeIndexBlockAnnotation(r *Realm, table *Table, annot *ast.Annotation) ksl.Diagnostics {
	var diags ksl.Diagnostics

	arg := annot.MustKwarg("columns").(*ast.List)

	columns := make([]*Column, 0, len(arg.Values))
	for _, v := range arg.Values {
		var ref string
		diags = append(diags, kslspec.DecodeExpression(v, d.ctx, &ref)...)
		qualRef := qualifiedColumnName(ref).Normalize(table.Schema.Name, table.Name)
		if qualRef.Schema != table.Schema.Name || qualRef.Table != table.Name {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidColumnReference,
				Detail:   fmt.Sprintf("Columns in an index must be in the same table as the index, but column %q is in %s.%s.", ref, qualRef.Schema, qualRef.Table),
				Subject:  v.Range().Ptr(),
			})
			continue
		}

		col, ok := table.Column(qualRef.Column)
		if !ok {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidColumnReference,
				Detail:   fmt.Sprintf("No column exists with name %q.", qualRef.Column),
				Subject:  v.Range().Ptr(),
			})
			continue
		}
		columns = append(columns, col)
	}

	var indexName string

	if arg, ok := annot.Arg(0); ok {
		diags = append(diags, kslspec.DecodeExpression(arg, d.ctx, &indexName)...)
	} else {
		parts := []string{table.Name}
		for _, col := range columns {
			parts = append(parts, col.Name)
		}
		parts = append(parts, "idx")
		indexName = strings.Join(parts, "_")
	}

	idx := table.GetOrCreateIndex(indexName).AddColumns(columns...)
	if arg, ok := annot.Kwarg("unique"); ok {
		diags = append(diags, kslspec.DecodeExpression(arg, d.ctx, &idx.Unique)...)
	}

	if arg, ok := annot.Kwarg("type"); ok {
		typ := &IndexType{}
		diags = append(diags, kslspec.DecodeExpression(arg, d.ctx, &typ.T)...)
		idx.AddAttrs(typ)
	}

	return diags
}

func (d *sqldecoder) decodeForeignKeyBlockAnnotation(r *Realm, table *Table, annot *ast.Annotation) ksl.Diagnostics {
	var diags ksl.Diagnostics

	columnsArg := annot.MustKwarg("columns").(*ast.List)
	refColumnsArg := annot.MustKwarg("references").(*ast.List)

	columns := make([]*Column, 0, len(columnsArg.Values))
	for _, v := range columnsArg.Values {
		var ref string
		diags = append(diags, kslspec.DecodeExpression(v, d.ctx, &ref)...)

		qualRef := qualifiedColumnName(ref).Normalize(table.Schema.Name, table.Name)
		if qualRef.Schema != table.Schema.Name || qualRef.Table != table.Name {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidColumnReference,
				Detail:   fmt.Sprintf("Columns in a foreign key must be in the same table as the foreign key, but column %q is in %s.%s.", ref, qualRef.Schema, qualRef.Table),
				Subject:  v.Range().Ptr(),
			})
			continue
		}
		col, ok := table.Column(qualRef.Column)
		if !ok {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidColumnReference,
				Detail:   fmt.Sprintf("No column exists with name %q.", qualRef.Column),
				Subject:  v.Range().Ptr(),
			})
			continue
		}
		columns = append(columns, col)
	}

	refColumns := make([]*Column, 0, len(refColumnsArg.Values))
	var refTable *Table

	for _, v := range refColumnsArg.Values {
		var ref string
		diags = append(diags, kslspec.DecodeExpression(v, d.ctx, &ref)...)

		qualRef := qualifiedColumnName(ref).Normalize(table.Schema.Name, table.Name)
		if refTable == nil {
			rt, ok := r.Table(qualRef.Schema, qualRef.Table)
			if !ok {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagUnknownTableReference,
					Detail:   fmt.Sprintf("No table exists with name %s.%s.", qualRef.Schema, qualRef.Table),
					Subject:  v.Range().Ptr(),
				})
				continue
			}
			refTable = rt
		} else if qualRef.Schema != refTable.Schema.Name || qualRef.Table != refTable.Name {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidColumnReference,
				Detail:   "References in a foreign key must all be in the same table.",
				Subject:  v.Range().Ptr(),
			})
			continue
		}

		col, ok := refTable.Column(qualRef.Column)
		if !ok {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidColumnReference,
				Detail:   fmt.Sprintf("No column exists with name %q.", qualRef.Column),
				Subject:  v.Range().Ptr(),
			})
			continue
		}
		refColumns = append(refColumns, col)
	}

	var name string

	if arg, ok := annot.Arg(0); ok {
		diags = append(diags, kslspec.DecodeExpression(arg, d.ctx, &name)...)
	} else {
		name = DefaultForeignKeyName(table, columns...)
	}

	fk := NewForeignKey(name).AddColumns(columns...).SetRefTable(refTable).AddRefColumns(refColumns...)
	if e, ok := annot.Kwarg("on_delete"); ok {
		diags = append(diags, kslspec.DecodeExpression(e, d.ctx, &fk.OnDelete)...)
	}
	if e, ok := annot.Kwarg("on_update"); ok {
		diags = append(diags, kslspec.DecodeExpression(e, d.ctx, &fk.OnUpdate)...)
	}
	table.AddForeignKeys(fk)
	return diags
}

func (d *sqldecoder) decodePrimaryKeyBlockAnnotation(r *Realm, table *Table, annot *ast.Annotation) ksl.Diagnostics {
	var diags ksl.Diagnostics

	if table.PrimaryKey != nil {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagDuplicatePrimaryKey,
			Detail:   "A table can only have one primary key.",
			Subject:  annot.Range().Ptr(),
		})
		return diags
	}

	arg := annot.MustArg(0).(*ast.List)

	columns := make([]*Column, 0, len(arg.Values))
	for _, v := range arg.Values {
		var ref string
		diags = append(diags, kslspec.DecodeExpression(v, d.ctx, &ref)...)

		qualRef := qualifiedColumnName(ref).Normalize(table.Schema.Name, table.Name)
		if qualRef.Schema != table.Schema.Name || qualRef.Table != table.Name {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidColumnReference,
				Detail:   fmt.Sprintf("Columns in a primary key must be in the same table as the primary key, but column %q is in %s.%s.", ref, qualRef.Schema, qualRef.Table),
				Subject:  v.Range().Ptr(),
			})
			continue
		}
		col, ok := table.Column(qualRef.Column)
		if !ok {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidColumnReference,
				Detail:   fmt.Sprintf("No column exists with name %q.", qualRef.Column),
				Subject:  v.Range().Ptr(),
			})
			continue
		}
		columns = append(columns, col)
	}

	table.SetPrimaryKey(NewPrimaryKey(columns...).SetName(DefaultPrimaryKeyName(table)))
	return diags
}

func (d *sqldecoder) decodeColumn(r *Realm, c *Column, def *ast.Definition) ksl.Diagnostics {
	var diags ksl.Diagnostics
	typeName := def.GetTypeName()
	columnType := &ColumnType{Nullable: def.IsNullable()}

	if spec, hasSpec := d.types.FindSpecNamed(typeName); hasSpec {
		typeName = spec.Type
		ct := &kslspec.ConcreteType{Type: typeName}

		for name := range spec.Annotations {
			if annot, ok := def.Annotation(name); ok {
				lit := &kslspec.LiteralValue{}
				diags = append(diags, kslspec.DecodeExpression(annot.MustArg(0), d.ctx, &lit.Value)...)
				ct.AddAttr(name, lit)
			}
		}
		typ, err := ConvertType(ct)
		if err != nil {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidColumnType,
				Detail:   fmt.Sprintf("Column %q has invalid type %q.", def.GetName(), typeName),
				Subject:  def.Range().Ptr(),
			})
			return diags
		}

		if def.IsArray() {
			columnType.Type = &ArrayType{Type: typ, T: typeName + "[]"}
			columnType.Raw = typeName + "[]"
		} else {
			columnType.Type = typ
			columnType.Raw, err = FormatType(typ)
		}

		if err != nil {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidColumnType,
				Detail:   fmt.Sprintf("Column %q has invalid type %q.", def.GetName(), typeName),
				Subject:  def.Range().Ptr(),
			})
			return diags
		}
	} else {
		qualName := getQualifiedTypeName(typeName)
		e, ok := r.Enum(qualName.Schema, qualName.Name)
		if !ok {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnknownTypeReference,
				Detail:   fmt.Sprintf("No type exists with name %q.", typeName),
				Subject:  def.Range().Ptr(),
			})
			return diags
		}
		columnType.Type = &EnumType{T: e.Name, Values: e.Values[:], Schema: e.Schema}
		columnType.Raw = typeName
	}

	for _, annot := range def.GetAnnotations() {
		switch annot.GetName() {
		case "id":
			if c.Table.PrimaryKey != nil {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagDuplicatePrimaryKey,
					Detail:   "A table can only have one primary key.",
					Subject:  annot.Range().Ptr(),
				})
			} else {

				c.Table.PrimaryKey = NewPrimaryKey(c).SetName(DefaultPrimaryKeyName(c.Table)).SetTable(c.Table)
			}
		case "default":
			defaultValue := &LiteralExpr{}
			diags = append(diags, kslspec.DecodeExpression(annot.MustArg(0), d.ctx, &defaultValue.Value)...)
			c.Default = defaultValue

			if e, ok := columnType.Type.(*EnumType); ok {
				if !slices.Contains(e.Values, defaultValue.Value) {
					diags = append(diags, &ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  DiagUnknownTypeReference,
						Detail:   fmt.Sprintf("Enum %q does not contain value %q.", e.T, defaultValue.Value),
						Subject:  annot.Range().Ptr(),
					})
				}
			}
		case "unique":
			diags = append(diags, d.decodeUniqueColumn(r, c, def, annot)...)
		case "foreign_key":
			diags = append(diags, d.decodeColumnForeignKey(r, c, def, annot)...)
		case "index":
			diags = append(diags, d.decodeColumnIndex(r, c, def, annot)...)
		}
	}

	c.Type = columnType

	return diags
}

func (d *sqldecoder) decodeColumnIndex(r *Realm, c *Column, def *ast.Definition, annot *ast.Annotation) ksl.Diagnostics {
	var diags ksl.Diagnostics
	var indexName string
	if arg, ok := annot.Arg(0); ok {
		diags = append(diags, kslspec.DecodeExpression(arg, d.ctx, &indexName)...)
	} else {
		indexName = DefaultIndexName(c.Table, c)
	}

	index := c.Table.GetOrCreateIndex(indexName)

	if typ, ok := annot.Kwarg("type"); ok {
		typePart := &IndexType{}
		diags = append(diags, kslspec.DecodeExpression(typ, d.ctx, &typePart.T)...)
		index.AddAttrs(typePart)
	}
	if kw, ok := annot.Kwarg("unique"); ok {
		diags = append(diags, kslspec.DecodeExpression(kw, d.ctx, &index.Unique)...)
	}
	index.AddColumns(c)
	return diags
}

func (d *sqldecoder) decodeUniqueColumn(r *Realm, c *Column, def *ast.Definition, annot *ast.Annotation) ksl.Diagnostics {
	var diags ksl.Diagnostics

	idx := c.Table.GetOrCreateIndex(DefaultUniqueIndexName(c.Table, c)).SetUnique(true)
	if idx.HasColumn(c.Name) {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagDuplicateIndexColumn,
			Detail:   fmt.Sprintf("Column %q is already in index %q.", c.Name, idx.Name),
			Subject:  annot.Range().Ptr(),
		})
		return diags
	}

	idx.AddColumns(c)
	return diags
}

func (d *sqldecoder) decodeColumnForeignKey(r *Realm, c *Column, def *ast.Definition, annot *ast.Annotation) ksl.Diagnostics {
	var diags ksl.Diagnostics

	name := DefaultForeignKeyName(c.Table, c)

	var columnRef string
	arg := annot.MustArg(0)
	diags = append(diags, kslspec.DecodeExpression(arg, d.ctx, &columnRef)...)
	colName := qualifiedColumnName(columnRef).Normalize(c.Table.Schema.Name, c.Table.Name)

	refTable, ok := r.Table(colName.Schema, colName.Table)
	if !ok {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagUnknownTableReference,
			Detail:   fmt.Sprintf("No table exists with name %q in schema %q.", colName.Table, colName.Schema),
			Subject:  annot.Range().Ptr(),
		})
		return diags
	}

	refColumn, ok := refTable.Column(colName.Column)
	if !ok {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidColumnReference,
			Detail:   fmt.Sprintf("Unknown column %q in table %q.", colName.Column, colName.Table),
			Subject:  annot.Range().Ptr(),
		})
		return diags
	}

	fk := NewForeignKey(name).SetTable(c.Table).AddColumns(c).SetRefTable(refTable).AddRefColumns(refColumn)
	if e, ok := annot.Kwarg("on_delete"); ok {
		diags = append(diags, kslspec.DecodeExpression(e, d.ctx, &fk.OnDelete)...)
	}
	if e, ok := annot.Kwarg("on_update"); ok {
		diags = append(diags, kslspec.DecodeExpression(e, d.ctx, &fk.OnUpdate)...)
	}

	c.Table.AddForeignKeys(fk)
	return diags
}

func getQualifiedTypeName(name string) *QualifiedTypeName {
	if idx := strings.LastIndexByte(name, '.'); idx != -1 {
		return &QualifiedTypeName{Schema: name[:idx], Name: name[idx+1:]}
	}
	return &QualifiedTypeName{Schema: "public", Name: name}
}

func qualifiedColumnName(name string) *QualifiedColumnName {
	parts := strings.SplitN(name, ".", 3)
	switch len(parts) {
	case 1:
		return &QualifiedColumnName{Column: parts[0]}
	case 2:
		return &QualifiedColumnName{Table: parts[0], Column: parts[1]}
	default:
		return &QualifiedColumnName{Schema: parts[0], Table: parts[1], Column: parts[2]}
	}
}
