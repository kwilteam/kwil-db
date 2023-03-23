package sql

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/pkg/errors"
	"kwil/internal/pkg/sqlite"
	"kwil/pkg/kl/ast"
	"kwil/pkg/kl/scanner"
	"kwil/pkg/kl/token"
	"strings"
	"sync"
)

var (
	argNow   = []string{"'now'"}
	argNow2  = []string{"*", "'now'"}
	argEmpty = []string{}
)

var klStaticData struct {
	once sync.Once

	banKeywords    map[string]bool
	banFunctions   map[string][]string
	banJoins       map[string]bool
	allowedJoinOps map[string]bool
}

func buildMap(data []string) map[string]bool {
	m := make(map[string]bool)
	for _, v := range data {
		m[v] = true
	}
	return m
}

func buildMapFn(fn []string, argsMap map[string][]string) map[string][]string {
	m := make(map[string][]string)
	for _, v := range fn {
		args, ok := argsMap[v]
		if !ok {
			args = argEmpty
		}
		m[v] = args
	}
	return m
}

func banDataInit() {
	banFunctions := []string{
		// time
		"date", "time", "datetime", "julianday", "unixepoch", "strftime",
		// random
		"random", "randomblob",
		//changes
		"changes", "last_insert_rowid", "total_changes",
		// math
		"acos", "acosh", "asin", "asinh", "atan", "atan2", "atanh", "ceil", "ceiling", "cos", "cosh", "degrees",
		"exp", "floor", "ln", "log", "log10", "log2", "mod", "pi", "pow", "power", "radians", "sin", "sinh",
		"sqrt", "tan", "tanh", "trunc",
	}

	banFunctionArgs := map[string][]string{
		"date":      argNow,
		"time":      argNow,
		"datetime":  argNow,
		"julianday": argNow,
		"unixepoch": argNow,
		"strftime":  argNow2,
	}

	banKeywords := []string{
		"current_time", "current_date", "current_timestamp", //time
	}
	banJoins := []string{"cross", "natural"} // explicit cross join
	allowedJoinOps := []string{"=", "!="}

	klStaticData.banFunctions = buildMapFn(banFunctions, banFunctionArgs)
	klStaticData.banKeywords = buildMap(banKeywords)
	klStaticData.banJoins = buildMap(banJoins)
	klStaticData.allowedJoinOps = buildMap(allowedJoinOps)
}

func KlSQLInit() {
	banData := &klStaticData
	banData.once.Do(banDataInit)
}

type errorHandler struct {
	CurLine int
	Errors  scanner.ErrorList
}

func (eh *errorHandler) Add(column int, err error) {
	eh.Errors.Add(token.Position{
		Line:   token.Pos(eh.CurLine),
		Column: token.Pos(column),
	}, err.Error())
}

type sqliteErrorListener struct {
	*antlr.DefaultErrorListener
	*errorHandler
}

func (s *sqliteErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	s.errorHandler.Add(column, errors.Wrap(ErrSyntax, msg))
}

type KlSqliteListener struct {
	*sqlite.BaseSQLiteParserListener
	*errorHandler

	ctx ast.ActionContext

	trace bool

	iuStarted bool //insert, update

	joinCond      map[string]bool
	joinCondStack []string
	joinStarted   bool
	joinConsCnt   int
	joinOpCnt     int

	exprStated bool
	exprStack  []string

	// fn name, fn params
	fnParams  []string
	fnStarted bool
}

var _ sqlite.SQLiteParserListener = &KlSqliteListener{}

type KlSqliteListenerOption func(*KlSqliteListener)

func WithTrace() KlSqliteListenerOption {
	return func(l *KlSqliteListener) {
		l.trace = true
	}
}

func NewKlSqliteListener(eh *errorHandler, ctx ast.ActionContext, opts ...KlSqliteListenerOption) *KlSqliteListener {
	k := &KlSqliteListener{errorHandler: eh, ctx: ctx}
	for _, opt := range opts {
		opt(k)
	}
	return k
}

func (tl *KlSqliteListener) startIU() {
	tl.iuStarted = true
}

func (tl *KlSqliteListener) endIU() {
	tl.iuStarted = false
}

func (tl *KlSqliteListener) inIU() bool {
	return tl.iuStarted
}

func (tl *KlSqliteListener) startJoin() {
	tl.joinOpCnt = 0
	tl.joinConsCnt = 0
	tl.joinStarted = true
}

