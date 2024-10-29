package parse

import (
	"encoding/hex"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/core/types/validation"
	"github.com/kwilteam/kwil-db/parse/gen"
)

// schemaVisitor is a visitor for converting Kuneiform's ANTLR
// generated parse tree into our native schema type. It will perform
// syntax validation on actions and procedures.
type schemaVisitor struct {
	antlr.BaseParseTreeVisitor
	// schema is the schema that was parsed.
	// If no schema was parsed, it will be nil.
	schema *types.Schema
	// schemaInfo holds information on the position
	// of certain blocks in the schema.
	schemaInfo *SchemaInfo
	// errs is used for passing errors back to the caller.
	errs *errorListener
	// stream is the input stream
	stream *antlr.InputStream
	// both procedures and actions are only needed if parsing
	// an entire top-level schema, and will not be called if
	// parsing only an action or procedure body, or SQL.
	// procedures maps the asts of all parsed procedures
	procedures map[string][]ProcedureStmt
	// actions maps the asts of all parsed actions
	actions map[string][]ActionStmt
}

// getTextFromStream gets the text from the input stream for a given range.
// This is a hack over a bug in the generated antlr code, where it will try
// to access index out of bounds.
func (s *schemaVisitor) getTextFromStream(start, stop int) (str string) {
	defer func() {
		if r := recover(); r != nil {
			str = ""
		}
	}()

	return s.stream.GetText(start, stop)
}

// newSchemaVisitor creates a new schema visitor.
func newSchemaVisitor(stream *antlr.InputStream, errLis *errorListener) *schemaVisitor {
	return &schemaVisitor{
		schemaInfo: &SchemaInfo{
			Blocks: make(map[string]*Block),
		},
		errs:   errLis,
		stream: stream,
	}
}

var _ gen.KuneiformParserVisitor = (*schemaVisitor)(nil)

// below are the 4 top-level entry points for the visitor.
func (s *schemaVisitor) VisitSchema_entry(ctx *gen.Schema_entryContext) any {
	return ctx.Schema().Accept(s)
}

func (s *schemaVisitor) VisitAction_entry(ctx *gen.Action_entryContext) any {
	return ctx.Action_block().Accept(s)
}

func (s *schemaVisitor) VisitProcedure_entry(ctx *gen.Procedure_entryContext) any {
	return ctx.Procedure_block().Accept(s)
}

func (s *schemaVisitor) VisitSql_entry(ctx *gen.Sql_entryContext) any {
	return ctx.Sql_stmt().Accept(s)
}

// unknownExpression creates a new literal with an unknown type and null value.
// It should be used when we have to return early from a visitor method that
// returns an expression.
func unknownExpression(ctx antlr.ParserRuleContext) *ExpressionLiteral {
	e := &ExpressionLiteral{
		Type:  types.UnknownType,
		Value: nil,
	}

	e.Set(ctx)

	return e
}

func (s *schemaVisitor) VisitString_literal(ctx *gen.String_literalContext) any {
	str := ctx.STRING_().GetText()

	if !strings.HasPrefix(str, "'") || !strings.HasSuffix(str, "'") || len(str) < 2 {
		panic("invalid string literal")
	}
	str = str[1 : len(str)-1]

	n := &ExpressionLiteral{
		Type:  types.TextType,
		Value: str,
	}

	n.Set(ctx)

	return n
}

var (
	maxInt64 = big.NewInt(math.MaxInt64)
	minInt64 = big.NewInt(math.MinInt64)
)

