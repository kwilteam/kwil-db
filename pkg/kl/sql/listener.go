package sql

import (
	"fmt"
	"github.com/antlr/antlr4/runtime/Go/antlr/v4"
	"kwil/internal/pkg/sqlite"
	"kwil/pkg/kl/scanner"
	"kwil/pkg/kl/token"
	"strings"
	"sync"
)

var klStaticData struct {
	once          sync.Once
	banKeywords   []string
	banKeywordMap map[string]bool

	banFunctions    []string
	banFunctionsMap map[string]bool

	banJoins   []string
	banJoinMap map[string]bool

	allowedJoinOps   []string
	allowedJoinOpMap map[string]bool
}

func buildMap(data []string) map[string]bool {
	m := make(map[string]bool)
	for _, v := range data {
		m[v] = true
	}
	return m
}

func banDataInit() {
	klStaticData.banFunctions = []string{
		"date", "time", "datetime", "julianday", "unixepoch", //time
		//"strftime", //time with special parameter
		"random", "randomblob", //random
		"changes", "last_insert_rowid", "total_changes", //changes
	}

	klStaticData.banKeywords = []string{
		"current_time", "current_date", "current_timestamp", //time
	}

	klStaticData.banJoins = []string{"cross", "natural"} // explicit cross join

	klStaticData.allowedJoinOps = []string{"=", "!="}

	klStaticData.banFunctionsMap = buildMap(klStaticData.banFunctions)
	klStaticData.banJoinMap = buildMap(klStaticData.banJoins)
	klStaticData.allowedJoinOpMap = buildMap(klStaticData.allowedJoinOps)
}

func KlSQLInit() {
	banData := &klStaticData
	banData.once.Do(banDataInit)
}

type errorHandler struct {
	CurLine int
	Errors  scanner.ErrorList
}

func (eh *errorHandler) Add(column int, msg string) {
	eh.Errors.Add(token.Position{
		Line:   token.Pos(eh.CurLine),
		Column: token.Pos(column),
	}, msg)
}

type sqliteErrorListener struct {
	*antlr.DefaultErrorListener
	*errorHandler
}

func (s *sqliteErrorListener) SyntaxError(recognizer antlr.Recognizer, offendingSymbol interface{}, line, column int, msg string, e antlr.RecognitionException) {
	s.errorHandler.Add(column, msg)
}

type KlSqliteListener struct {
	*sqlite.BaseSQLiteParserListener
	*errorHandler

	trace bool

	joinCond      map[string]bool
	joinCondStack []string
	inJoin        bool
	joinCondCnt   int
}

var _ sqlite.SQLiteParserListener = &KlSqliteListener{}

