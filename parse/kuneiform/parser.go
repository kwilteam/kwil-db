package kuneiform

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/kuneiform/gen"
	parseTypes "github.com/kwilteam/kwil-db/parse/types"
)

// Parse parses a Kuneiform file and returns the parsed schema.
// It will also parse the SQL to perform validity checks.
func Parse(kf string) (schema *types.Schema, info *parseTypes.SchemaInfo, parseErrs parseTypes.ParseErrors, err error) {
	errorListener := parseTypes.NewErrorListener()

	visitor := &kfVisitor{
		registeredNames: make(map[string]struct{}),
		schemaInfo:      &parseTypes.SchemaInfo{Blocks: make(map[string]*parseTypes.Block)},
		errs:            errorListener,
	}

	stream := antlr.NewInputStream(kf)
	lexer := gen.NewKuneiformLexer(stream)

	// remove default error listener
	lexer.RemoveErrorListeners()
	lexer.AddErrorListener(errorListener)

	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := gen.NewKuneiformParser(tokenStream)

	// remove default error listener
	p.RemoveErrorListeners()
	p.AddErrorListener(errorListener)

	p.BuildParseTrees = true

	defer func() {
		if e := recover(); e != nil {
			var ok bool
			err, ok = e.(error)
			if !ok {
				err = fmt.Errorf("panic: %v", e)
			}
		}
	}()

	res, ok := p.Program().Accept(visitor).(*types.Schema)
	if !ok {
		return nil, nil, nil, fmt.Errorf("unexpected result type: %T", res)
	}

	return res, visitor.schemaInfo, errorListener.Errs, nil
}

type kfVisitor struct {
	*gen.BaseKuneiformParserVisitor
	// registeredNames tracks the names of all tables, columns, and indexes
	registeredNames map[string]struct{}
	schemaInfo      *parseTypes.SchemaInfo
	errs            parseTypes.AntlrErrorListener
}

// registerBlock registers a new block (table, action, procedure, etc.)
// it checks that the name is unique and panics if it is not.
func (k *kfVisitor) registerBlock(ctx antlr.ParserRuleContext, name string) {
	lower := strings.ToLower(name)
	if _, ok := k.registeredNames[lower]; ok {
		k.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSemantic, "conflicting name: "+name)
		return
	}

	k.registeredNames[lower] = struct{}{}

	node := &parseTypes.Node{}
	node.Set(ctx)
	k.schemaInfo.Blocks[lower] = &parseTypes.Block{
		Node:     *node,
		AbsStart: ctx.GetStart().GetStart(),
		AbsEnd:   ctx.GetStop().GetStop(),
	}
}

var _ gen.KuneiformParserVisitor = &kfVisitor{}

func (k *kfVisitor) VisitProgram(ctx *gen.ProgramContext) any {
	schema := &types.Schema{
		Name:              ctx.Database_declaration().Accept(k).(string),
		Tables:            arr[*types.Table](len(ctx.AllTable_declaration())),
		Extensions:        arr[*types.Extension](len(ctx.AllUse_declaration())),
		ForeignProcedures: arr[*types.ForeignProcedure](len(ctx.AllForeign_declaration())),
	}

	for i, tbl := range ctx.AllTable_declaration() {
		schema.Tables[i] = tbl.Accept(k).(*types.Table)
		k.registeredNames[schema.Tables[i].Name] = struct{}{}
	}

	for i, ext := range ctx.AllUse_declaration() {
		schema.Extensions[i] = ext.Accept(k).(*types.Extension)
	}

	for _, stmt := range ctx.AllStmt_mode() {
		switch {
		case stmt.Action_declaration() != nil:
			act := stmt.Accept(k).(*types.Action)
			schema.Actions = append(schema.Actions, act)
		case stmt.Procedure_declaration() != nil:
			proc := stmt.Accept(k).(*types.Procedure)
			schema.Procedures = append(schema.Procedures, proc)
		default:
			panic("unexpected stmt mode")
		}
	}

	for i, foreign := range ctx.AllForeign_declaration() {
		schema.ForeignProcedures[i] = foreign.Accept(k).(*types.ForeignProcedure)
	}

	return schema
}