func (s *schemaVisitor) VisitInteger_literal(ctx *gen.Integer_literalContext) any {
	i := ctx.DIGITS_().GetText()
	if ctx.MINUS() != nil {
		i = "-" + i
	}

	// integer literal can be either a uint256 or int64
	bigNum := new(big.Int)
	_, ok := bigNum.SetString(i, 10)
	if !ok {
		s.errs.RuleErr(ctx, ErrSyntax, "invalid integer literal: %s", i)
		return unknownExpression(ctx)
	}

	if bigNum.Cmp(maxInt64) > 0 {
		// it is a uint256
		u256, err := types.Uint256FromBig(bigNum)
		if err != nil {
			s.errs.RuleErr(ctx, ErrSyntax, "invalid integer literal: %s", i)
			return unknownExpression(ctx)
		}

		e := &ExpressionLiteral{
			Type:  types.Uint256Type,
			Value: u256,
		}
		e.Set(ctx)
		return e
	}
	if bigNum.Cmp(minInt64) < 0 {
		s.errs.RuleErr(ctx, ErrType, "integer literal out of range: %s", i)
		return unknownExpression(ctx)
	}

	e := &ExpressionLiteral{
		Type:  types.IntType,
		Value: bigNum.Int64(),
	}
	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitDecimal_literal(ctx *gen.Decimal_literalContext) any {
	// our decimal library can parse the decimal, so we simply pass it there
	txt := ctx.GetText()

	dec, err := decimal.NewFromString(txt)
	if err != nil {
		s.errs.RuleErr(ctx, err, "invalid decimal literal: %s", txt)
		return unknownExpression(ctx)
	}

	typ, err := types.NewDecimalType(dec.Precision(), dec.Scale())
	if err != nil {
		s.errs.RuleErr(ctx, err, "invalid decimal literal: %s", txt)
		return unknownExpression(ctx)
	}

	e := &ExpressionLiteral{
		Type:  typ,
		Value: dec,
	}
	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitBoolean_literal(ctx *gen.Boolean_literalContext) any {
	var b bool
	switch {
	case ctx.TRUE() != nil:
		b = true
	case ctx.FALSE() != nil:
		b = false
	default:
		panic("unknown boolean literal")
	}

	e := &ExpressionLiteral{
		Type:  types.BoolType,
		Value: b,
	}

	e.Set(ctx)

	return e
}

func (s *schemaVisitor) VisitNull_literal(ctx *gen.Null_literalContext) any {
	e := &ExpressionLiteral{
		Type:  types.NullType,
		Value: nil,
	}
	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitBinary_literal(ctx *gen.Binary_literalContext) any {
	b := ctx.GetText()
	// trim off beginning 0x
	if b[:2] != "0x" {
		// this should get caught by the parser
		s.errs.RuleErr(ctx, ErrSyntax, "invalid blob literal: %s", b)
	}

	b = b[2:]

	decoded, err := hex.DecodeString(b)
	if err != nil {
		// this should get caught by the parser
		s.errs.RuleErr(ctx, ErrSyntax, "invalid blob literal: %s", b)
	}

	e := &ExpressionLiteral{
		Type:  types.BlobType,
		Value: decoded,
	}
	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitIdentifier_list(ctx *gen.Identifier_listContext) any {
	var ident []string
	for _, i := range ctx.AllIdentifier() {
		ident = append(ident, i.Accept(s).(string))
	}

	return ident
}

func (s *schemaVisitor) VisitIdentifier(ctx *gen.IdentifierContext) any {
	return s.getIdent(ctx.IDENTIFIER())
}

func (s *schemaVisitor) VisitType(ctx *gen.TypeContext) any {
	dt := &types.DataType{
		Name: s.getIdent(ctx.IDENTIFIER()),
	}

	if ctx.LPAREN() != nil {
		// there should be 2 digits
		prec, err := strconv.ParseInt(ctx.DIGITS_(0).GetText(), 10, 64)
		if err != nil {
			s.errs.RuleErr(ctx, ErrSyntax, "invalid precision: %s", ctx.DIGITS_(0).GetText())
			return types.UnknownType
		}

		scale, err := strconv.ParseInt(ctx.DIGITS_(1).GetText(), 10, 64)
		if err != nil {
			s.errs.RuleErr(ctx, ErrSyntax, "invalid scale: %s", ctx.DIGITS_(1).GetText())
			return types.UnknownType
		}

		met := [2]uint16{uint16(prec), uint16(scale)}
		dt.Metadata = &met
	}

	if ctx.LBRACKET() != nil {
		dt.IsArray = true
	}

	err := dt.Clean()
	if err != nil {
		s.errs.RuleErr(ctx, err, "invalid type: %s", dt.String())
		return types.UnknownType
	}

	return dt
}

func (s *schemaVisitor) VisitType_cast(ctx *gen.Type_castContext) any {
	return s.Visit(ctx.Type_()).(*types.DataType)
}

func (s *schemaVisitor) VisitVariable(ctx *gen.VariableContext) any {
	var e *ExpressionVariable
	var tok antlr.Token
	switch {
	case ctx.VARIABLE() != nil:
		e = &ExpressionVariable{
			Name:   strings.ToLower(ctx.GetText()),
			Prefix: VariablePrefixDollar,
		}
		tok = ctx.VARIABLE().GetSymbol()
	case ctx.CONTEXTUAL_VARIABLE() != nil:
		e = &ExpressionVariable{
			Name:   strings.ToLower(ctx.GetText()),
			Prefix: VariablePrefixAt,
		}
		tok = ctx.CONTEXTUAL_VARIABLE().GetSymbol()

		_, ok := SessionVars[e.Name[1:]]
		if !ok {
			s.errs.RuleErr(ctx, ErrUnknownContextualVariable, e.Name)
		}

	default:
		panic("unknown variable")
	}

	s.validateVariableIdentifier(tok, e.Name)

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitVariable_list(ctx *gen.Variable_listContext) any {
	var vars []*ExpressionVariable
	for _, v := range ctx.AllVariable() {
		vars = append(vars, v.Accept(s).(*ExpressionVariable))
	}

	return vars
}

func (s *schemaVisitor) VisitSchema(ctx *gen.SchemaContext) any {
	s.schema = &types.Schema{
		Name:       ctx.Database_declaration().Accept(s).(string),
		Tables:     arr[*types.Table](len(ctx.AllTable_declaration())),
		Extensions: arr[*types.Extension](len(ctx.AllUse_declaration())),
		Actions:    arr[*types.Action](len(ctx.AllAction_declaration())),
		Procedures: arr[*types.Procedure](len(ctx.AllProcedure_declaration())),
	}

	for i, t := range ctx.AllTable_declaration() {
		s.schema.Tables[i] = t.Accept(s).(*types.Table)
		s.registerBlock(t, s.schema.Tables[i].Name)
	}

	// only now that we have visited all tables can we validate
	// foreign keys
	for _, t := range s.schema.Tables {
		for _, fk := range t.ForeignKeys {
			// the best we can do is get the position of the full
			// table.
			pos, ok := s.schemaInfo.Blocks[strings.ToLower(fk.ParentTable)]
			if !ok {
				pos2, ok2 := s.schemaInfo.Blocks[t.Name]
				if ok2 {
					s.errs.AddErr(pos2, ErrUnknownTable, fk.ParentTable)
				} else {
					s.errs.RuleErr(ctx, ErrUnknownTable, fk.ParentTable)
				}
				continue
			}

			// check that all ParentKeys exist
			parentTable, ok := s.schema.FindTable(fk.ParentTable)
			if !ok {
				s.errs.AddErr(pos, ErrUnknownTable, fk.ParentTable)
				continue
			}

			for _, col := range fk.ParentKeys {
				if _, ok := parentTable.FindColumn(col); !ok {
					s.errs.AddErr(pos, ErrUnknownColumn, col)
				}
			}
		}
	}

	for i, e := range ctx.AllUse_declaration() {
		s.schema.Extensions[i] = e.Accept(s).(*types.Extension)
		s.registerBlock(e, s.schema.Extensions[i].Alias)
	}

	for i, a := range ctx.AllAction_declaration() {
		s.schema.Actions[i] = a.Accept(s).(*types.Action)
		s.registerBlock(a, s.schema.Actions[i].Name)
	}

	for i, p := range ctx.AllProcedure_declaration() {
		s.schema.Procedures[i] = p.Accept(s).(*types.Procedure)
		s.registerBlock(p, s.schema.Procedures[i].Name)
	}

	return s.schema
}

// registerBlock registers a top-level block (table, action, procedure, etc.),
// ensuring uniqueness
func (s *schemaVisitor) registerBlock(ctx antlr.ParserRuleContext, name string) {
	lower := strings.ToLower(name)
	if _, ok := s.schemaInfo.Blocks[lower]; ok {
		s.errs.RuleErr(ctx, ErrDuplicateBlock, lower)
		return
	}

	if _, ok := Functions[lower]; ok {
		s.errs.RuleErr(ctx, ErrReservedKeyword, lower)
		return
	}

	if validation.IsKeyword(lower) {
		s.errs.RuleErr(ctx, ErrReservedKeyword, lower)
		return
	}

	node := &Position{}
	node.Set(ctx)

	s.schemaInfo.Blocks[lower] = &Block{
		Position: *node,
		AbsStart: ctx.GetStart().GetStart(),
		AbsEnd:   ctx.GetStop().GetStop(),
	}
}

// arr will make an array of type A if the input is greater than 0
func arr[A any](b int) []A {
	if b > 0 {
		return make([]A, b)
	}
	return nil
}

func (s *schemaVisitor) VisitAnnotation(ctx *gen.AnnotationContext) any {
	// we will parse but reconstruct annotations, so they can later be consumed by the gateway
	str := strings.Builder{}

	str.WriteString(s.getIdent(ctx.CONTEXTUAL_VARIABLE()))
	str.WriteString("(")
	for i, l := range ctx.AllLiteral() {
		if i > 0 {
			str.WriteString(", ")
		}

		str.WriteString(s.getIdent(ctx.IDENTIFIER(i)))
		str.WriteString("=")
		// we do not touch the literal, since case should be preserved
		str.WriteString(l.GetText())
	}
	str.WriteString(")")

	return str.String()
}

// isErrNode is true if an antlr terminal node is an error node.
func isErrNode(node antlr.TerminalNode) bool {
	_, ok := node.(antlr.ErrorNode)
	return ok
}

func (s *schemaVisitor) VisitDatabase_declaration(ctx *gen.Database_declarationContext) any {
	// needed to avoid https://github.com/kwilteam/kwil-db/issues/752
	if isErrNode(ctx.DATABASE()) {
		return ""
	}

	return s.getIdent(ctx.IDENTIFIER())
}

func (s *schemaVisitor) VisitUse_declaration(ctx *gen.Use_declarationContext) any {
	// the first identifier is the extension name, the last is the alias,
	// and all in between are keys in the initialization.
	e := &types.Extension{
		Name:           s.getIdent(ctx.IDENTIFIER(0)),
		Initialization: arr[*types.ExtensionConfig](len(ctx.AllIDENTIFIER()) - 2),
		Alias:          s.getIdent(ctx.IDENTIFIER(len(ctx.AllIDENTIFIER()) - 1)),
	}

	for i, id := range ctx.AllIDENTIFIER()[1 : len(ctx.AllIDENTIFIER())-1] {
		val := ctx.Literal(i).Accept(s).(*ExpressionLiteral)

		e.Initialization[i] = &types.ExtensionConfig{
			Key:   s.getIdent(id),
			Value: val.String(),
		}
	}

	return e
}

func (s *schemaVisitor) VisitTable_declaration(ctx *gen.Table_declarationContext) any {
	t := &types.Table{
		Name:        s.getIdent(ctx.IDENTIFIER()),
		Columns:     arr[*types.Column](len(ctx.AllColumn_def())),
		Indexes:     arr[*types.Index](len(ctx.AllIndex_def())),
		ForeignKeys: arr[*types.ForeignKey](len(ctx.AllForeign_key_def())),
	}

	for i, c := range ctx.AllColumn_def() {
		t.Columns[i] = c.Accept(s).(*types.Column)
	}

	for i, idx := range ctx.AllIndex_def() {
		t.Indexes[i] = idx.Accept(s).(*types.Index)

		// check that all columns in indexes and foreign key children exist
		for _, col := range t.Indexes[i].Columns {
			if _, ok := t.FindColumn(col); !ok {
				s.errs.RuleErr(idx, ErrUnknownColumn, col)
			}
		}
	}

	for i, fk := range ctx.AllForeign_key_def() {
		t.ForeignKeys[i] = fk.Accept(s).(*types.ForeignKey)

		// check that all ChildKeys exist.
		// we will have to check for parent keys in a later stage,
		// since not all tables are parsed yet.
		for _, col := range t.ForeignKeys[i].ChildKeys {
			if _, ok := t.FindColumn(col); !ok {
				s.errs.RuleErr(fk, ErrUnknownColumn, col)
			}
		}
	}

	_, err := t.GetPrimaryKey()
	if err != nil {
		s.errs.RuleErr(ctx, ErrNoPrimaryKey, err.Error())
	}

	return t
}

func (s *schemaVisitor) VisitColumn_def(ctx *gen.Column_defContext) any {
	col := &types.Column{
		Name: s.getIdent(ctx.IDENTIFIER()),
		Type: ctx.Type_().Accept(s).(*types.DataType),
	}

	// due to unfortunate lexing edge cases to support min/max, we
	// have to parse the constraints here. Each constraint is a text, and should be
	// one of:
	// MIN/MAX/MINLEN/MAXLEN/MIN_LENGTH/MAX_LENGTH/NOTNULL/NOT/NULL/PRIMARY/KEY/PRIMARY_KEY/PK/DEFAULT/UNIQUE
	// If NOT is present, it needs to be followed by NULL; similarly, if NULL is present, it needs to be preceded by NOT.
	// If PRIMARY is present, it can be followed by key, but does not have to be. key must be preceded by primary.
	// MIN, MAX, MINLEN, MAXLEN, MIN_LENGTH, MAX_LENGTH, and DEFAULT must also have a literal following them.
	type constraint struct {
		ident string
		lit   *string
	}
	constraints := make([]constraint, len(ctx.AllConstraint()))
	for i, c := range ctx.AllConstraint() {
		con := constraint{}
		switch {
		case c.IDENTIFIER() != nil:
			con.ident = c.IDENTIFIER().GetText()
		case c.PRIMARY() != nil:
			con.ident = "primary_key"
		case c.NOT() != nil:
			con.ident = "notnull"
		case c.DEFAULT() != nil:
			con.ident = "default"
		case c.UNIQUE() != nil:
			con.ident = "unique"
		default:
			panic("unknown constraint")
		}

		if c.Literal() != nil {
			l := strings.ToLower(c.Literal().Accept(s).(*ExpressionLiteral).String())
			con.lit = &l
		}
		constraints[i] = con
	}

	for i := range constraints {
		switch constraints[i].ident {
		case "min":
			if constraints[i].lit == nil {
				s.errs.RuleErr(ctx, ErrSyntax, "missing literal for min constraint")
				return col
			}
			col.Attributes = append(col.Attributes, &types.Attribute{
				Type:  types.MIN,
				Value: *constraints[i].lit,
			})
		case "max":
			if constraints[i].lit == nil {
				s.errs.RuleErr(ctx, ErrSyntax, "missing literal for max constraint")
				return col
			}
			col.Attributes = append(col.Attributes, &types.Attribute{
				Type:  types.MAX,
				Value: *constraints[i].lit,
			})
		case "minlen", "min_length":
			if constraints[i].lit == nil {
				s.errs.RuleErr(ctx, ErrSyntax, "missing literal for min length constraint")
				return col
			}
			col.Attributes = append(col.Attributes, &types.Attribute{
				Type:  types.MIN_LENGTH,
				Value: *constraints[i].lit,
			})
		case "maxlen", "max_length":
			if constraints[i].lit == nil {
				s.errs.RuleErr(ctx, ErrSyntax, "missing literal for max length constraint")
				return col
			}
			col.Attributes = append(col.Attributes, &types.Attribute{
				Type:  types.MAX_LENGTH,
				Value: *constraints[i].lit,
			})
		case "notnull":
			if constraints[i].lit != nil {
				s.errs.RuleErr(ctx, ErrSyntax, "unexpected literal for not null constraint")
				return col
			}
			col.Attributes = append(col.Attributes, &types.Attribute{
				Type: types.NOT_NULL,
			})
		case "primary_key", "pk":
			if constraints[i].lit != nil {
				s.errs.RuleErr(ctx, ErrSyntax, "unexpected literal for primary key constraint")
				return col
			}
			col.Attributes = append(col.Attributes, &types.Attribute{
				Type: types.PRIMARY_KEY,
			})
		case "default":
			if constraints[i].lit == nil {
				s.errs.RuleErr(ctx, ErrSyntax, "missing literal for default constraint")
				return col
			}
			col.Attributes = append(col.Attributes, &types.Attribute{
				Type:  types.DEFAULT,
				Value: *constraints[i].lit,
			})
		case "unique":
			if constraints[i].lit != nil {
				s.errs.RuleErr(ctx, ErrSyntax, "unexpected literal for unique constraint")
				return col
			}
			col.Attributes = append(col.Attributes, &types.Attribute{
				Type: types.UNIQUE,
			})
		default:
			s.errs.RuleErr(ctx, ErrSyntax, "unknown constraint: %s", constraints[i].ident)
			return col
		}
	}

	for _, con := range col.Attributes {
		err := con.Clean(col)
		if err != nil {
			s.errs.RuleErr(ctx, ErrColumnConstraint, err.Error())
			return col
		}
	}

	return col
}

func (s *schemaVisitor) VisitConstraint(ctx *gen.ConstraintContext) any {
	panic("VisitConstraint should not be called, as the logic should be implemented in VisitColumn_def")
}

func (s *schemaVisitor) VisitIndex_def(ctx *gen.Index_defContext) any {
	name := ctx.HASH_IDENTIFIER().GetText()
	name = strings.TrimLeft(name, "#")
	idx := &types.Index{
		Name:    strings.ToLower(name),
		Columns: ctx.Identifier_list().Accept(s).([]string),
	}

	s.validateVariableIdentifier(ctx.HASH_IDENTIFIER().GetSymbol(), idx.Name)

	switch {
	case ctx.INDEX() != nil:
		idx.Type = types.BTREE
	case ctx.UNIQUE() != nil:
		idx.Type = types.UNIQUE_BTREE
	case ctx.PRIMARY() != nil:
		idx.Type = types.PRIMARY
	default:
		panic("unknown index type")
	}

	return idx
}

func (s *schemaVisitor) VisitForeign_key_def(ctx *gen.Foreign_key_defContext) any {
	fk := &types.ForeignKey{
		ChildKeys:   ctx.GetChild_keys().Accept(s).([]string),
		ParentKeys:  ctx.GetParent_keys().Accept(s).([]string),
		ParentTable: strings.ToLower(ctx.GetParent_table().GetText()),
		Actions:     arr[*types.ForeignKeyAction](len(ctx.AllForeign_key_action())),
	}

	for i, a := range ctx.AllForeign_key_action() {
		fk.Actions[i] = a.Accept(s).(*types.ForeignKeyAction)
	}

	return fk
}

func (s *schemaVisitor) VisitForeign_key_action(ctx *gen.Foreign_key_actionContext) any {
	ac := &types.ForeignKeyAction{}
	switch {
	case ctx.UPDATE() != nil, ctx.LEGACY_ON_UPDATE() != nil:
		ac.On = types.ON_UPDATE
	case ctx.DELETE() != nil, ctx.LEGACY_ON_DELETE() != nil:
		ac.On = types.ON_DELETE
	default:
		panic("unknown foreign key action")
	}

	switch {
	case ctx.ACTION() != nil, ctx.LEGACY_NO_ACTION() != nil:
		ac.Do = types.DO_NO_ACTION
	case ctx.NULL() != nil, ctx.LEGACY_SET_NULL() != nil:
		ac.Do = types.DO_SET_NULL
	case ctx.DEFAULT() != nil, ctx.LEGACY_SET_DEFAULT() != nil:
		ac.Do = types.DO_SET_DEFAULT
		// cascade and restrict do not have legacys
	case ctx.CASCADE() != nil:
		ac.Do = types.DO_CASCADE
	case ctx.RESTRICT() != nil:
		ac.Do = types.DO_RESTRICT
	default:
		panic("unknown foreign key action")
	}

	return ac
}

func (s *schemaVisitor) VisitType_list(ctx *gen.Type_listContext) any {
	var ts []*types.DataType
	for _, t := range ctx.AllType_() {
		ts = append(ts, t.Accept(s).(*types.DataType))
	}

	return ts
}

func (s *schemaVisitor) VisitNamed_type_list(ctx *gen.Named_type_listContext) any {
	var ts []*types.NamedType
	for i, t := range ctx.AllIDENTIFIER() {
		ts = append(ts, &types.NamedType{
			Name: s.getIdent(t),
			Type: ctx.Type_(i).Accept(s).(*types.DataType),
		})
	}

	return ts
}

func (s *schemaVisitor) VisitTyped_variable_list(ctx *gen.Typed_variable_listContext) any {
	var vars []*types.ProcedureParameter
	for i, v := range ctx.AllVariable() {
		vars = append(vars, &types.ProcedureParameter{
			Name: v.Accept(s).(*ExpressionVariable).String(),
			Type: ctx.Type_(i).Accept(s).(*types.DataType),
		})
	}

	return vars
}

func (s *schemaVisitor) VisitAccess_modifier(ctx *gen.Access_modifierContext) any {
	// we will have to parse this at a later stage, since this is either public/private,
	// or a types.Modifier
	panic("VisitAccess_modifier should not be called")
}

// getModifiersAndPublicity parses access modifiers and returns. it should be used when
// parsing procedures and actions
func getModifiersAndPublicity(ctxs []gen.IAccess_modifierContext) (public bool, mods []types.Modifier) {
	for _, ctx := range ctxs {
		switch {
		case ctx.PUBLIC() != nil:
			public = true
		case ctx.PRIVATE() != nil:
			public = false
		case ctx.VIEW() != nil:
			mods = append(mods, types.ModifierView)
		case ctx.OWNER() != nil:
			mods = append(mods, types.ModifierOwner)
		default:
			// should not happen, as this would suggest a bug in the parser
			panic("unknown access modifier")
		}
	}

	return
}

func (s *schemaVisitor) VisitAction_declaration(ctx *gen.Action_declarationContext) any {
	act := &types.Action{
		Name:        s.getIdent(ctx.IDENTIFIER()),
		Annotations: arr[string](len(ctx.AllAnnotation())),
	}

	for i, a := range ctx.AllAnnotation() {
		act.Annotations[i] = a.Accept(s).(string)
	}

	public, mods := getModifiersAndPublicity(ctx.AllAccess_modifier())
	act.Public = public
	act.Modifiers = mods

	if ctx.Variable_list() != nil {
		params := ctx.Variable_list().Accept(s).([]*ExpressionVariable)
		paramStrs := make([]string, len(params))
		for i, p := range params {
			paramStrs[i] = p.String()
		}
		act.Parameters = paramStrs
	}

	act.Body = s.getTextFromStream(ctx.Action_block().GetStart().GetStart(), ctx.Action_block().GetStop().GetStop())

	ast := ctx.Action_block().Accept(s).([]ActionStmt)
	s.actions[act.Name] = ast

	return act
}

func (s *schemaVisitor) VisitProcedure_declaration(ctx *gen.Procedure_declarationContext) any {
	proc := &types.Procedure{
		Name:        s.getIdent(ctx.IDENTIFIER()),
		Annotations: arr[string](len(ctx.AllAnnotation())),
	}

	if ctx.Typed_variable_list() != nil {
		proc.Parameters = ctx.Typed_variable_list().Accept(s).([]*types.ProcedureParameter)
	}

	if ctx.Procedure_return() != nil {
		proc.Returns = ctx.Procedure_return().Accept(s).(*types.ProcedureReturn)
	}

	for i, a := range ctx.AllAnnotation() {
		proc.Annotations[i] = a.Accept(s).(string)
	}

	public, mods := getModifiersAndPublicity(ctx.AllAccess_modifier())
	proc.Public = public
	proc.Modifiers = mods

	ast := ctx.Procedure_block().Accept(s).([]ProcedureStmt)
	s.procedures[proc.Name] = ast

	proc.Body = s.getTextFromStream(ctx.Procedure_block().GetStart().GetStart(), ctx.Procedure_block().GetStop().GetStop())

	return proc
}

func (s *schemaVisitor) VisitProcedure_return(ctx *gen.Procedure_returnContext) any {
	ret := &types.ProcedureReturn{}

	switch {
	case ctx.GetReturn_columns() != nil:
		ret.Fields = ctx.GetReturn_columns().Accept(s).([]*types.NamedType)
	case ctx.GetUnnamed_return_types() != nil:
		ret.Fields = make([]*types.NamedType, len(ctx.GetUnnamed_return_types().AllType_()))
		for i, t := range ctx.GetUnnamed_return_types().AllType_() {
			ret.Fields[i] = &types.NamedType{
				Name: "col" + fmt.Sprint(i),
				Type: t.Accept(s).(*types.DataType),
			}
		}
	default:
		panic("unknown return type")
	}

	if ctx.TABLE() != nil {
		ret.IsTable = true
	}

	return ret
}

// VisitSql_stmt_s visits a SQL statement. It is the top-level SQL visitor.
func (s *schemaVisitor) VisitSql_stmt(ctx *gen.Sql_stmtContext) any {
	// NOTE: this should be temporary; we should combine dml and ddl.
	if ctx.Sql_statement() != nil {
		return ctx.Sql_statement().Accept(s)
	} else {
		return ctx.Ddl_stmt().Accept(s)
	}
}

// VisitDdl_stmt visits a SQL DDL statement.
func (s *schemaVisitor) VisitDdl_stmt(ctx *gen.Ddl_stmtContext) any {
	switch {
	case ctx.Create_table_statement() != nil:
		return ctx.Create_table_statement().Accept(s).(*CreateTableStatement)
	case ctx.Alter_table_statement() != nil:
		return ctx.Alter_table_statement().Accept(s).(*AlterTableStatement)
	case ctx.Create_index_statement() != nil:
		return ctx.Create_index_statement().Accept(s).(*CreateIndexStatement)
	case ctx.Drop_index_statement() != nil:
		return ctx.Drop_index_statement().Accept(s).(*DropIndexStatement)
	default:
		panic("unknown DDL statement")
	}
}

// VisitSql_statement visits a SQL DML statement. It is called by all nested
// sql statements (e.g. in procedures and actions)
func (s *schemaVisitor) VisitSql_statement(ctx *gen.Sql_statementContext) any {
	stmt := &SQLStatement{
		CTEs: arr[*CommonTableExpression](len(ctx.AllCommon_table_expression())),
	}

	for i, cte := range ctx.AllCommon_table_expression() {
		stmt.CTEs[i] = cte.Accept(s).(*CommonTableExpression)
	}

	if ctx.RECURSIVE() != nil {
		stmt.Recursive = true
	}

	switch {
	case ctx.Select_statement() != nil:
		stmt.SQL = ctx.Select_statement().Accept(s).(*SelectStatement)
	case ctx.Update_statement() != nil:
		stmt.SQL = ctx.Update_statement().Accept(s).(*UpdateStatement)
	case ctx.Insert_statement() != nil:
		stmt.SQL = ctx.Insert_statement().Accept(s).(*InsertStatement)
	case ctx.Delete_statement() != nil:
		stmt.SQL = ctx.Delete_statement().Accept(s).(*DeleteStatement)
	default:
		panic("unknown dml statement")
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitCommon_table_expression(ctx *gen.Common_table_expressionContext) any {
	// first identifier is the table name, the rest are the columns
	cte := &CommonTableExpression{
		Name:  ctx.Identifier(0).Accept(s).(string),
		Query: ctx.Select_statement().Accept(s).(*SelectStatement),
	}

	for _, id := range ctx.AllIdentifier()[1:] {
		cte.Columns = append(cte.Columns, id.Accept(s).(string))
	}

	cte.Set(ctx)

	return cte
}

func (s *schemaVisitor) VisitCreate_table_statement(ctx *gen.Create_table_statementContext) any {
	stmt := &CreateTableStatement{
		Name:        ctx.GetName().Accept(s).(string),
		Columns:     arr[*Column](len(ctx.AllTable_column_def())),
		Constraints: arr[Constraint](len(ctx.AllTable_constraint_def())),
		Indexes:     arr[*TableIndex](len(ctx.AllTable_index_def())),
	}

	if ctx.EXISTS() != nil {
		stmt.IfNotExists = true
	}

	// for basic validation
	var primaryKey []string
	allColumns := make(map[string]bool)
	allConstraints := make(map[string]bool)
	allIndexes := make(map[string]bool)

	if len(ctx.AllTable_column_def()) == 0 {
		s.errs.RuleErr(ctx, ErrTableDefinition, "no column definitions found")
	}
	for i, c := range ctx.AllTable_column_def() {
		col := c.Accept(s).(*Column)
		stmt.Columns[i] = col
		if allColumns[col.Name] {
			s.errs.RuleErr(c, ErrCollation, "constraint name exists")
		} else {
			allColumns[col.Name] = true
		}
	}

	for _, column := range stmt.Columns {
		for _, constraint := range column.Constraints {
			switch constraint.(type) {
			case *ConstraintPrimaryKey:
				primaryKey = []string{column.Name}
			}
		}
	}

	for i, c := range ctx.AllTable_constraint_def() {
		constraint := c.Accept(s).(Constraint)
		stmt.Constraints[i] = constraint

		switch cc := constraint.(type) {
		case *ConstraintPrimaryKey:
			if len(primaryKey) != 0 {
				s.errs.RuleErr(c, ErrRedeclarePrimaryKey, "primary key redeclared")
				continue
			}

			for _, col := range cc.Columns {
				if !allColumns[col] {
					s.errs.RuleErr(c, ErrUnknownColumn, "primary key on unknown column")
				}
			}

			primaryKey = cc.Columns
		case *ConstraintCheck:
			if cc.Name != "" {
				if allConstraints[cc.Name] {
					s.errs.RuleErr(c, ErrCollation, "constraint name exists")
				} else {
					allConstraints[cc.Name] = true
				}
			}
		case *ConstraintUnique:
			if cc.Name != "" {
				if allConstraints[cc.Name] {
					s.errs.RuleErr(c, ErrCollation, "constraint name exists")
				} else {
					allConstraints[cc.Name] = true
				}
			}

			for _, col := range cc.Columns {
				if !allColumns[col] {
					s.errs.RuleErr(c, ErrUnknownColumn, "primary key on unknown column")
				}
			}
		case *ConstraintForeignKey:
			if cc.Name != "" {
				if allConstraints[cc.Name] {
					s.errs.RuleErr(c, ErrCollation, "constraint name exists")
				} else {
					allConstraints[cc.Name] = true
				}
			}

			if !allColumns[cc.RefColumn] {
				s.errs.RuleErr(c, ErrUnknownColumn, "index on unknown column")
			}
		default:
			// should not happen
			panic("unknown constraint type")
		}
	}

	for i, c := range ctx.AllTable_index_def() {
		idx := c.Accept(s).(*TableIndex)
		stmt.Indexes[i] = idx

		if idx.Name != "" {
			if allIndexes[idx.Name] {
				s.errs.RuleErr(c, ErrCollation, "index name exists")
			} else {
				allIndexes[idx.Name] = true
			}
		}

		for _, col := range idx.Columns {
			if !allColumns[col] {
				s.errs.RuleErr(c, ErrUnknownColumn, "index on unknown column")
			}
		}
	}

	if len(primaryKey) == 0 {
		s.errs.RuleErr(ctx, ErrNoPrimaryKey, "no primary key declared")
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitTable_column_def(ctx *gen.Table_column_defContext) interface{} {
	column := &Column{
		Name:        s.getIdent(ctx.IDENTIFIER()),
		Type:        ctx.Type_().Accept(s).(*types.DataType),
		Constraints: arr[Constraint](len(ctx.AllInline_constraint())),
	}

	for i, c := range ctx.AllInline_constraint() {
		column.Constraints[i] = c.Accept(s).(Constraint)
	}

	column.Set(ctx)
	return column
}

func (s *schemaVisitor) VisitInline_constraint(ctx *gen.Inline_constraintContext) any {
	switch {
	case ctx.PRIMARY() != nil:
		c := &ConstraintPrimaryKey{}
		c.Set(ctx)
		return c
	case ctx.UNIQUE() != nil:
		c := &ConstraintUnique{}
		c.Set(ctx)
		return c
	case ctx.NOT() != nil:
		c := &ConstraintNotNull{}
		c.Set(ctx)
		return c
	case ctx.DEFAULT() != nil:
		c := &ConstraintDefault{
			Value: ctx.Literal().Accept(s).(*ExpressionLiteral),
		}
		c.Set(ctx)
		return c
	case ctx.CHECK() != nil:
		c := &ConstraintCheck{
			Param: ctx.Sql_expr().Accept(s).(Expression),
		}
		c.Set(ctx)
		return c
	case ctx.Fk_constraint() != nil:
		c := ctx.Fk_constraint().Accept(s).(*ConstraintForeignKey)
		return c
	default:
		panic("unknown constraint")
	}
}

func (s *schemaVisitor) VisitFk_constraint(ctx *gen.Fk_constraintContext) any {
	c := &ConstraintForeignKey{
		RefTable:  ctx.GetTable().Accept(s).(string),
		RefColumn: ctx.GetColumn().Accept(s).(string),
		Ons:       arr[ForeignKeyActionOn](len(ctx.AllFk_action())),
		Dos:       arr[ForeignKeyActionDo](len(ctx.AllFk_action())),
	}

	for i, a := range ctx.AllFk_action() {
		switch {
		case a.UPDATE() != nil:
			c.Ons[i] = ON_UPDATE
		case a.DELETE() != nil:
			c.Ons[i] = ON_DELETE
		default:
			panic("unknown foreign key on condition")
		}

		switch {
		case a.NULL() != nil:
			c.Dos[i] = DO_SET_NULL
		case a.DEFAULT() != nil:
			c.Dos[i] = DO_SET_DEFAULT
		case a.RESTRICT() != nil:
			c.Dos[i] = DO_RESTRICT
		case a.ACTION() != nil:
			c.Dos[i] = DO_NO_ACTION
		case a.CASCADE() != nil:
			c.Dos[i] = DO_CASCADE
		default:
			panic("unknown foreign key action")
		}
	}

	c.Set(ctx)
	return c
}

func (s *schemaVisitor) VisitFk_action(ctx *gen.Fk_actionContext) interface{} {
	panic("implement me")
}

func (s *schemaVisitor) VisitTable_constraint_def(ctx *gen.Table_constraint_defContext) any {
	name := ""
	if ctx.GetName() != nil {
		name = ctx.GetName().Accept(s).(string)
	}

	switch {
	case ctx.PRIMARY() != nil:
		if name != "" {
			s.errs.RuleErr(ctx, ErrTableDefinition, "primary key has name")
		}
		c := &ConstraintPrimaryKey{
			Columns: ctx.Identifier_list().Accept(s).([]string),
		}
		c.Set(ctx)
		return c
	case ctx.UNIQUE() != nil:
		c := &ConstraintUnique{
			Name:    name,
			Columns: ctx.Identifier_list().Accept(s).([]string),
		}
		c.Set(ctx)
		return c
	case ctx.CHECK() != nil:
		param := ctx.Sql_expr().Accept(s).(Expression)
		c := &ConstraintCheck{
			Name:  name,
			Param: param,
		}
		c.Set(ctx)
		return c
	case ctx.FOREIGN() != nil:
		c := ctx.Fk_constraint().Accept(s).(*ConstraintForeignKey)
		c.Name = name
		c.Column = ctx.GetColumn().Accept(s).(string)
		return c
	default:
		panic("unknown constraint")
	}
}

func (s *schemaVisitor) VisitTable_index_def(ctx *gen.Table_index_defContext) any {
	index := &TableIndex{
		Name:    ctx.Identifier().Accept(s).(string),
		Columns: ctx.Identifier_list().Accept(s).([]string),
		Type:    IndexTypeBTree,
	}

	if ctx.UNIQUE() != nil {
		index.Type = IndexTypeUnique
	}

	index.Set(ctx)
	index.Set(ctx)
	return index
}

func (s *schemaVisitor) VisitAlter_table_statement(ctx *gen.Alter_table_statementContext) any {
	stmt := &AlterTableStatement{
		Table:  ctx.Identifier().Accept(s).(string),
		Action: ctx.Alter_table_action().Accept(s).(AlterTableAction),
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitAdd_column_constraint(ctx *gen.Add_column_constraintContext) any {
	a := &AddColumnConstraint{
		Column: ctx.Identifier().Accept(s).(string),
	}

	if ctx.NULL() != nil {
		a.Type = NOT_NULL
	} else {
		a.Type = DEFAULT
		a.Value = ctx.Literal().Accept(s).(*ExpressionLiteral)
	}

	a.Set(ctx)
	return a
}

func (s *schemaVisitor) VisitDrop_column_constraint(ctx *gen.Drop_column_constraintContext) any {
	a := &DropColumnConstraint{
		Column: ctx.Identifier(0).Accept(s).(string),
	}

	switch {
	case ctx.NULL() != nil:
		a.Type = NOT_NULL
	case ctx.DEFAULT() != nil:
		a.Type = DEFAULT
	case ctx.CONSTRAINT() != nil:
		a.Name = ctx.Identifier(1).Accept(s).(string)
	default:
		panic("unknown constraint")
	}

	a.Set(ctx)
	return a
}

func (s *schemaVisitor) VisitAdd_column(ctx *gen.Add_columnContext) any {
	a := &AddColumn{
		Name: ctx.Identifier().Accept(s).(string),
		Type: ctx.Type_().Accept(s).(*types.DataType),
	}

	a.Set(ctx)
	return a
}

func (s *schemaVisitor) VisitDrop_column(ctx *gen.Drop_columnContext) any {
	a := &DropColumn{
		Name: ctx.Identifier().Accept(s).(string),
	}

	a.Set(ctx)
	return a
}

func (s *schemaVisitor) VisitRename_column(ctx *gen.Rename_columnContext) any {
	a := &RenameColumn{
		OldName: ctx.GetOld_column().Accept(s).(string),
		NewName: ctx.GetNew_column().Accept(s).(string),
	}

	a.Set(ctx)
	return a
}

func (s *schemaVisitor) VisitRename_table(ctx *gen.Rename_tableContext) any {
	a := &RenameTable{
		Name: ctx.Identifier().Accept(s).(string),
	}

	a.Set(ctx)
	return a
}

func (s *schemaVisitor) VisitAdd_table_constraint(ctx *gen.Add_table_constraintContext) any {
	a := &AddTableConstraint{
		Cons: ctx.Table_constraint_def().Accept(s).(Constraint),
	}

	a.Set(ctx)
	return a
}

func (s *schemaVisitor) VisitDrop_table_constraint(ctx *gen.Drop_table_constraintContext) any {
	a := &DropTableConstraint{
		Name: ctx.Identifier().Accept(s).(string),
	}

	a.Set(ctx)
	return a
}

func (s *schemaVisitor) VisitCreate_index_statement(ctx *gen.Create_index_statementContext) any {
	a := &CreateIndexStatement{
		On:      ctx.GetTable().Accept(s).(string),
		Columns: ctx.GetColumns().Accept(s).([]string),
		Type:    IndexTypeBTree,
	}

	if ctx.EXISTS() != nil {
		a.IfNotExists = true
	}

	if ctx.GetName() != nil {
		a.Name = ctx.GetName().Accept(s).(string)
	}

	if ctx.UNIQUE() != nil {
		a.Type = IndexTypeUnique
	}

	a.Set(ctx)
	return a
}

func (s *schemaVisitor) VisitDrop_index_statement(ctx *gen.Drop_index_statementContext) interface{} {
	a := &DropIndexStatement{
		Name: ctx.Identifier().Accept(s).(string),
	}

	if ctx.EXISTS() != nil {
		a.CheckExist = true
	}

	a.Set(ctx)
	return a
}

func (s *schemaVisitor) VisitSelect_statement(ctx *gen.Select_statementContext) any {
	stmt := &SelectStatement{
		SelectCores:       arr[*SelectCore](len(ctx.AllSelect_core())),
		CompoundOperators: arr[CompoundOperator](len(ctx.AllCompound_operator())),
		Ordering:          arr[*OrderingTerm](len(ctx.AllOrdering_term())),
	}

	for i, core := range ctx.AllSelect_core() {
		stmt.SelectCores[i] = core.Accept(s).(*SelectCore)
	}

	for i, op := range ctx.AllCompound_operator() {
		stmt.CompoundOperators[i] = op.Accept(s).(CompoundOperator)
	}

	for i, ord := range ctx.AllOrdering_term() {
		stmt.Ordering[i] = ord.Accept(s).(*OrderingTerm)
	}

	if ctx.GetLimit() != nil {
		stmt.Limit = ctx.GetLimit().Accept(s).(Expression)
	}
	if ctx.GetOffset() != nil {
		stmt.Offset = ctx.GetOffset().Accept(s).(Expression)
	}

	stmt.Set(ctx)

	return stmt
}

func (s *schemaVisitor) VisitCompound_operator(ctx *gen.Compound_operatorContext) any {
	switch {
	case ctx.UNION() != nil:
		if ctx.ALL() != nil {
			return CompoundOperatorUnionAll
		}
		return CompoundOperatorUnion
	case ctx.INTERSECT() != nil:
		return CompoundOperatorIntersect
	case ctx.EXCEPT() != nil:
		return CompoundOperatorExcept
	default:
		panic("unknown compound operator")
	}
}

func (s *schemaVisitor) VisitOrdering_term(ctx *gen.Ordering_termContext) any {
	ord := &OrderingTerm{
		Expression: ctx.Sql_expr().Accept(s).(Expression),
		Order:      OrderTypeAsc,
		Nulls:      NullOrderLast,
	}

	if ctx.DESC() != nil {
		ord.Order = OrderTypeDesc
	}

	if ctx.FIRST() != nil {
		ord.Nulls = NullOrderFirst
	}

	ord.Set(ctx)

	return ord
}

func (s *schemaVisitor) VisitSelect_core(ctx *gen.Select_coreContext) any {
	stmt := &SelectCore{
		Columns: arr[ResultColumn](len(ctx.AllResult_column())),
		Joins:   arr[*Join](len(ctx.AllJoin())),
	}

	if ctx.DISTINCT() != nil {
		stmt.Distinct = true
	}

	for i, col := range ctx.AllResult_column() {
		stmt.Columns[i] = col.Accept(s).(ResultColumn)
	}

	if ctx.Relation() != nil {
		stmt.From = ctx.Relation().Accept(s).(Table)
	}

	for i, join := range ctx.AllJoin() {
		stmt.Joins[i] = join.Accept(s).(*Join)
	}

	if ctx.GetWhere() != nil {
		stmt.Where = ctx.GetWhere().Accept(s).(Expression)
	}

	if ctx.GetGroup_by() != nil {
		stmt.GroupBy = ctx.GetGroup_by().Accept(s).([]Expression)
	}

	if ctx.GetHaving() != nil {
		stmt.Having = ctx.GetHaving().Accept(s).(Expression)
	}

	if ctx.WINDOW() != nil {
		// the only Identifier used in the SELECT CORE grammar is for naming windows,
		// so we can safely get identifiers by index here.
		for i, window := range ctx.AllWindow() {
			name := s.getIdent(ctx.Identifier(i).IDENTIFIER())

			win := window.Accept(s).(*WindowImpl)

			stmt.Windows = append(stmt.Windows, &struct {
				Name   string
				Window *WindowImpl
			}{
				Name:   name,
				Window: win,
			})
		}
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitTable_relation(ctx *gen.Table_relationContext) any {
	t := &RelationTable{
		Table: strings.ToLower(ctx.GetTable_name().Accept(s).(string)),
	}

	if ctx.GetAlias() != nil {
		t.Alias = strings.ToLower(ctx.GetAlias().Accept(s).(string))
	}

	t.Set(ctx)
	return t
}

func (s *schemaVisitor) VisitSubquery_relation(ctx *gen.Subquery_relationContext) any {
	t := &RelationSubquery{
		Subquery: ctx.Select_statement().Accept(s).(*SelectStatement),
	}

	// alias is technially required here, but we allow it in the grammar
	// to throw a better error message here.
	if ctx.Identifier() != nil {
		t.Alias = ctx.Identifier().Accept(s).(string)
	}

	t.Set(ctx)
	return t
}

func (s *schemaVisitor) VisitJoin(ctx *gen.JoinContext) any {
	j := &Join{
		Relation: ctx.Relation().Accept(s).(Table),
		On:       ctx.Sql_expr().Accept(s).(Expression),
	}
	switch {
	case ctx.LEFT() != nil:
		j.Type = JoinTypeLeft
	case ctx.RIGHT() != nil:
		j.Type = JoinTypeRight
	case ctx.FULL() != nil:
		j.Type = JoinTypeFull
	case ctx.INNER() != nil:
		j.Type = JoinTypeInner
	default:
		// default to inner join
		j.Type = JoinTypeInner
	}

	j.Set(ctx)
	return j
}

func (s *schemaVisitor) VisitExpression_result_column(ctx *gen.Expression_result_columnContext) any {
	col := &ResultColumnExpression{
		Expression: ctx.Sql_expr().Accept(s).(Expression),
	}

	if ctx.Identifier() != nil {
		col.Alias = ctx.Identifier().Accept(s).(string)
	}

	col.Set(ctx)
	return col
}

func (s *schemaVisitor) VisitWildcard_result_column(ctx *gen.Wildcard_result_columnContext) any {
	col := &ResultColumnWildcard{}

	if ctx.Identifier() != nil {
		col.Table = ctx.Identifier().Accept(s).(string)
	}

	col.Set(ctx)
	return col
}

func (s *schemaVisitor) VisitUpdate_statement(ctx *gen.Update_statementContext) any {
	up := &UpdateStatement{
		Table:     ctx.GetTable_name().Accept(s).(string),
		SetClause: arr[*UpdateSetClause](len(ctx.AllUpdate_set_clause())),
		Joins:     arr[*Join](len(ctx.AllJoin())),
	}

	if ctx.GetAlias() != nil {
		up.Alias = ctx.GetAlias().Accept(s).(string)
	}

	for i, set := range ctx.AllUpdate_set_clause() {
		up.SetClause[i] = set.Accept(s).(*UpdateSetClause)
	}

	if ctx.Relation() != nil {
		up.From = ctx.Relation().Accept(s).(Table)
	}

	for i, join := range ctx.AllJoin() {
		up.Joins[i] = join.Accept(s).(*Join)
	}

	if ctx.GetWhere() != nil {
		up.Where = ctx.GetWhere().Accept(s).(Expression)
	}

	up.Set(ctx)
	return up
}

func (s *schemaVisitor) VisitUpdate_set_clause(ctx *gen.Update_set_clauseContext) any {
	u := &UpdateSetClause{
		Column: ctx.GetColumn().Accept(s).(string),
		Value:  ctx.Sql_expr().Accept(s).(Expression),
	}

	u.Set(ctx)
	return u
}

func (s *schemaVisitor) VisitInsert_statement(ctx *gen.Insert_statementContext) any {
	ins := &InsertStatement{
		Table: ctx.GetTable_name().Accept(s).(string),
	}

	// can either be INSERT INTO table VALUES (1, 2, 3) or
	// INSERT INTO table SELECT * FROM table2
	if ctx.Select_statement() != nil {
		ins.Select = ctx.Select_statement().Accept(s).(*SelectStatement)
	} else {
		for _, valList := range ctx.AllSql_expr_list() {
			ins.Values = append(ins.Values, valList.Accept(s).([]Expression))
		}
	}

	if ctx.GetAlias() != nil {
		ins.Alias = ctx.GetAlias().Accept(s).(string)
	}

	if ctx.Identifier_list() != nil {
		ins.Columns = ctx.Identifier_list().Accept(s).([]string)
	}

	if ctx.Upsert_clause() != nil {
		ins.OnConflict = ctx.Upsert_clause().Accept(s).(*OnConflict)
	}

	ins.Set(ctx)
	return ins
}

func (s *schemaVisitor) VisitUpsert_clause(ctx *gen.Upsert_clauseContext) any {
	u := &OnConflict{}

	if ctx.GetConflict_columns() != nil {
		u.ConflictColumns = ctx.GetConflict_columns().Accept(s).([]string)
	}

	if ctx.GetConflict_where() != nil {
		u.ConflictWhere = ctx.GetConflict_where().Accept(s).(Expression)
	}

	if ctx.UPDATE() != nil {
		for _, set := range ctx.AllUpdate_set_clause() {
			u.DoUpdate = append(u.DoUpdate, set.Accept(s).(*UpdateSetClause))
		}
	}

	if ctx.GetUpdate_where() != nil {
		u.UpdateWhere = ctx.GetUpdate_where().Accept(s).(Expression)
	}

	u.Set(ctx)
	return u
}

func (s *schemaVisitor) VisitDelete_statement(ctx *gen.Delete_statementContext) any {
	d := &DeleteStatement{
		Table: ctx.GetTable_name().Accept(s).(string),
	}

	if ctx.GetAlias() != nil {
		d.Alias = ctx.GetAlias().Accept(s).(string)
	}

	if ctx.GetWhere() != nil {
		d.Where = ctx.GetWhere().Accept(s).(Expression)
	}

	d.Set(ctx)
	return d
}

func (s *schemaVisitor) VisitColumn_sql_expr(ctx *gen.Column_sql_exprContext) any {
	e := &ExpressionColumn{
		Column: ctx.GetColumn().Accept(s).(string),
	}

	if ctx.GetTable() != nil {
		e.Table = ctx.GetTable().Accept(s).(string)
	}

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitLogical_sql_expr(ctx *gen.Logical_sql_exprContext) any {
	e := &ExpressionLogical{
		Left:  ctx.Sql_expr(0).Accept(s).(Expression),
		Right: ctx.Sql_expr(1).Accept(s).(Expression),
	}

	switch {
	case ctx.AND() != nil:
		e.Operator = LogicalOperatorAnd
	case ctx.OR() != nil:
		e.Operator = LogicalOperatorOr
	default:
		panic("unknown logical operator")
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitArray_access_sql_expr(ctx *gen.Array_access_sql_exprContext) any {
	e := &ExpressionArrayAccess{
		Array: ctx.Sql_expr(0).Accept(s).(Expression),
	}

	s.makeArray(e, ctx.GetSingle(), ctx.GetLeft(), ctx.GetRight())

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}

	e.Set(ctx)
	return e
}

// makeArray modifies the passed ExpressionArrayAccess based on the single, left, and right fields.
// single: arr[1], left: arr[1:], right: arr[:1], left and right: arr[1:2], neither: arr[:]
func (s *schemaVisitor) makeArray(e *ExpressionArrayAccess, single, left, right antlr.ParserRuleContext) {
	if single != nil {
		e.Index = single.Accept(s).(Expression)
		return
	}

	var start, end Expression
	if left != nil {
		start = left.Accept(s).(Expression)
	}
	if right != nil {
		end = right.Accept(s).(Expression)
	}

	e.FromTo = [2]Expression{start, end}
}

func (s *schemaVisitor) VisitField_access_sql_expr(ctx *gen.Field_access_sql_exprContext) any {
	e := &ExpressionFieldAccess{
		Record: ctx.Sql_expr().Accept(s).(Expression),
		Field:  ctx.Identifier().Accept(s).(string),
	}

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitComparison_sql_expr(ctx *gen.Comparison_sql_exprContext) any {
	e := &ExpressionComparison{
		Left:  ctx.Sql_expr(0).Accept(s).(Expression),
		Right: ctx.Sql_expr(1).Accept(s).(Expression),
	}

	switch {
	case ctx.EQUALS() != nil || ctx.EQUATE() != nil:
		e.Operator = ComparisonOperatorEqual
	case ctx.NEQ() != nil:
		e.Operator = ComparisonOperatorNotEqual
	case ctx.LT() != nil:
		e.Operator = ComparisonOperatorLessThan
	case ctx.LTE() != nil:
		e.Operator = ComparisonOperatorLessThanOrEqual
	case ctx.GT() != nil:
		e.Operator = ComparisonOperatorGreaterThan
	case ctx.GTE() != nil:
		e.Operator = ComparisonOperatorGreaterThanOrEqual
	default:
		panic("unknown comparison operator")
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitLiteral_sql_expr(ctx *gen.Literal_sql_exprContext) any {
	e := ctx.Literal().Accept(s).(*ExpressionLiteral)

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitBetween_sql_expr(ctx *gen.Between_sql_exprContext) any {
	e := &ExpressionBetween{
		Expression: ctx.GetElement().Accept(s).(Expression),
		Lower:      ctx.GetLower().Accept(s).(Expression),
		Upper:      ctx.GetUpper().Accept(s).(Expression),
	}

	if ctx.NOT() != nil {
		e.Not = true
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitFunction_call_sql_expr(ctx *gen.Function_call_sql_exprContext) any {
	call := ctx.Sql_function_call().Accept(s).(*ExpressionFunctionCall)

	if ctx.Type_cast() != nil {
		call.Cast(ctx.Type_cast().Accept(s).(*types.DataType))
	}

	call.Set(ctx)

	return call
}

func (s *schemaVisitor) VisitWindow_function_call_sql_expr(ctx *gen.Window_function_call_sql_exprContext) any {
	e := &ExpressionWindowFunctionCall{
		FunctionCall: ctx.Sql_function_call().Accept(s).(*ExpressionFunctionCall),
	}

	if ctx.IDENTIFIER() != nil {
		name := s.getIdent(ctx.IDENTIFIER())
		wr := &WindowReference{
			Name: name,
		}
		wr.SetToken(ctx.IDENTIFIER().GetSymbol())
		e.Window = wr
	} else {
		e.Window = ctx.Window().Accept(s).(*WindowImpl)
	}

	if ctx.FILTER() != nil {
		e.Filter = ctx.Sql_expr().Accept(s).(Expression)
	}

	e.Set(ctx)

	return e
}

func (s *schemaVisitor) VisitParen_sql_expr(ctx *gen.Paren_sql_exprContext) any {
	e := &ExpressionParenthesized{
		Inner: ctx.Sql_expr().Accept(s).(Expression),
	}
	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}
	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitCollate_sql_expr(ctx *gen.Collate_sql_exprContext) any {
	e := &ExpressionCollate{
		Expression: ctx.Sql_expr().Accept(s).(Expression),
		Collation:  ctx.Identifier().Accept(s).(string),
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitVariable_sql_expr(ctx *gen.Variable_sql_exprContext) any {
	e := ctx.Variable().Accept(s).(*ExpressionVariable)

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitIs_sql_expr(ctx *gen.Is_sql_exprContext) any {
	e := &ExpressionIs{
		Left: ctx.GetLeft().Accept(s).(Expression),
	}

	if ctx.NOT() != nil {
		e.Not = true
	}

	if ctx.DISTINCT() != nil {
		e.Distinct = true
	}

	switch {
	case ctx.NULL() != nil:
		e.Right = &ExpressionLiteral{
			Type: types.NullType,
		}
		e.Right.SetToken(ctx.NULL().GetSymbol())
	case ctx.TRUE() != nil:
		e.Right = &ExpressionLiteral{
			Type:  types.BoolType,
			Value: true,
		}
		e.Right.SetToken(ctx.TRUE().GetSymbol())
	case ctx.FALSE() != nil:
		e.Right = &ExpressionLiteral{
			Type:  types.BoolType,
			Value: false,
		}
		e.Right.SetToken(ctx.FALSE().GetSymbol())
	case ctx.GetRight() != nil:
		e.Right = ctx.GetRight().Accept(s).(Expression)
	default:
		panic("unknown right side of IS")
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitLike_sql_expr(ctx *gen.Like_sql_exprContext) any {
	e := &ExpressionStringComparison{
		Left:  ctx.GetLeft().Accept(s).(Expression),
		Right: ctx.GetRight().Accept(s).(Expression),
	}

	if ctx.NOT() != nil {
		e.Not = true
	}

	switch {
	case ctx.LIKE() != nil:
		e.Operator = StringComparisonOperatorLike
	case ctx.ILIKE() != nil:
		e.Operator = StringComparisonOperatorILike
	default:
		panic("unknown string comparison operator")
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitArithmetic_sql_expr(ctx *gen.Arithmetic_sql_exprContext) any {
	e := &ExpressionArithmetic{
		Left:  ctx.GetLeft().Accept(s).(Expression),
		Right: ctx.GetRight().Accept(s).(Expression),
	}

	switch {
	case ctx.PLUS() != nil:
		e.Operator = ArithmeticOperatorAdd
	case ctx.MINUS() != nil:
		e.Operator = ArithmeticOperatorSubtract
	case ctx.STAR() != nil:
		e.Operator = ArithmeticOperatorMultiply
	case ctx.DIV() != nil:
		e.Operator = ArithmeticOperatorDivide
	case ctx.MOD() != nil:
		e.Operator = ArithmeticOperatorModulo
	case ctx.CONCAT() != nil:
		e.Operator = ArithmeticOperatorConcat
	default:
		panic("unknown arithmetic operator")
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitUnary_sql_expr(ctx *gen.Unary_sql_exprContext) any {
	e := &ExpressionUnary{
		Expression: ctx.Sql_expr().Accept(s).(Expression),
	}

	// this is the only known unary right now
	switch {
	case ctx.NOT() != nil:
		e.Operator = UnaryOperatorNot
	case ctx.MINUS() != nil:
		e.Operator = UnaryOperatorNeg
	case ctx.PLUS() != nil:
		e.Operator = UnaryOperatorPos
	default:
		panic("unknown unary operator")
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitSubquery_sql_expr(ctx *gen.Subquery_sql_exprContext) any {
	e := &ExpressionSubquery{
		Subquery: ctx.Select_statement().Accept(s).(*SelectStatement),
	}

	if ctx.NOT() != nil {
		e.Not = true
	}

	if ctx.EXISTS() != nil {
		e.Exists = true
	}

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitCase_expr(ctx *gen.Case_exprContext) any {
	e := &ExpressionCase{}
	if ctx.GetCase_clause() != nil {
		e.Case = ctx.GetCase_clause().Accept(s).(Expression)
	}

	for i := range ctx.AllWhen_then_clause() {
		wt := ctx.AllWhen_then_clause()[i].Accept(s).([2]Expression)

		e.WhenThen = append(e.WhenThen, wt)
	}

	if ctx.GetElse_clause() != nil {
		e.Else = ctx.GetElse_clause().Accept(s).(Expression)
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitWhen_then_clause(ctx *gen.When_then_clauseContext) any {
	when := ctx.Sql_expr(0).Accept(s).(Expression)
	then := ctx.Sql_expr(1).Accept(s).(Expression)
	return [2]Expression{when, then}
}

func (s *schemaVisitor) VisitIn_sql_expr(ctx *gen.In_sql_exprContext) any {
	e := &ExpressionIn{
		Expression: ctx.Sql_expr().Accept(s).(Expression),
	}

	if ctx.NOT() != nil {
		e.Not = true
	}

	if ctx.Sql_expr_list() != nil {
		e.List = ctx.Sql_expr_list().Accept(s).([]Expression)
	} else {
		e.Subquery = ctx.Select_statement().Accept(s).(*SelectStatement)
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitSql_expr_list(ctx *gen.Sql_expr_listContext) any {

	var e []Expression
	for _, e2 := range ctx.AllSql_expr() {
		e = append(e, e2.Accept(s).(Expression))
	}

	return e
}

func (s *schemaVisitor) VisitNormal_call_sql(ctx *gen.Normal_call_sqlContext) any {
	call := &ExpressionFunctionCall{
		Name: ctx.Identifier().Accept(s).(string),
	}

	// function calls can have either of these types for args,
	// or none at all.
	switch {
	case ctx.Sql_expr_list() != nil:
		call.Args = ctx.Sql_expr_list().Accept(s).([]Expression)
		if ctx.DISTINCT() != nil {
			call.Distinct = true
		}
	case ctx.STAR() != nil:
		call.Star = true
	}

	call.Set(ctx)
	return call
}

func (s *schemaVisitor) VisitAction_block(ctx *gen.Action_blockContext) any {
	var stmts []ActionStmt

	for _, stmt := range ctx.AllAction_statement() {
		stmts = append(stmts, stmt.Accept(s).(ActionStmt))
	}

	return stmts
}

func (s *schemaVisitor) VisitSql_action(ctx *gen.Sql_actionContext) any {
	stmt := &ActionStmtSQL{
		SQL: ctx.Sql_statement().Accept(s).(*SQLStatement),
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitLocal_action(ctx *gen.Local_actionContext) any {
	stmt := &ActionStmtActionCall{
		Action: s.getIdent(ctx.IDENTIFIER()),
	}

	if ctx.Procedure_expr_list() != nil {
		stmt.Args = ctx.Procedure_expr_list().Accept(s).([]Expression)
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitExtension_action(ctx *gen.Extension_actionContext) any {
	stmt := &ActionStmtExtensionCall{
		Extension: s.getIdent(ctx.IDENTIFIER(0)),
		Method:    s.getIdent(ctx.IDENTIFIER(1)),
	}

	if ctx.Procedure_expr_list() != nil {
		stmt.Args = ctx.Procedure_expr_list().Accept(s).([]Expression)
	}

	if ctx.Variable_list() != nil {
		varList := ctx.Variable_list().Accept(s).([]*ExpressionVariable)
		for _, v := range varList {
			stmt.Receivers = append(stmt.Receivers, v.String())
		}
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitProcedure_block(ctx *gen.Procedure_blockContext) any {
	var stmts []ProcedureStmt

	for _, stmt := range ctx.AllProc_statement() {
		stmts = append(stmts, stmt.Accept(s).(ProcedureStmt))
	}

	return stmts
}

func (s *schemaVisitor) VisitField_access_procedure_expr(ctx *gen.Field_access_procedure_exprContext) any {
	e := &ExpressionFieldAccess{
		Record: ctx.Procedure_expr().Accept(s).(Expression),
		Field:  s.getIdent(ctx.IDENTIFIER()),
	}

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}

	e.Set(ctx)

	return e
}

func (s *schemaVisitor) VisitLiteral_procedure_expr(ctx *gen.Literal_procedure_exprContext) any {
	e := ctx.Literal().Accept(s).(*ExpressionLiteral)

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitParen_procedure_expr(ctx *gen.Paren_procedure_exprContext) any {
	e := &ExpressionParenthesized{
		Inner: ctx.Procedure_expr().Accept(s).(Expression),
	}

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitVariable_procedure_expr(ctx *gen.Variable_procedure_exprContext) any {
	e := ctx.Variable().Accept(s).(*ExpressionVariable)

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitIs_procedure_expr(ctx *gen.Is_procedure_exprContext) any {
	e := &ExpressionIs{
		Left: ctx.Procedure_expr(0).Accept(s).(Expression),
	}

	if ctx.NOT() != nil {
		e.Not = true
	}

	if ctx.DISTINCT() != nil {
		e.Distinct = true
	}

	switch {
	case ctx.NULL() != nil:
		e.Right = &ExpressionLiteral{
			Type: types.NullType,
		}
		e.Right.SetToken(ctx.NULL().GetSymbol())
	case ctx.TRUE() != nil:
		e.Right = &ExpressionLiteral{
			Type:  types.BoolType,
			Value: true,
		}
		e.Right.SetToken(ctx.TRUE().GetSymbol())
	case ctx.FALSE() != nil:
		e.Right = &ExpressionLiteral{
			Type:  types.BoolType,
			Value: false,
		}
		e.Right.SetToken(ctx.FALSE().GetSymbol())
	case ctx.GetRight() != nil:
		e.Right = ctx.GetRight().Accept(s).(Expression)
	default:
		panic("unknown right side of IS")
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitLogical_procedure_expr(ctx *gen.Logical_procedure_exprContext) any {
	e := &ExpressionLogical{
		Left:  ctx.Procedure_expr(0).Accept(s).(Expression),
		Right: ctx.Procedure_expr(1).Accept(s).(Expression),
	}

	switch {
	case ctx.AND() != nil:
		e.Operator = LogicalOperatorAnd
	case ctx.OR() != nil:
		e.Operator = LogicalOperatorOr
	default:
		panic("unknown logical operator")
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitProcedure_expr_arithmetic(ctx *gen.Procedure_expr_arithmeticContext) any {
	e := &ExpressionArithmetic{
		Left:  ctx.Procedure_expr(0).Accept(s).(Expression),
		Right: ctx.Procedure_expr(1).Accept(s).(Expression),
	}

	switch {
	case ctx.PLUS() != nil:
		e.Operator = ArithmeticOperatorAdd
	case ctx.MINUS() != nil:
		e.Operator = ArithmeticOperatorSubtract
	case ctx.STAR() != nil:
		e.Operator = ArithmeticOperatorMultiply
	case ctx.DIV() != nil:
		e.Operator = ArithmeticOperatorDivide
	case ctx.MOD() != nil:
		e.Operator = ArithmeticOperatorModulo
	case ctx.CONCAT() != nil:
		e.Operator = ArithmeticOperatorConcat
	default:
		panic("unknown arithmetic operator")
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitComparison_procedure_expr(ctx *gen.Comparison_procedure_exprContext) any {
	e := &ExpressionComparison{
		Left:  ctx.Procedure_expr(0).Accept(s).(Expression),
		Right: ctx.Procedure_expr(1).Accept(s).(Expression),
	}

	switch {
	case ctx.EQUALS() != nil || ctx.EQUATE() != nil:
		e.Operator = ComparisonOperatorEqual
	case ctx.NEQ() != nil:
		e.Operator = ComparisonOperatorNotEqual
	case ctx.LT() != nil:
		e.Operator = ComparisonOperatorLessThan
	case ctx.LTE() != nil:
		e.Operator = ComparisonOperatorLessThanOrEqual
	case ctx.GT() != nil:
		e.Operator = ComparisonOperatorGreaterThan
	case ctx.GTE() != nil:
		e.Operator = ComparisonOperatorGreaterThanOrEqual
	default:
		panic("unknown comparison operator")
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitFunction_call_procedure_expr(ctx *gen.Function_call_procedure_exprContext) any {
	call := ctx.Procedure_function_call().Accept(s).(*ExpressionFunctionCall)

	if ctx.Type_cast() != nil {
		call.Cast(ctx.Type_cast().Accept(s).(*types.DataType))
	}

	call.Set(ctx)

	return call
}

func (s *schemaVisitor) VisitArray_access_procedure_expr(ctx *gen.Array_access_procedure_exprContext) any {
	e := &ExpressionArrayAccess{
		Array: ctx.Procedure_expr(0).Accept(s).(Expression),
	}

	s.makeArray(e, ctx.GetSingle(), ctx.GetLeft(), ctx.GetRight())

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitUnary_procedure_expr(ctx *gen.Unary_procedure_exprContext) any {
	e := &ExpressionUnary{
		Expression: ctx.Procedure_expr().Accept(s).(Expression),
	}

	// this is the only known unary right now
	switch {
	case ctx.EXCL() != nil || ctx.NOT() != nil:
		e.Operator = UnaryOperatorNot
	case ctx.MINUS() != nil:
		e.Operator = UnaryOperatorNeg
	case ctx.PLUS() != nil:
		e.Operator = UnaryOperatorPos
	default:
		panic("unknown unary operator")
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitMake_array_procedure_expr(ctx *gen.Make_array_procedure_exprContext) any {
	e := &ExpressionMakeArray{
		// golang interface assertions do not work for slices, so we simply
		// cast the result to []Expression. This comes from VisitProcedure_expr_list,
		// directly below.
	}

	// we could enforce this in the parser, but it is not super intuitive,
	// so we want to control the error message
	if ctx.Procedure_expr_list() == nil {
		s.errs.RuleErr(ctx, ErrSyntax, "cannot assign empty arrays. declare using `$arr type[];` instead`")
	}
	e.Values = ctx.Procedure_expr_list().Accept(s).([]Expression)

	if ctx.Type_cast() != nil {
		e.TypeCast = ctx.Type_cast().Accept(s).(*types.DataType)
	}

	e.Set(ctx)
	return e
}

func (s *schemaVisitor) VisitProcedure_expr_list(ctx *gen.Procedure_expr_listContext) any {
	// we do not return a type ExpressionList here, since ExpressionList is SQL specific,
	// and not supported in procedures. Instead, we return a slice of Expression.
	var exprs []Expression

	for _, e := range ctx.AllProcedure_expr() {
		exprs = append(exprs, e.Accept(s).(Expression))
	}

	return exprs
}

func (s *schemaVisitor) VisitStmt_variable_declaration(ctx *gen.Stmt_variable_declarationContext) any {
	stmt := &ProcedureStmtDeclaration{
		Variable: varFromTerminalNode(ctx.VARIABLE()),
	}

	if ctx.Type_() != nil {
		stmt.Type = ctx.Type_().Accept(s).(*types.DataType)
	}

	stmt.Set(ctx)
	return stmt
}

// varFromTerminalNode returns a variable from an antlr terminal node.
// If the string is not a valid variable, it panics.
// This is only meant to be used when creating expressions
// directly from the AST. It will lowercase the string to
// ensure case-insensitivity.
func varFromTerminalNode(node antlr.TerminalNode) *ExpressionVariable {
	s := node.GetText()
	e := varFromString(s)
	e.SetToken(node.GetSymbol())

	return e
}

// varFromString returns a variable from a string.
func varFromString(s string) *ExpressionVariable {
	e := &ExpressionVariable{}
	if len(s) < 2 {
		panic("invalid variable: " + s)
	}

	switch {
	case s[0] == '$':
		e.Prefix = VariablePrefixDollar
	case s[0] == '@':
		e.Prefix = VariablePrefixAt
	default:
		panic("invalid variable: " + s)
	}

	e.Name = strings.ToLower(s)

	return e
}

func (s *schemaVisitor) VisitStmt_procedure_call(ctx *gen.Stmt_procedure_callContext) any {
	stmt := &ProcedureStmtCall{
		Call: ctx.Procedure_function_call().Accept(s).(*ExpressionFunctionCall),
	}

	for i, v := range ctx.AllVariable_or_underscore() {
		// check for nil since nil pointer will fail *string type assertion
		if v.Accept(s) == nil {
			stmt.Receivers = append(stmt.Receivers, nil)
			continue
		}

		stmt.Receivers = append(stmt.Receivers, varFromString(*v.Accept(s).(*string)))
		stmt.Receivers[i].Set(v)
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitVariable_or_underscore(ctx *gen.Variable_or_underscoreContext) any {
	if ctx.UNDERSCORE() != nil {
		return nil
	}

	str := s.getIdent(ctx.VARIABLE())
	return &str
}

func (s *schemaVisitor) VisitStmt_variable_assignment(ctx *gen.Stmt_variable_assignmentContext) any {
	stmt := &ProcedureStmtAssign{}

	assignVariable := ctx.Procedure_expr(0).Accept(s).(Expression)

	assignable, ok := assignVariable.(Assignable)
	if !ok {
		s.errs.RuleErr(ctx.Procedure_expr(0), ErrSyntax, "cannot assign to %T", assignVariable)
	}
	stmt.Variable = assignable
	stmt.Value = ctx.Procedure_expr(1).Accept(s).(Expression)

	if ctx.Type_() != nil {
		stmt.Type = ctx.Type_().Accept(s).(*types.DataType)
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitStmt_for_loop(ctx *gen.Stmt_for_loopContext) any {
	stmt := &ProcedureStmtForLoop{
		Receiver: varFromTerminalNode(ctx.VARIABLE()),
		Body:     arr[ProcedureStmt](len(ctx.AllProc_statement())),
	}

	switch {
	case ctx.Range_() != nil:
		stmt.LoopTerm = ctx.Range_().Accept(s).(*LoopTermRange)
	case ctx.GetTarget_variable() != nil:
		v := ctx.GetTarget_variable().Accept(s).(*ExpressionVariable)
		stmt.LoopTerm = &LoopTermVariable{
			Variable: v,
		}

		if v.GetTypeCast() != nil {
			s.errs.RuleErr(ctx.GetTarget_variable(), ErrSyntax, "cannot typecast loop variable")
		}

		stmt.LoopTerm.Set(ctx.GetTarget_variable())
	case ctx.Sql_statement() != nil:
		sqlStmt := ctx.Sql_statement().Accept(s).(*SQLStatement)
		stmt.LoopTerm = &LoopTermSQL{
			Statement: sqlStmt,
		}
		stmt.LoopTerm.Set(ctx.Sql_statement())
	default:
		panic("unknown loop term")
	}

	for i, st := range ctx.AllProc_statement() {
		stmt.Body[i] = st.Accept(s).(ProcedureStmt)
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitStmt_if(ctx *gen.Stmt_ifContext) any {
	stmt := &ProcedureStmtIf{
		IfThens: arr[*IfThen](len(ctx.AllIf_then_block())),
		Else:    arr[ProcedureStmt](len(ctx.AllProc_statement())),
	}

	for i, th := range ctx.AllIf_then_block() {
		stmt.IfThens[i] = th.Accept(s).(*IfThen)
	}

	for i, st := range ctx.AllProc_statement() {
		stmt.Else[i] = st.Accept(s).(ProcedureStmt)
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitIf_then_block(ctx *gen.If_then_blockContext) any {
	ifthen := &IfThen{
		If:   ctx.Procedure_expr().Accept(s).(Expression),
		Then: arr[ProcedureStmt](len(ctx.AllProc_statement())),
	}

	for i, st := range ctx.AllProc_statement() {
		ifthen.Then[i] = st.Accept(s).(ProcedureStmt)
	}

	ifthen.Set(ctx)

	return ifthen
}

func (s *schemaVisitor) VisitStmt_sql(ctx *gen.Stmt_sqlContext) any {
	stmt := &ProcedureStmtSQL{
		SQL: ctx.Sql_statement().Accept(s).(*SQLStatement),
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitStmt_break(ctx *gen.Stmt_breakContext) any {
	stmt := &ProcedureStmtBreak{}
	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitStmt_return(ctx *gen.Stmt_returnContext) any {
	stmt := &ProcedureStmtReturn{}

	switch {
	case ctx.Sql_statement() != nil:
		stmt.SQL = ctx.Sql_statement().Accept(s).(*SQLStatement)
	case ctx.Procedure_expr_list() != nil:
		// loop through and add since these are Expressions, not Expressions
		exprs := ctx.Procedure_expr_list().Accept(s).([]Expression)
		stmt.Values = append(stmt.Values, exprs...)
		// return can be nil if a procedure simply wants to exit early
	}

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitStmt_return_next(ctx *gen.Stmt_return_nextContext) any {
	stmt := &ProcedureStmtReturnNext{}

	vals := ctx.Procedure_expr_list().Accept(s).([]Expression)
	stmt.Values = append(stmt.Values, vals...)

	stmt.Set(ctx)
	return stmt
}

func (s *schemaVisitor) VisitNormal_call_procedure(ctx *gen.Normal_call_procedureContext) any {
	call := &ExpressionFunctionCall{
		Name: s.getIdent(ctx.IDENTIFIER()),
	}

	// distinct and * cannot be used in procedure function calls
	if ctx.Procedure_expr_list() != nil {
		call.Args = ctx.Procedure_expr_list().Accept(s).([]Expression)
	}

	call.Set(ctx)
	return call
}

func (s *schemaVisitor) VisitRange(ctx *gen.RangeContext) any {
	r := &LoopTermRange{
		Start: ctx.Procedure_expr(0).Accept(s).(Expression),
		End:   ctx.Procedure_expr(1).Accept(s).(Expression),
	}

	r.Set(ctx)
	return r
}

func (s *schemaVisitor) VisitWindow(ctx *gen.WindowContext) any {
	win := &WindowImpl{}

	if ctx.GetPartition() != nil {
		win.PartitionBy = ctx.GetPartition().Accept(s).([]Expression)
	}

	if ctx.ORDER() != nil {
		for _, o := range ctx.AllOrdering_term() {
			win.OrderBy = append(win.OrderBy, o.Accept(s).(*OrderingTerm))
		}
	}

	// currently does not support frame

	win.Set(ctx)
	return win
}

func (s *schemaVisitor) Visit(tree antlr.ParseTree) interface {
} {
	return tree.Accept(s)
}

// getIdent returns the text of an identifier.
// It checks that the identifier is not too long.
// It also converts the identifier to lowercase.
func (s *schemaVisitor) getIdent(i antlr.TerminalNode) string {
	ident := i.GetText()
	s.validateVariableIdentifier(i.GetSymbol(), ident)
	return strings.ToLower(ident)
}

// validateVariableIdentifier validates that a variable's identifier is not too long.
// It doesn't check if it is a keyword, since variables have $ prefixes.
func (s *schemaVisitor) validateVariableIdentifier(i antlr.Token, str string) {
	if len(str) > validation.MAX_IDENT_NAME_LENGTH {
		s.errs.TokenErr(i, ErrIdentifier, "maximum identifier length is %d", maxIdentifierLength)
	}
}

// pg max is 63, but Kwil sometimes adds extra characters
var maxIdentifierLength = 32
