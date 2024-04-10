package kuneiform

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/kwilteam/kwil-db/core/types"
	sqlparser "github.com/kwilteam/kwil-db/internal/parse/sql"
	"github.com/kwilteam/kwil-db/kuneiform/gen"
)

// Parse parses a Kuneiform file and returns the parsed schema.
// It will also parse the SQL to perform validity checks.
func Parse(kf string) (schema *types.Schema, err error) {
	visitor := &kfVisitor{
		registeredNames: make(map[string]struct{}),
	}

	errorListener := sqlparser.NewErrorListener()

	stream := antlr.NewInputStream(kf)
	lexer := gen.NewKuneiformLexer(stream)
	tokenStream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := gen.NewKuneiformParser(tokenStream)

	if errorListener != nil {
		// remove default error visitor
		p.RemoveErrorListeners()
		p.AddErrorListener(errorListener)
	}

	p.BuildParseTrees = true

	defer func() {
		if e := recover(); e != nil {
			errorListener.Add(fmt.Sprintf("panic: %v", e))
		}

		if err != nil {
			errorListener.AddError(err)
		}

		err = errorListener.Err()
	}()

	res, ok := p.Program().Accept(visitor).(*types.Schema)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", res)
	}

	if errorListener.Err() != nil {
		return nil, errorListener.Err()
	}

	return res, nil
}

type kfVisitor struct {
	*gen.BaseKuneiformParserVisitor
	// registeredNames tracks the names of all tables, columns, and indexes
	registeredNames map[string]struct{}
}

// checkUniqueName checks if the name is unique.
// If it is not, it will panic.
func (k *kfVisitor) checkUniqueName(name string) {
	lower := strings.ToLower(name)
	if _, ok := k.registeredNames[lower]; ok {
		panic(fmt.Sprintf("conflicting name: %s", name))
	}

	k.registeredNames[lower] = struct{}{}
}

var _ gen.KuneiformParserVisitor = &kfVisitor{}