// START OF Table Declaration Parsing:

func (k *kfVisitor) VisitTable_declaration(ctx *gen.Table_declarationContext) any {
	tbl := &types.Table{
		Name:        ctx.IDENTIFIER().GetText(),
		Columns:     arr[*types.Column](len(ctx.AllColumn_def())),
		Indexes:     arr[*types.Index](len(ctx.AllIndex_def())),
		ForeignKeys: arr[*types.ForeignKey](len(ctx.AllForeign_key_def())),
	}

	k.registerBlock(ctx, tbl.Name)

	for i, col := range ctx.AllColumn_def() {
		tbl.Columns[i] = col.Accept(k).(*types.Column)
	}

	for i, idx := range ctx.AllIndex_def() {
		tbl.Indexes[i] = idx.Accept(k).(*types.Index)
	}

	for i, fk := range ctx.AllForeign_key_def() {
		tbl.ForeignKeys[i] = fk.Accept(k).(*types.ForeignKey)
	}

	return tbl
}

func (k *kfVisitor) VisitIndex_def(ctx *gen.Index_defContext) any {
	idx := &types.Index{
		Name: ctx.INDEX_NAME().GetText()[1:], // trim off the leading #
	}

	switch {
	case ctx.PRIMARY() != nil:
		idx.Type = types.PRIMARY
	case ctx.UNIQUE() != nil:
		idx.Type = types.UNIQUE_BTREE
	case ctx.INDEX() != nil:
		idx.Type = types.BTREE
	default:
		panic("unexpected index type")
	}

	idx.Columns = ctx.GetColumns().Accept(k).([]string)

	return idx
}

func (k *kfVisitor) VisitForeign_key_def(ctx *gen.Foreign_key_defContext) any {
	fk := &types.ForeignKey{
		ChildKeys:   ctx.GetChild_keys().Accept(k).([]string),
		ParentKeys:  ctx.GetParent_keys().Accept(k).([]string),
		ParentTable: ctx.GetParent_table().GetText(),
		Actions:     arr[*types.ForeignKeyAction](len(ctx.AllForeign_key_action())),
	}

	for i, act := range ctx.AllForeign_key_action() {
		fk.Actions[i] = act.Accept(k).(*types.ForeignKeyAction)
	}

	return fk
}

func (k *kfVisitor) VisitForeign_key_action(ctx *gen.Foreign_key_actionContext) any {
	act := &types.ForeignKeyAction{}

	switch {
	case ctx.ON_UPDATE() != nil:
		act.On = types.ON_UPDATE
	case ctx.ON_DELETE() != nil:
		act.On = types.ON_DELETE
	default:
		panic("unexpected foreign key action on type")
	}

	switch {
	case ctx.DO_CASCADE() != nil:
		act.Do = types.DO_CASCADE
	case ctx.DO_SET_NULL() != nil:
		act.Do = types.DO_SET_NULL
	case ctx.DO_SET_DEFAULT() != nil:
		act.Do = types.DO_SET_DEFAULT
	case ctx.DO_NO_ACTION() != nil:
		act.Do = types.DO_NO_ACTION
	case ctx.DO_RESTRICT() != nil:
		act.Do = types.DO_RESTRICT
	default:
		panic("unexpected foreign key action do type")
	}

	return act
}

func (k *kfVisitor) VisitColumn_def(ctx *gen.Column_defContext) any {
	col := &types.Column{
		Name: ctx.IDENTIFIER().GetText(),
		Type: ctx.GetType_().Accept(k).(*types.DataType),
	}

	constraints := arr[*types.Attribute](len(ctx.AllConstraint()))
	for i, c := range ctx.AllConstraint() {
		constraints[i] = c.Accept(k).(*types.Attribute)
	}

	col.Attributes = constraints

	return col
}

func (k *kfVisitor) VisitDEFAULT(ctx *gen.DEFAULTContext) any {
	val := ctx.Literal().Accept(k).(string)

	return &types.Attribute{
		Type:  types.DEFAULT,
		Value: val,
	}
}