func NewSqliteListener(eh *errorHandler) *KlSqliteListener {
	return &KlSqliteListener{errorHandler: eh}
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

func (tl *KlSqliteListener) EnterFunction_name(ctx *sqlite.Function_nameContext) {
	if tl.trace {
		fmt.Println("enter FUNCTION ", ctx.GetText())
	}

	tok := ctx.GetStart()
	name := ctx.GetText()
	name = strings.ToLower(name)
	if _, ok := klStaticData.banFunctionsMap[name]; ok {
		tl.errorHandler.Add(tok.GetColumn(), fmt.Sprintf("function %s is not supported", name))
	}
}

func (tl *KlSqliteListener) EnterSelect_stmt(ctx *sqlite.Select_stmtContext) {
	if tl.trace {
		fmt.Println("enter SELECT ", ctx.GetText())
	}

	//cds := ctx.GetChildren()
	//for _, cd := range cds {
	//	switch c := cd.(type) {
	//	case *sqlite.Common_table_stmtContext:
	//	case *sqlite.Select_coreContext:
	//	case *sqlite.Order_by_stmtContext:
	//	case *sqlite.Limit_stmtContext:
	//	case *sqlite.Compound_operatorContext:
	//	}
	//}
}

func (tl *KlSqliteListener) EnterSelect_core(ctx *sqlite.Select_coreContext) {
	if tl.trace {
		fmt.Println("enter SELECT CORE ", ctx.GetText())
	}

	cds := ctx.GetChildren()
	tbs := 0
	for _, cd := range cds {
		switch c := cd.(type) {
		case *sqlite.Table_or_subqueryContext:
			fmt.Println("==table or subquery ", c.GetText())
			tbs++
			//case *sqlite.Result_columnContext:
			//case *sqlite.Join_clauseContext:
			//case *sqlite.ExprContext:
			//case *sqlite.Window_nameContext:
			//case *sqlite.Window_functionContext:
			//case *sqlite.Values_clauseContext:
		}
	}

	if tbs > 1 {
		tl.errorHandler.Add(0, "implicit cartesian join(1) is not supported")
	}
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

	cds := ctx.GetChildren()

	tl.joinCondCnt = 0
	tl.inJoin = true

	tl.joinCond = make(map[string]bool, 0)

	joinLevel := 0
	//joinCondCnt := 0

	for _, cd := range cds {
		switch cd.(type) {
		//case *sqlite.Table_or_subqueryContext:
		case *sqlite.Join_operatorContext:
			joinLevel++
			//case *sqlite.Join_constraintContext:
		}
	}

	if joinLevel > 1 {
		tok := ctx.GetStart()
		tl.errorHandler.Add(tok.GetColumn(), "multi-level join is not supported")
	}
}

func (tl *KlSqliteListener) ExitJoin_clause(ctx *sqlite.Join_clauseContext) {
	if tl.trace {
		fmt.Println("exit Join clause ", ctx.GetText())
	}

	tl.inJoin = false

	if tl.joinCondCnt == 0 {
		tok := ctx.GetStart()
		tl.errorHandler.Add(tok.GetColumn(), "implicit cartesian join(2) is not supported")
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

	cond := ctx.GetText()
	tl.joinCond[cond] = false

	op := ctx.GetStart()
	tt := strings.ToLower(op.GetText())
	if tt == "using" {
		tl.errorHandler.Add(op.GetColumn(), "join using is not supported")
		return
	}

	tl.joinCondCnt++

	ccds := ctx.GetChildren()
	// ccds[0] is "ON"

	// TODO: use separate Enter hooks
	for _, ccd := range ccds {
		switch ccc := ccd.(type) {
		case *sqlite.ExprContext:
			fmt.Println("----expr ", ccc.GetText())
			fmt.Println("----expr child count ", ccc.GetChildCount())
			exprs := ccc.GetChildren()
			// infix expr
			if len(exprs) == 3 {
				for _, expr := range exprs {
					switch e := expr.(type) {
					case *sqlite.ExprContext:
					case *antlr.TerminalNodeImpl:
						tok := e.GetSymbol()
						tokName := strings.ToLower(tok.GetText())
						if _, ok := klStaticData.allowedJoinOpMap[tokName]; !ok {
							tl.errorHandler.Add(tok.GetColumn(), fmt.Sprintf("join op(%s) is not supported", tokName))
						}
					}
				}
			} else {
				tl.errorHandler.Add(op.GetColumn(), "join condition op(unary) is not supported")
			}
			//case *antlr.TerminalNodeImpl:
			//	// `on` or `using`
			//	fmt.Println("+-terminal node ", ccc.GetText())
		}
	}
}

func (tl *KlSqliteListener) EnterExpr(ctx *sqlite.ExprContext) {
	if tl.trace {
		fmt.Println("EnterExpr", ctx.GetText())
	}
}

func (tl *KlSqliteListener) EnterLiteral_value(ctx *sqlite.Literal_valueContext) {
	if tl.trace {
		fmt.Println("EnterLiteral_value", ctx.GetText())
	}
}

func (tl *KlSqliteListener) ExitLiteral_value(ctx *sqlite.Literal_valueContext) {
	if tl.trace {
		fmt.Println("ExitLiteral_value", ctx.GetText())
	}

	//tl.joinPush(ctx.GetText())
}

func (tl *KlSqliteListener) ExitExpr(ctx *sqlite.ExprContext) {
	if tl.trace {
		fmt.Println("ExitExpr", ctx.GetText())
	}
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
}

func (tl *KlSqliteListener) ExitJoin_operator(ctx *sqlite.Join_operatorContext) {
	if tl.trace {
		fmt.Println("exit JoinOper ", ctx.GetText())
	}

	tok := ctx.GetStart()
	name := strings.ToLower(tok.GetText())
	if _, ok := klStaticData.banJoinMap[name]; ok {
		tl.errorHandler.Add(tok.GetColumn(), fmt.Sprintf("%s join is not supported", name))
	}
}