func (k *kfVisitor) VisitProgram(ctx *gen.ProgramContext) interface {
} {
	schema := &types.Schema{
		Name:       ctx.Database_declaration().Accept(k).(string),
		Tables:     arr[*types.Table](len(ctx.AllTable_declaration())),
		Extensions: arr[*types.Extension](len(ctx.AllUse_declaration())),
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

	return schema
}

// START OF Table Declaration Parsing:

func (k *kfVisitor) VisitTable_declaration(ctx *gen.Table_declarationContext) interface {
} {
	tbl := &types.Table{
		Name:        ctx.IDENTIFIER().GetText(),
		Columns:     arr[*types.Column](len(ctx.AllColumn_def())),
		Indexes:     arr[*types.Index](len(ctx.AllIndex_def())),
		ForeignKeys: arr[*types.ForeignKey](len(ctx.AllForeign_key_def())),
	}

	k.checkUniqueName(tbl.Name)

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

func (k *kfVisitor) VisitIndex_def(ctx *gen.Index_defContext) interface {
} {
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

func (k *kfVisitor) VisitForeign_key_def(ctx *gen.Foreign_key_defContext) interface {
} {
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

func (k *kfVisitor) VisitForeign_key_action(ctx *gen.Foreign_key_actionContext) interface {
} {
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

func (k *kfVisitor) VisitColumn_def(ctx *gen.Column_defContext) interface {
} {
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

func (k *kfVisitor) VisitDEFAULT(ctx *gen.DEFAULTContext) interface {
} {
	val := ctx.Literal().Accept(k).(string)

	return &types.Attribute{
		Type:  types.DEFAULT,
		Value: val,
	}
}

func (k *kfVisitor) VisitMAX(ctx *gen.MAXContext) interface {
} {
	return &types.Attribute{
		Type:  types.MAX,
		Value: ctx.NUMERIC_LITERAL().GetText(),
	}
}

func (k *kfVisitor) VisitMAX_LEN(ctx *gen.MAX_LENContext) interface {
} {
	return &types.Attribute{
		Type:  types.MAX_LENGTH,
		Value: ctx.NUMERIC_LITERAL().GetText(),
	}
}

func (k *kfVisitor) VisitMIN(ctx *gen.MINContext) interface {
} {
	return &types.Attribute{
		Type:  types.MIN,
		Value: ctx.NUMERIC_LITERAL().GetText(),
	}
}

func (k *kfVisitor) VisitMIN_LEN(ctx *gen.MIN_LENContext) interface {
} {
	return &types.Attribute{
		Type:  types.MIN_LENGTH,
		Value: ctx.NUMERIC_LITERAL().GetText(),
	}
}

func (k *kfVisitor) VisitNOT_NULL(ctx *gen.NOT_NULLContext) interface {
} {
	return &types.Attribute{
		Type: types.NOT_NULL,
	}
}

func (k *kfVisitor) VisitPRIMARY_KEY(ctx *gen.PRIMARY_KEYContext) interface {
} {
	return &types.Attribute{
		Type: types.PRIMARY_KEY,
	}
}

func (k *kfVisitor) VisitUNIQUE(ctx *gen.UNIQUEContext) interface {
} {
	return &types.Attribute{
		Type: types.UNIQUE,
	}
}

// END OF Table Declaration Parsing

func (k *kfVisitor) VisitDatabase_declaration(ctx *gen.Database_declarationContext) interface {
} {
	return ctx.IDENTIFIER().GetText()
}

func (k *kfVisitor) VisitIdentifier_list(ctx *gen.Identifier_listContext) interface {
} {
	idents := arr[string](len(ctx.AllIDENTIFIER()))
	for i, ident := range ctx.AllIDENTIFIER() {
		idents[i] = ident.GetText()
	}

	return idents
}

func (k *kfVisitor) VisitLiteral(ctx *gen.LiteralContext) interface {
} {
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

func (k *kfVisitor) VisitType_selector(ctx *gen.Type_selectorContext) interface {
} {
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

func (k *kfVisitor) VisitUse_declaration(ctx *gen.Use_declarationContext) interface {
} {
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
			k.checkUniqueName(c.Alias)
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
func (k *kfVisitor) VisitStmt_mode(ctx *gen.Stmt_modeContext) interface {
} {
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
func (k *kfVisitor) VisitAction_declaration(ctx *gen.Action_declarationContext) interface {
} {
	name := ctx.STMT_IDENTIFIER().GetText()

	k.checkUniqueName(name)

	act := &types.Action{
		Name:        name,
		Annotations: arr[string](0),
		Parameters:  arr[string](len(ctx.AllSTMT_VAR())),
		Body:        parseBody(ctx.STMT_BODY().GetText()),
	}

	for i, v := range ctx.AllSTMT_VAR() {
		act.Parameters[i] = v.GetText()
	}

	var err error
	act.Modifiers, act.Public, err = getAccessModifiers(ctx.AllSTMT_ACCESS())
	if err != nil {
		panic(err)
	}

	return act
}

func (k *kfVisitor) VisitProcedure_declaration(ctx *gen.Procedure_declarationContext) interface {
} {
	name := ctx.GetProcedure_name().GetText()

	k.checkUniqueName(name)

	proc := &types.Procedure{
		Name:        name,
		Annotations: arr[string](0),
		Body:        parseBody(ctx.STMT_BODY().GetText()),
	}

	if ctx.Stmt_typed_param_list() != nil {
		proc.Parameters = ctx.Stmt_typed_param_list().Accept(k).([]*types.ProcedureParameter)
	}

	var err error
	proc.Modifiers, proc.Public, err = getAccessModifiers(ctx.AllSTMT_ACCESS())
	if err != nil {
		panic(err)
	}

	switch {
	case ctx.Table_return() != nil:
		proc.Returns = &types.ProcedureReturn{
			Table: ctx.Table_return().Accept(k).([]*types.NamedType),
		}
	case ctx.Stmt_type_list() != nil:
		proc.Returns = &types.ProcedureReturn{
			Types: ctx.Stmt_type_list().Accept(k).([]*types.DataType),
		}
	default:
		proc.Returns = nil
	}

	return proc
}

// parseBody will parse the body of a procedure or action, removing
// leading and traily braces and whitespace.
func parseBody(b string) string {
	b = strings.TrimPrefix(b, "{")
	b = strings.TrimSuffix(b, "}")
	b = strings.TrimSpace(b)

	return b
}

func (k *kfVisitor) VisitStmt_type_list(ctx *gen.Stmt_type_listContext) interface {
} {
	list := arr[*types.DataType](len(ctx.AllStmt_type_selector()))

	for i, t := range ctx.AllStmt_type_selector() {
		list[i] = t.Accept(k).(*types.DataType)
	}

	return list
}

func (k *kfVisitor) VisitStmt_typed_param_list(ctx *gen.Stmt_typed_param_listContext) interface {
} {
	params := arr[*types.ProcedureParameter](len(ctx.AllSTMT_VAR()))

	for i, v := range ctx.AllSTMT_VAR() {
		params[i] = &types.ProcedureParameter{
			Name: v.GetText(),
			Type: ctx.Stmt_type_selector(i).Accept(k).(*types.DataType),
		}
	}

	return params
}

func (k *kfVisitor) VisitStmt_type_selector(ctx *gen.Stmt_type_selectorContext) interface {
} {
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

func (k *kfVisitor) VisitTable_return(ctx *gen.Table_returnContext) interface {
} {
	cols := arr[*types.NamedType](len(ctx.AllSTMT_IDENTIFIER()))

	for i, c := range ctx.AllSTMT_IDENTIFIER() {
		cols[i] = &types.NamedType{
			Name: c.GetText(),
			Type: ctx.Stmt_type_selector(i).Accept(k).(*types.DataType),
		}
	}

	return cols
}

// getAccessModifiers returns the access modifiers for the given context.
// It also returns whether or not the action/procedure is public or private.
// If it can't find public or private, it will return an error.
func getAccessModifiers(mods []antlr.TerminalNode) (modifiers []types.Modifier, public bool, err error) {
	foundPublicOrPrivate := false
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
			err = fmt.Errorf("unexpected modifier: %s", m.GetText())
			return
		}
	}

	if !foundPublicOrPrivate {
		err = fmt.Errorf("missing public or private")
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