func (k *kfVisitor) VisitMAX(ctx *gen.MAXContext) any {
	return &types.Attribute{
		Type:  types.MAX,
		Value: ctx.NUMERIC_LITERAL().GetText(),
	}
}

func (k *kfVisitor) VisitMAX_LEN(ctx *gen.MAX_LENContext) any {
	return &types.Attribute{
		Type:  types.MAX_LENGTH,
		Value: ctx.NUMERIC_LITERAL().GetText(),
	}
}

func (k *kfVisitor) VisitMIN(ctx *gen.MINContext) any {
	return &types.Attribute{
		Type:  types.MIN,
		Value: ctx.NUMERIC_LITERAL().GetText(),
	}
}

func (k *kfVisitor) VisitMIN_LEN(ctx *gen.MIN_LENContext) any {
	return &types.Attribute{
		Type:  types.MIN_LENGTH,
		Value: ctx.NUMERIC_LITERAL().GetText(),
	}
}

func (k *kfVisitor) VisitNOT_NULL(ctx *gen.NOT_NULLContext) any {
	return &types.Attribute{
		Type: types.NOT_NULL,
	}
}

func (k *kfVisitor) VisitPRIMARY_KEY(ctx *gen.PRIMARY_KEYContext) any {
	return &types.Attribute{
		Type: types.PRIMARY_KEY,
	}
}

func (k *kfVisitor) VisitUNIQUE(ctx *gen.UNIQUEContext) any {
	return &types.Attribute{
		Type: types.UNIQUE,
	}
}

// END OF Table Declaration Parsing

func (k *kfVisitor) VisitDatabase_declaration(ctx *gen.Database_declarationContext) any {
	return ctx.IDENTIFIER().GetText()
}

func (k *kfVisitor) VisitIdentifier_list(ctx *gen.Identifier_listContext) any {
	idents := arr[string](len(ctx.AllIDENTIFIER()))
	for i, ident := range ctx.AllIDENTIFIER() {
		idents[i] = ident.GetText()
	}

	return idents
}

func (k *kfVisitor) VisitLiteral(ctx *gen.LiteralContext) any {
	switch {
	case ctx.NUMERIC_LITERAL() != nil:
		return ctx.NUMERIC_LITERAL().GetText()
	case ctx.TEXT_LITERAL() != nil:
		// we do not trim the quotes, and will leeave it to
		// the backend to handle it
		return ctx.TEXT_LITERAL().GetText()
	case ctx.BOOLEAN_LITERAL() != nil:
		return ctx.BOOLEAN_LITERAL().GetText()
	case ctx.BLOB_LITERAL() != nil:
		return ctx.BLOB_LITERAL().GetText()
	default:
		panic("unexpected literal")
	}
}

func (k *kfVisitor) VisitType_selector(ctx *gen.Type_selectorContext) any {
	c := &types.DataType{}

	c.Name = ctx.GetType_().GetText()

	if ctx.LBRACKET() != nil {
		c.IsArray = true
	}

	err := c.Clean()
	if err != nil {
		panic(err)
	}

	return c
}

func (k *kfVisitor) VisitUse_declaration(ctx *gen.Use_declarationContext) any {
	c := &types.Extension{}
	for i, ident := range ctx.AllIDENTIFIER() {
		// the first identifier is the extension name,
		// the last identifier is the alias name,
		// all of the in-between identifiers are keys
		// for the extension config
		if i == 0 {
			c.Name = ident.GetText()
			continue
		}
		if i == len(ctx.AllIDENTIFIER())-1 {
			c.Alias = ident.GetText()
			k.registerBlock(ctx, c.Alias)
			continue
		}

		value := ctx.Literal(i - 1).Accept(k).(string)
		c.Initialization = append(c.Initialization, &types.ExtensionConfig{
			Key:   ident.GetText(),
			Value: value,
		})
	}

	return c
}

// action/procedure parsing:

// returns either *types.Action or *types.Procedure
func (k *kfVisitor) VisitStmt_mode(ctx *gen.Stmt_modeContext) any {
	annotations := arr[string](len(ctx.AllANNOTATION()))
	for i, a := range ctx.AllANNOTATION() {
		annotations[i] = a.GetText()
	}

	switch {
	case ctx.Action_declaration() != nil:
		act := ctx.Action_declaration().Accept(k).(*types.Action)
		act.Annotations = annotations

		return act
	case ctx.Procedure_declaration() != nil:
		proc := ctx.Procedure_declaration().Accept(k).(*types.Procedure)
		proc.Annotations = annotations

		return proc
	default:
		panic("unexpected stmt mode")
	}
}

// returns *types.Action
func (k *kfVisitor) VisitAction_declaration(ctx *gen.Action_declarationContext) any {
	name := ctx.STMT_IDENTIFIER().GetText()

	k.registerBlock(ctx, name)

	act := &types.Action{
		Name:        name,
		Annotations: arr[string](0),
		Parameters:  arr[string](len(ctx.AllSTMT_VAR())),
		Body:        parseBody(ctx.STMT_BODY().GetText()),
	}

	for i, v := range ctx.AllSTMT_VAR() {
		act.Parameters[i] = v.GetText()
	}

	var hasPubOrPriv bool
	act.Modifiers, act.Public, hasPubOrPriv = k.getAccessModifiers(ctx.AllSTMT_ACCESS())
	if !hasPubOrPriv {
		k.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSemantic, "missing public or private modifier")
	}

	return act
}

func (k *kfVisitor) VisitProcedure_declaration(ctx *gen.Procedure_declarationContext) any {
	name := ctx.GetProcedure_name().GetText()

	k.registerBlock(ctx, name)

	proc := &types.Procedure{
		Name:        name,
		Annotations: arr[string](0),
		Body:        parseBody(ctx.STMT_BODY().GetText()),
	}

	if ctx.Stmt_typed_param_list() != nil {
		proc.Parameters = ctx.Stmt_typed_param_list().Accept(k).([]*types.ProcedureParameter)
	}

	var hasPubOrPriv bool
	proc.Modifiers, proc.Public, hasPubOrPriv = k.getAccessModifiers(ctx.AllSTMT_ACCESS())
	if !hasPubOrPriv {
		k.errs.RuleErr(ctx, parseTypes.ParseErrorTypeSemantic, "missing public or private modifier")
	}

	if ctx.Stmt_return() != nil {
		proc.Returns = ctx.Stmt_return().Accept(k).(*types.ProcedureReturn)
	}

	return proc
}

func (k *kfVisitor) VisitStmt_return(ctx *gen.Stmt_returnContext) any {
	r := &types.ProcedureReturn{
		IsTable: ctx.STMT_TABLE() != nil,
		Fields:  arr[*types.NamedType](len(ctx.AllSTMT_IDENTIFIER())),
	}

	for i, c := range ctx.AllSTMT_IDENTIFIER() {
		r.Fields[i] = &types.NamedType{
			Name: c.GetText(),
			Type: ctx.Stmt_type_selector(i).Accept(k).(*types.DataType),
		}
	}

	return r
}

// parseBody will parse the body of a procedure or action, removing
// leading and traily braces and whitespace.
func parseBody(b string) string {
	b = strings.TrimPrefix(b, "{")
	b = strings.TrimSuffix(b, "}")
	return b
}

func (k *kfVisitor) VisitStmt_typed_param_list(ctx *gen.Stmt_typed_param_listContext) any {
	params := arr[*types.ProcedureParameter](len(ctx.AllSTMT_VAR()))

	for i, v := range ctx.AllSTMT_VAR() {
		params[i] = &types.ProcedureParameter{
			Name: v.GetText(),
			Type: ctx.Stmt_type_selector(i).Accept(k).(*types.DataType),
		}
	}

	return params
}

func (k *kfVisitor) VisitStmt_type_selector(ctx *gen.Stmt_type_selectorContext) any {
	c := &types.DataType{}

	c.Name = ctx.GetType_().GetText()

	if ctx.STMT_ARRAY() != nil {
		c.IsArray = true
	}

	err := c.Clean()
	if err != nil {
		panic(err)
	}

	return c
}

