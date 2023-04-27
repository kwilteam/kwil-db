package sql_parser

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"github.com/kwilteam/kwil-db/internal/pkg/sqlite"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/scanner"
	"github.com/kwilteam/kwil-db/pkg/kuneiform/token"
	"github.com/pkg/errors"

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

type ErrorHandler struct {
	CurLine int
	Errors  scanner.ErrorList
}

func NewErrorHandler(currentLine int) *ErrorHandler {
	return &ErrorHandler{
		CurLine: currentLine,
	}
}

func (eh *ErrorHandler) Add(column int, err error) {
	eh.Errors.Add(token.Position{
		Line:   token.Pos(eh.CurLine),
		Column: token.Pos(column),
	}, err.Error())
}

type sqliteErrorListener struct {
	*antlr.DefaultErrorListener
	*ErrorHandler

	symbol string
}

func newSqliteErrorListener(eh *ErrorHandler) *sqliteErrorListener {
	return &sqliteErrorListener{
		ErrorHandler: eh,
		symbol:       "",
	}
}

func (s *sqliteErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	symbol := offendingSymbol.(antlr.Token)
	if s.symbol == "" {
		s.symbol = symbol.GetText()
	}
	// calculate relative line number
	relativeLine := line - 1
	defer func() {
		s.ErrorHandler.CurLine -= relativeLine
	}()
	s.ErrorHandler.CurLine += relativeLine
	s.ErrorHandler.Add(column, errors.Wrap(ErrSyntax, msg))
}

func (s *sqliteErrorListener) ReportAmbiguity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, exact bool, ambigAlts *antlr.BitSet, configs antlr.ATNConfigSet) {
	//s.ErrorHandler.Add(startIndex, errors.Wrap(ErrAmbiguity, "ambiguity"))
}

func (s *sqliteErrorListener) ReportAttemptingFullContext(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, conflictingAlts *antlr.BitSet, configs antlr.ATNConfigSet) {
	//s.ErrorHandler.Add(startIndex, errors.Wrap(ErrAttemptingFullContext, "attempting full context"))
}

func (s *sqliteErrorListener) ReportContextSensitivity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex, prediction int, configs antlr.ATNConfigSet) {
	//s.ErrorHandler.Add(startIndex, errors.Wrap(ErrContextSensitivity, "context sensitivity"))
}

type KlSqliteListener struct {
	*sqlite.BaseSQLiteParserListener
	*ErrorHandler

	actionCtx ActionContext
	dbCtx     DatabaseContext

	trace bool

	iuStarted bool //insert, update

	joinStarted bool
	joinConsCnt int
	joinOpCnt   int

	joinConsStarted bool

	exprLevel int
	exprStack []string

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

func NewKlSqliteListener(eh *ErrorHandler, actionName string, ctx DatabaseContext, opts ...KlSqliteListenerOption) *KlSqliteListener {

	k := &KlSqliteListener{ErrorHandler: eh, actionCtx: ctx.Actions[actionName], dbCtx: ctx}
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

func (tl *KlSqliteListener) startJoinCons() {
	tl.joinConsStarted = true
}

func (tl *KlSqliteListener) endJoinCons() {
	tl.joinConsStarted = false
}

func (tl *KlSqliteListener) inJoinCons() bool {
	return tl.joinConsStarted
}

func (tl *KlSqliteListener) enterExpr() {
	tl.exprLevel++
}

func (tl *KlSqliteListener) exitExpr() {
	tl.exprLevel--
}

func (tl *KlSqliteListener) joinOnExprTooDeep() bool {
	return tl.exprLevel > 2
}

func (tl *KlSqliteListener) inExpr() bool {
	return tl.exprLevel > 0
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
			tl.ErrorHandler.Add(tok.GetColumn(), errors.Wrap(ErrFunctionNotSupported, fnName))
		}
	}
}

func (tl *KlSqliteListener) VisitTerminal(node antlr.TerminalNode) {
	if tl.trace {
		fmt.Println("visit terminal ", node.GetText())
	}
}

func (tl *KlSqliteListener) EnterColumn_name(ctx *sqlite.Column_nameContext) {
	if tl.trace {
		fmt.Println("enter COLUMN ", ctx.GetText())
	}

	// TODO: validate column name exist
}