func (tl *KlSqliteListener) endJoin() {
	tl.joinOpCnt = -1
	tl.joinConsCnt = -1
	tl.joinStarted = false
}

func (tl *KlSqliteListener) inJoin() bool {
	return tl.joinStarted
}

func (tl *KlSqliteListener) startExpr() {
	tl.exprStated = true
}

func (tl *KlSqliteListener) endExpr() {
	tl.exprStated = false
}

func (tl *KlSqliteListener) inExpr() bool {
	return tl.exprStated
}

func (tl *KlSqliteListener) startFn() {
	tl.fnStarted = true
}

func (tl *KlSqliteListener) endFn() {
	tl.fnStarted = false
}

func (tl *KlSqliteListener) inFn() bool {
	return tl.fnStarted
}

// banFunction bans functions based on the function name and arguments
func (tl *KlSqliteListener) banFunction(ctx *sqlite.ExprContext) {
	defer func() {
		tl.endFn()
	}()

	//if !tl.inIU() { // short circuit on only select/delete
	//	return
	//}

	fnName, fnArgs := tl.fnParams[0], tl.fnParams[1:]
	if args, ok := klStaticData.banFunctions[fnName]; ok {
		if len(fnArgs) < len(args) {
			args = args[:len(fnArgs)]
		}

		// fn match
		argMatch := true
		for i, arg := range args {
			if args[i] == "*" {
				continue
			}

			if arg != fnArgs[i] {
				argMatch = false
				break
			}
		}
		if argMatch {
			tok := ctx.GetStart()
			tl.errorHandler.Add(tok.GetColumn(), errors.Wrap(ErrFunctionNotSupported, fnName))
		}
	}
}

func (tl *KlSqliteListener) joinPush(cond string) {
	tl.joinCondStack = append(tl.joinCondStack, cond)
}

func (tl *KlSqliteListener) joinPop() string {
	if len(tl.joinCondStack) == 0 {
		panic("joinPop: empty stack")
	}

	cond := tl.joinCondStack[len(tl.joinCondStack)-1]
	tl.joinCondStack = tl.joinCondStack[:len(tl.joinCondStack)-1]
	return cond
}

func (tl *KlSqliteListener) VisitTerminal(node antlr.TerminalNode) {
	if tl.trace {
		fmt.Println("visit terminal ", node.GetText())
	}
}

func (tl *KlSqliteListener) EnterFunction_name(ctx *sqlite.Function_nameContext) {
	if tl.trace {
		fmt.Println("enter FUNCTION ", ctx.GetText())
	}

	name := ctx.GetText()
	name = strings.ToLower(name)

	tl.fnParams = append(tl.fnParams, name)
}

func (tl *KlSqliteListener) EnterSelect_stmt(ctx *sqlite.Select_stmtContext) {
	if tl.trace {
		fmt.Println("enter SELECT ", ctx.GetText())
	}
}

func (tl *KlSqliteListener) EnterSelect_core(ctx *sqlite.Select_coreContext) {
	if tl.trace {
		fmt.Println("enter SELECT CORE ", ctx.GetText())
	}

	cds := ctx.GetChildren()
	tbs := 0
	for _, cd := range cds {
		switch cd.(type) {
		case *sqlite.Table_or_subqueryContext:
			tbs++
		}
	}

	if tbs > 1 {
		tl.errorHandler.Add(0, ErrSelectFromMultipleTables)
	}
}