func (k *kfVisitor) VisitForeign_declaration(ctx *gen.Foreign_declarationContext) interface {
} {
	foreign := &types.ForeignProcedure{
		Name: ctx.GetName().GetText(),
	}

	// can either be unnamed_params or named_params.
	// regardless, we can throw away the name; it is just for consistency wit other
	// parts of the language.
	if p := ctx.GetUnnamed_params(); p != nil {
		foreign.Parameters = p.Accept(k).([]*types.DataType)
	} else if p := ctx.GetNamed_params(); p != nil {
		named := p.Accept(k).([]*types.NamedType)
		foreign.Parameters = make([]*types.DataType, len(named))
		for i, n := range named {
			foreign.Parameters[i] = n.Type
		}
	}

	// there can be three cases for the return:
	// 1. "returns table(id int, name text)"
	// 2. "returns (id int, name text)"
	// 3. "returns (int, text)"

	if ctx.GetReturn_columns() != nil {
		// this will be not nil if either 1 or 2 is chosen
		foreign.Returns = &types.ProcedureReturn{
			IsTable: ctx.TABLE() != nil,
			Fields:  ctx.GetReturn_columns().Accept(k).([]*types.NamedType),
		}
	} else if ctx.GetUnnamed_return_types() != nil {
		// this will be not nil if 3 is chosen
		dataTypes := ctx.GetUnnamed_return_types().Accept(k).([]*types.DataType)
		namedTypes := make([]*types.NamedType, len(dataTypes))
		for i, dt := range dataTypes {
			namedTypes[i] = &types.NamedType{
				Name: fmt.Sprintf("col%d", i), // this gets thrown away anyways in this case
				Type: dt,
			}
		}

		foreign.Returns = &types.ProcedureReturn{
			IsTable: false,
			Fields:  namedTypes,
		}
	}

	return foreign
}

func (k *kfVisitor) VisitTyped_variable_list(ctx *gen.Typed_variable_listContext) interface {
} {
	typs := arr[*types.NamedType](len(ctx.AllType_selector()))
	for i, ts := range ctx.AllVAR() {
		typs[i] = &types.NamedType{
			Name: ts.GetText(),
			Type: ctx.Type_selector(i).Accept(k).(*types.DataType),
		}
	}

	return typs
}

func (k *kfVisitor) VisitNamed_type_list(ctx *gen.Named_type_listContext) interface {
} {
	typs := arr[*types.NamedType](len(ctx.AllIDENTIFIER()))
	for i, ident := range ctx.AllIDENTIFIER() {
		typs[i] = &types.NamedType{
			Name: ident.GetText(),
			Type: ctx.Type_selector(i).Accept(k).(*types.DataType),
		}
	}

	return typs
}

func (k *kfVisitor) VisitType_selector_list(ctx *gen.Type_selector_listContext) interface {
} {
	typs := arr[*types.DataType](len(ctx.AllType_selector()))
	for i, ts := range ctx.AllType_selector() {
		typs[i] = ts.Accept(k).(*types.DataType)
	}

	return typs
}

// getAccessModifiers returns the access modifiers for the given context.
// It also returns whether or not the action/procedure is public or private.
// If it can't find public or private, it will return false
func (k *kfVisitor) getAccessModifiers(mods []antlr.TerminalNode) (modifiers []types.Modifier, public bool, foundPublicOrPrivate bool) {
	for _, m := range mods {
		switch strings.ToLower(m.GetText()) {
		case "public":
			public = true
			foundPublicOrPrivate = true
		case "private":
			public = false
			foundPublicOrPrivate = true
		case "view":
			modifiers = append(modifiers, types.ModifierView)
		case "owner":
			modifiers = append(modifiers, types.ModifierOwner)
		default:
			m.GetSymbol()
			k.errs.TokenErr(m.GetSymbol(), parseTypes.ParseErrorTypeSemantic, "unexpected modifier: "+m.GetText())
			return
		}
	}

	return
}

// arr will make an array of type A if the input is greater than 0
func arr[A any](b int) []A {
	if b > 0 {
		return make([]A, b)
	}
	return nil
}