func (tl *KlSqliteListener) EnterTable_name(ctx *sqlite.Table_nameContext) {
	if tl.trace {
		fmt.Println("enter TABLE ", ctx.GetText())
	}

	name := ctx.GetText()
	if _, ok := tl.dbCtx.Tables[name]; !ok {
		tok := ctx.GetStart()
		tl.ErrorHandler.Add(tok.GetColumn(), errors.Wrap(ErrTableNotFound, name))
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

func (tl *KlSqliteListener) EnterCommon_table_expression(ctx *sqlite.Common_table_expressionContext) {
	if tl.trace {
		fmt.Println("enter CTE ", ctx.GetText())
	}

	table := ctx.Table_name().GetText()

	// support table alias
	tl.dbCtx.Tables[table] = TableContext{}
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
		tl.ErrorHandler.Add(0, ErrSelectFromMultipleTables)
	}
}

func (tl *KlSqliteListener) EnterCreate_table_stmt(ctx *sqlite.Create_table_stmtContext) {
	if tl.trace {
		fmt.Println("enter CREATE TABLE ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.ErrorHandler.Add(tok.GetColumn(), ErrCreateTableNotSupported)
}

func (tl *KlSqliteListener) EnterCreate_index_stmt(ctx *sqlite.Create_index_stmtContext) {
	if tl.trace {
		fmt.Println("enter CREATE INDEX ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.ErrorHandler.Add(tok.GetColumn(), ErrCreateIndexNotSupported)
}

func (tl *KlSqliteListener) EnterCreate_trigger_stmt(ctx *sqlite.Create_trigger_stmtContext) {
	if tl.trace {
		fmt.Println("enter CREATE TRIGGER ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.ErrorHandler.Add(tok.GetColumn(), ErrCreateTriggerNotSupported)
}

func (tl *KlSqliteListener) EnterCreate_view_stmt(ctx *sqlite.Create_view_stmtContext) {
	if tl.trace {
		fmt.Println("enter CREATE VIEW ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.ErrorHandler.Add(tok.GetColumn(), ErrCreateViewNotSupported)
}

func (tl *KlSqliteListener) EnterCreate_virtual_table_stmt(ctx *sqlite.Create_virtual_table_stmtContext) {
	if tl.trace {
		fmt.Println("enter CREATE VIRTUAL TABLE ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.ErrorHandler.Add(tok.GetColumn(), ErrCreateVirtualTableNotSupported)
}

func (tl *KlSqliteListener) EnterDrop_stmt(ctx *sqlite.Drop_stmtContext) {
	if tl.trace {
		fmt.Println("enter DROP TABLE ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.ErrorHandler.Add(tok.GetColumn(), ErrDropTableNotSupported)
}

func (tl *KlSqliteListener) EnterAlter_table_stmt(ctx *sqlite.Alter_table_stmtContext) {
	if tl.trace {
		fmt.Println("enter ALTER TABLE ", ctx.GetText())
	}

	tok := ctx.GetStart()
	tl.ErrorHandler.Add(tok.GetColumn(), ErrAlterTableNotSupported)
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
	defer func() {
		tl.endJoin()
		if tl.trace {
			fmt.Println("exit Join clause ", ctx.GetText())
		}
	}()

	if tl.joinOpCnt > JoinCountAllowed {
		tok := ctx.GetStart()
		tl.ErrorHandler.Add(tok.GetColumn(), ErrMultiJoinNotSupported)
	}

	if tl.joinConsCnt == 0 {
		tok := ctx.GetStart()
		tl.ErrorHandler.Add(tok.GetColumn(), ErrJoinWithoutCondition)
	}
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
		tl.ErrorHandler.Add(op.GetColumn(), ErrJoinUsingNotSupported)
		return
	}

	tl.startJoinCons()
	tl.joinConsCnt++
}

func (tl *KlSqliteListener) ExitJoin_constraint(ctx *sqlite.Join_constraintContext) {
	defer func() {
		tl.endJoinCons()
		if tl.trace {
			fmt.Println("exit Join constraint ", ctx.GetText())
		}
	}()
}

func (tl *KlSqliteListener) EnterExpr(ctx *sqlite.ExprContext) {
	if tl.trace {
		fmt.Println("EnterExpr", ctx.GetText())
	}

	tl.enterExpr()

	cnt := ctx.GetChildCount()
	switch cnt {
	case 1:
		// literal or name
	case 2:
		// unary op
	default:
		// if first child is function, it is function
		first := ctx.GetChild(0)
		if _, ok := first.(*sqlite.Function_nameContext); ok {
			tl.startFn()
		} else {
			// infix op
		}
	}
}

func (tl *KlSqliteListener) ExitExpr(ctx *sqlite.ExprContext) {
	defer func() {
		tl.exitExpr()
		if tl.trace {
			fmt.Println("ExitExpr", ctx.GetText())
		}
	}()

	cnt := ctx.GetChildCount()
	switch cnt {
	case 1:
		// literal or name

		// TODO: validate this after the whole expression is parsed?
		if v := ctx.BIND_PARAMETER(); v != nil {
			tok := ctx.GetStart()
			param := tok.GetText()
			paramPrefix := param[:1]
			paramName := param[1:]
			switch paramPrefix {
			case BindParameterPrefix:
				// refer to action parameter
				p := param
				if _, ok := tl.actionCtx[p]; !ok {
					tl.ErrorHandler.Add(tok.GetColumn(), errors.Wrap(ErrBindParameterNotFound, p))
				}
			case ModifierPrefix:
				// refer to modifier
				p := paramName
				if _, ok := Modifiers[Modifier(p)]; !ok {
					tl.ErrorHandler.Add(tok.GetColumn(), errors.Wrap(ErrModifierNotSupported, p))
				}
			default:
				tl.ErrorHandler.Add(tok.GetColumn(), errors.Wrap(ErrBindParameterPrefixNotSupported, paramPrefix))
			}
		}
	case 2:
		// unary op
		op := ctx.GetStart()
		if tl.inJoinCons() {
			tl.ErrorHandler.Add(op.GetColumn(), errors.Wrap(ErrJoinConditionOpNotSupported, "unary"))
		}
	default:
		first := ctx.GetChild(0)
		// if first child is function, it is function expr
		fnExpr, isFn := first.(*sqlite.Function_nameContext)

		if !isFn {
			dot := ctx.DOT(0)
			if dot != nil { // always check table.column existence
				cds := ctx.GetChildren()
				tableName := cds[0].(*sqlite.Table_nameContext).GetText()
				operator := cds[1].(*antlr.TerminalNodeImpl).GetSymbol()
				columnName := cds[2].(*sqlite.Column_nameContext).GetText()

				if _, ok := tl.dbCtx.Tables[tableName]; !ok {
					tl.ErrorHandler.Add(operator.GetColumn(), errors.Wrap(ErrTableNotFound, tableName))
					return
				}

				found := false
				for _, c := range tl.dbCtx.Tables[tableName].Columns {
					if c == columnName {
						found = true
						break
					}
				}
				if !found {
					tl.ErrorHandler.Add(operator.GetColumn(),
						errors.Wrap(ErrColumnNotFound, fmt.Sprintf("%s.%s", tableName, columnName)))
					return
				}
			}
		}

		if tl.inJoinCons() {
			if isFn {
				tl.ErrorHandler.Add(ctx.GetStart().GetColumn(), errors.Wrap(ErrJoinConditionFuncNotSupported, fnExpr.GetText()))
				return
			}

			// dot have been checked above
			dot := ctx.DOT(0)
			if dot != nil {
				return
			}

			// normal infix expr
			cds := ctx.GetChildren()
			leftExpr, leftIsExpr := cds[0].(*sqlite.ExprContext)
			rightExpr, rightIsExpr := cds[2].(*sqlite.ExprContext)
			if tl.joinOnExprTooDeep() && (leftIsExpr || rightIsExpr) {
				tl.ErrorHandler.Add(ctx.GetStart().GetColumn(), errors.Wrap(ErrJoinConditionTooDeep, ctx.GetText()))
				return
			}

			operator := cds[1].(*antlr.TerminalNodeImpl).GetSymbol()
			opName := strings.ToLower(operator.GetText())

			if _, ok := klStaticData.allowedJoinOps[opName]; !ok {
				tl.ErrorHandler.Add(operator.GetColumn(), errors.Wrap(ErrJoinConditionOpNotSupported, opName))
				return
			}

			if leftIsExpr && rightIsExpr {
				if localEval(opName, leftExpr.GetText(), rightExpr.GetText()) {
					tl.ErrorHandler.Add(operator.GetColumn(), errors.Wrap(ErrJoinWithTrueCondition, opName))
					return
				}
			}
		} else if isFn {
			tl.banFunction(ctx)
		}
	}
}

func (tl *KlSqliteListener) EnterLiteral_value(ctx *sqlite.Literal_valueContext) {
	if tl.trace {
		fmt.Println("EnterLiteral_value", ctx.GetText())
	}

	ctx.TRUE_()

	name := ctx.GetText()
	name = strings.ToLower(name)
	if _, ok := klStaticData.banKeywords[name]; ok {
		tok := ctx.GetStart()
		tl.ErrorHandler.Add(tok.GetColumn(), errors.Wrap(ErrKeywordNotSupported, name))
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
}

func (tl *KlSqliteListener) EnterJoin_operator(ctx *sqlite.Join_operatorContext) {
	if tl.trace {
		fmt.Println("enter JoinOper ", ctx.GetText())
	}

	tl.joinOpCnt++
}

func (tl *KlSqliteListener) ExitJoin_operator(ctx *sqlite.Join_operatorContext) {
	defer func() {
		if tl.trace {
			fmt.Println("exit JoinOper ", ctx.GetText())
		}
	}()

	tok := ctx.GetStart()
	name := strings.ToLower(tok.GetText())
	if _, ok := klStaticData.banJoins[name]; ok {
		tl.ErrorHandler.Add(tok.GetColumn(), errors.Wrap(ErrJoinNotSupported, name))
	}
}

func (tl *KlSqliteListener) EnterAny_name(ctx *sqlite.Any_nameContext) {
	if tl.trace {
		fmt.Println("enter AnyName ", ctx.GetText())
	}
}