func (tl *KlSqliteListener) EnterCreate_table_stmt(ctx *sqlite.Create_table_stmtContext) {
	if tl.trace {
		fmt.Println("enter CREATE TABLE ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.errorHandler.Add(tok.GetColumn(), ErrCreateTableNotSupported)
}

func (tl *KlSqliteListener) EnterCreate_index_stmt(ctx *sqlite.Create_index_stmtContext) {
	if tl.trace {
		fmt.Println("enter CREATE INDEX ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.errorHandler.Add(tok.GetColumn(), ErrCreateIndexNotSupported)
}

func (tl *KlSqliteListener) EnterCreate_trigger_stmt(ctx *sqlite.Create_trigger_stmtContext) {
	if tl.trace {
		fmt.Println("enter CREATE TRIGGER ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.errorHandler.Add(tok.GetColumn(), ErrCreateTriggerNotSupported)
}

func (tl *KlSqliteListener) EnterCreate_view_stmt(ctx *sqlite.Create_view_stmtContext) {
	if tl.trace {
		fmt.Println("enter CREATE VIEW ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.errorHandler.Add(tok.GetColumn(), ErrCreateViewNotSupported)
}

func (tl *KlSqliteListener) EnterCreate_virtual_table_stmt(ctx *sqlite.Create_virtual_table_stmtContext) {
	if tl.trace {
		fmt.Println("enter CREATE VIRTUAL TABLE ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.errorHandler.Add(tok.GetColumn(), ErrCreateVirtualTableNotSupported)
}

func (tl *KlSqliteListener) EnterDrop_stmt(ctx *sqlite.Drop_stmtContext) {
	if tl.trace {
		fmt.Println("enter DROP TABLE ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.errorHandler.Add(tok.GetColumn(), ErrDropTableNotSupported)
}

func (tl *KlSqliteListener) EnterAlter_table_stmt(ctx *sqlite.Alter_table_stmtContext) {
	if tl.trace {
		fmt.Println("enter ALTER TABLE ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.errorHandler.Add(tok.GetColumn(), ErrAlterTableNotSupported)
}

func (tl *KlSqliteListener) EnterInsert_stmt(ctx *sqlite.Insert_stmtContext) {
	if tl.trace {
		fmt.Println("enter INSERT ", ctx.GetText())
	}

	tl.startIU()
}

func (tl *KlSqliteListener) ExitInsert_stmt(ctx *sqlite.Insert_stmtContext) {
	if tl.trace {
		fmt.Println("exit INSERT ", ctx.GetText())
	}

	tl.endIU()
}

func (tl *KlSqliteListener) EnterUpdate_stmt(ctx *sqlite.Update_stmtContext) {
	if tl.trace {
		fmt.Println("enter UPDATE ", ctx.GetText())
	}

	tl.startIU()
}

func (tl *KlSqliteListener) ExitUpdate_stmt(ctx *sqlite.Update_stmtContext) {
	if tl.trace {
		fmt.Println("exit UPDATE ", ctx.GetText())
	}

	tl.endIU()
}

func (tl *KlSqliteListener) EnterUpsert_clause(ctx *sqlite.Upsert_clauseContext) {
	if tl.trace {
		fmt.Println("enter UPSERT ", ctx.GetText())
	}

	tl.startIU()
}

func (tl *KlSqliteListener) ExitUpsert_clause(ctx *sqlite.Upsert_clauseContext) {
	if tl.trace {
		fmt.Println("exit UPSERT ", ctx.GetText())
	}

	tl.endIU()
}

func (tl *KlSqliteListener) EnterSimple_select_stmt(ctx *sqlite.Simple_select_stmtContext) {
	if tl.trace {
		fmt.Println("enter SIMPLE SELECT ", ctx.GetText())
	}
}

func (tl *KlSqliteListener) EnterJoin_clause(ctx *sqlite.Join_clauseContext) {
	if tl.trace {
		fmt.Println("enter Join clause ", ctx.GetText())
	}

	tl.startJoin()
}

func (tl *KlSqliteListener) ExitJoin_clause(ctx *sqlite.Join_clauseContext) {
	if tl.trace {
		fmt.Println("exit Join clause ", ctx.GetText())
	}

	if tl.joinOpCnt > JoinCountAllowed {
		tok := ctx.GetStart()
		tl.errorHandler.Add(tok.GetColumn(), ErrMultiJoinNotSupported)
	}

	if tl.joinConsCnt == 0 {
		tok := ctx.GetStart()
		tl.errorHandler.Add(tok.GetColumn(), ErrJoinWithoutCondition)
	}

	tl.endJoin()
}

func localEval(op, left, right string) bool {
	switch op {
	case "=":
		return left == right
	case "!=":
		return left != right
	}
	return true
}

func (tl *KlSqliteListener) EnterJoin_constraint(ctx *sqlite.Join_constraintContext) {
	if tl.trace {
		fmt.Println("enter Join constraint ", ctx.GetText())
	}

	op := ctx.GetStart()
	tt := strings.ToLower(op.GetText())
	// TODO: support "using"
	if tt == "using" {
		tl.errorHandler.Add(op.GetColumn(), ErrJoinUsingNotSupported)
		return
	}

	tl.joinConsCnt++
}

func (tl *KlSqliteListener) EnterExpr(ctx *sqlite.ExprContext) {
	if tl.trace {
		fmt.Println("EnterExpr", ctx.GetText())
	}

	cnt := ctx.GetChildCount()
	switch cnt {
	case 1:
		// literal or name
		// TODO: validate this after the whole expression is parsed?
		if v := ctx.BIND_PARAMETER(); v != nil {
			tok := ctx.GetStart()
			param := tok.GetText()
			paramPrefix := param[:1]
			switch paramPrefix {
			case BindParameterPrefix:
				// refer to action parameter
				p := strings.ReplaceAll(v.GetText(), "$", "")
				if _, ok := tl.ctx[p]; !ok {
					tl.errorHandler.Add(tok.GetColumn(), ErrBindParameterNotFound)
				}
			case ModifierPrefix:
				// refer to modifier
				p := strings.ReplaceAll(v.GetText(), "@", "")
				if _, ok := Modifiers[Modifier(p)]; !ok {
					tl.errorHandler.Add(tok.GetColumn(), ErrModifierNotSupported)
				}
			default:
				tl.errorHandler.Add(tok.GetColumn(), errors.Wrap(ErrBindParameterPrefixNotSupported, paramPrefix))
			}
		}
	case 2:
		// unary op
	default:
		// if first child is function, it is function
		first := ctx.GetChild(0)
		if _, ok := first.(*sqlite.Function_nameContext); ok {
			tl.startFn()
		}
	}
}

func (tl *KlSqliteListener) ExitExpr(ctx *sqlite.ExprContext) {
	if tl.trace {
		fmt.Println("ExitExpr", ctx.GetText())
	}

	cnt := ctx.GetChildCount()
	switch cnt {
	case 1:
		// literal or name
	case 2:
		// unary op
		op := ctx.GetStart()
		if tl.inJoin() {
			tl.errorHandler.Add(op.GetColumn(), errors.Wrap(ErrJoinConditionOpNotSupported, "unary"))
		}
	default:
		// if first child is function, it is function
		first := ctx.GetChild(0)
		if _, ok := first.(*sqlite.Function_nameContext); ok {
			tl.banFunction(ctx)
		} else if tl.inJoin() {
			// infix expr
			cds := ctx.GetChildren()
			operator := cds[1].(*antlr.TerminalNodeImpl).GetSymbol()
			tokName := strings.ToLower(operator.GetText())
			if _, ok := klStaticData.allowedJoinOps[tokName]; !ok {
				tl.errorHandler.Add(operator.GetColumn(), errors.Wrap(ErrJoinConditionOpNotSupported, tokName))
			}
		}
	}
}

func (tl *KlSqliteListener) EnterLiteral_value(ctx *sqlite.Literal_valueContext) {
	if tl.trace {
		fmt.Println("EnterLiteral_value", ctx.GetText())
	}

	name := ctx.GetText()
	name = strings.ToLower(name)
	if _, ok := klStaticData.banKeywords[name]; ok {
		tok := ctx.GetStart()
		tl.errorHandler.Add(tok.GetColumn(), errors.Wrap(ErrKeywordNotSupported, name))
		return
	}

	if tl.inFn() {
		tl.fnParams = append(tl.fnParams, name)
	}
}

func (tl *KlSqliteListener) ExitLiteral_value(ctx *sqlite.Literal_valueContext) {
	if tl.trace {
		fmt.Println("ExitLiteral_value", ctx.GetText())
	}

	tl.joinPush(ctx.GetText())
}

func (tl *KlSqliteListener) ExitJoin_constraint(ctx *sqlite.Join_constraintContext) {
	if tl.trace {
		fmt.Println("exit Join constraint ", ctx.GetText())
	}

	//tok := ctx.GetStop()
	// condition always return true
	//if tl.joinCond {
	//	tl.errorHandler.Add(tok.GetColumn(), "implicit cartesian join(3) is not supported")
	//}

	//fmt.Println("joinCondStack...... ", tl.joinCondStack)
}

func (tl *KlSqliteListener) EnterJoin_operator(ctx *sqlite.Join_operatorContext) {
	if tl.trace {
		fmt.Println("enter JoinOper ", ctx.GetText())
	}

	tl.joinOpCnt++
}

func (tl *KlSqliteListener) ExitJoin_operator(ctx *sqlite.Join_operatorContext) {
	if tl.trace {
		fmt.Println("exit JoinOper ", ctx.GetText())
	}

	tok := ctx.GetStart()
	name := strings.ToLower(tok.GetText())
	if _, ok := klStaticData.banJoins[name]; ok {
		tl.errorHandler.Add(tok.GetColumn(), errors.Wrap(ErrJoinNotSupported, name))
	}
}
