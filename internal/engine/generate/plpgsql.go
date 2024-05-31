package generate

import (
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/parse"
)

/*
	This file implements a visitor to generate Postgres compatible SQL and plpgsql
*/

// sqlVisitor creates Postgres compatible SQL from an AST
type sqlGenerator struct {
	parse.UnimplementedSqlVisitor
	// pgSchema is the schema name to prefix to the table names
	pgSchema string
	// numberParameters is a flag that indicates if we should number parameters as $1, $2, etc.,
	// instead of formatting their variable names. It should be set to true if we want to execute
	// SQL directly against postgres, instead of using it in a procedure.
	numberParameters bool
	// orderedParams is the order of parameters in the order they appear in the statement.
	// It is only set if numberParameters is true. For example, the statement SELECT $1, $2
	// would have orderedParams = ["$1", "$2"]
	orderedParams []string
}

func (s *sqlGenerator) VisitExpressionLiteral(p0 *parse.ExpressionLiteral) any {
	str, err := formatPGLiteral(p0.Value)
	if err != nil {
		panic(err)
	}

	if p0.GetTypeCast() != nil {
		pgStr, err := p0.GetTypeCast().PGString()
		if err != nil {
			panic(err)
		}
		str += "::" + pgStr
	}

	return str
}

func (s *sqlGenerator) VisitExpressionFunctionCall(p0 *parse.ExpressionFunctionCall) any {
	str := strings.Builder{}

	args := make([]string, len(p0.Args))
	for i, arg := range p0.Args {
		args[i] = arg.Accept(s).(string)
	}

	// if this is not a built-in function, we need to prefix it with
	// the schema name, since it is a local procedure
	fn, ok := parse.Functions[p0.Name]
	if !ok {
		// if not found, it is a local procedure
		str.WriteString(s.pgSchema)
		str.WriteString(".")
		str.WriteString(p0.Name)
		str.WriteString("(")
		for i, arg := range args {
			if i > 0 {
				str.WriteString(", ")
			}
			str.WriteString(arg)
		}
		str.WriteString(")")
		typeCast(p0, &str)
		return str.String()
	}

	pgFmt, err := fn.PGFormat(args, p0.Distinct, p0.Star)
	if err != nil {
		panic(err)
	}
	str.WriteString(pgFmt)

	typeCast(p0, &str)

	return str.String()
}

// typeCast adds a typecast to the string builder if the typecast is not nil
func typeCast(t interface{ GetTypeCast() *types.DataType }, s *strings.Builder) {
	if t.GetTypeCast() != nil {
		pgStr, err := t.GetTypeCast().PGString()
		if err != nil {
			panic(err)
		}

		s.WriteString("::")
		s.WriteString(pgStr)
	}
}

func (s *sqlGenerator) VisitExpressionForeignCall(p0 *parse.ExpressionForeignCall) any {
	str := strings.Builder{}
	str.WriteString(s.pgSchema)
	str.WriteString(".")
	str.WriteString(formatForeignProcedureName(p0.Name))
	str.WriteString("(")
	for i, arg := range append(p0.ContextualArgs, p0.Args...) {
		if i > 0 {
			str.WriteString(", ")
		}

		str.WriteString(arg.Accept(s).(string))
	}
	str.WriteString(")")

	typeCast(p0, &str)

	return str.String()
}

func (s *sqlGenerator) VisitExpressionVariable(p0 *parse.ExpressionVariable) any {
	// if a user param $, then we need to number it.
	// Vars using @ get set and accessed using postgres's current_setting function
	if s.numberParameters && p0.Prefix == parse.VariablePrefixDollar {
		str := p0.String()

		// if it already exists, we write it as that index.
		for i, v := range s.orderedParams {
			if v == str {
				return "$" + fmt.Sprint(i+1)
			}
		}

		// otherwise, we add it to the list.
		// Postgres uses $1, $2, etc. for numbered parameters.

		s.orderedParams = append(s.orderedParams, str)
		return "$" + fmt.Sprint(len(s.orderedParams))
	}

	str := strings.Builder{}
	str.WriteString(formatVariable(p0))
	typeCast(p0, &str)
	return str.String()
}

func (s *sqlGenerator) VisitExpressionArrayAccess(p0 *parse.ExpressionArrayAccess) any {
	str := strings.Builder{}
	str.WriteString(p0.Array.Accept(s).(string))
	str.WriteString("[")
	str.WriteString(p0.Index.Accept(s).(string))
	str.WriteString("]")
	typeCast(p0, &str)
	return str.String()
}

func (s *sqlGenerator) VisitExpressionMakeArray(p0 *parse.ExpressionMakeArray) any {
	str := strings.Builder{}
	str.WriteString("ARRAY[")
	for i, arg := range p0.Values {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(arg.Accept(s).(string))
	}
	str.WriteString("]")
	typeCast(p0, &str)
	return str.String()
}

func (s *sqlGenerator) VisitExpressionFieldAccess(p0 *parse.ExpressionFieldAccess) any {
	str := strings.Builder{}
	str.WriteString(p0.Record.Accept(s).(string))
	str.WriteString(".")
	str.WriteString(p0.Field)
	typeCast(p0, &str)
	return str.String()
}

func (s *sqlGenerator) VisitExpressionParenthesized(p0 *parse.ExpressionParenthesized) any {
	str := strings.Builder{}
	str.WriteString("(")
	str.WriteString(p0.Inner.Accept(s).(string))
	str.WriteString(")")
	typeCast(p0, &str)
	return str.String()
}

func (s *sqlGenerator) VisitExpressionComparison(p0 *parse.ExpressionComparison) any {
	str := strings.Builder{}
	str.WriteString(p0.Left.Accept(s).(string))
	str.WriteString(" ")
	str.WriteString(string(p0.Operator))
	str.WriteString(" ")
	str.WriteString(p0.Right.Accept(s).(string))
	// compare cannot be typecasted
	return str.String()
}

func (s *sqlGenerator) VisitExpressionLogical(p0 *parse.ExpressionLogical) any {
	str := strings.Builder{}
	str.WriteString(p0.Left.Accept(s).(string))
	str.WriteString(" ")
	str.WriteString(string(p0.Operator))
	str.WriteString(" ")
	str.WriteString(p0.Right.Accept(s).(string))
	// logical cannot be typecasted
	return str.String()
}

func (s *sqlGenerator) VisitExpressionArithmetic(p0 *parse.ExpressionArithmetic) any {
	str := strings.Builder{}
	str.WriteString(p0.Left.Accept(s).(string))
	str.WriteString(" ")
	str.WriteString(string(p0.Operator))
	str.WriteString(" ")
	str.WriteString(p0.Right.Accept(s).(string))
	// cannot be typecasted
	return str.String()
}

func (s *sqlGenerator) VisitExpressionUnary(p0 *parse.ExpressionUnary) any {
	str := strings.Builder{}
	str.WriteString(string(p0.Operator))
	str.WriteString(p0.Expression.Accept(s).(string))
	// cannot be typecasted
	return str.String()
}

func (s *sqlGenerator) VisitExpressionColumn(p0 *parse.ExpressionColumn) any {
	str := strings.Builder{}
	if p0.Table != "" {
		str.WriteString(p0.Table)
		str.WriteString(".")
	}
	str.WriteString(p0.Column)
	typeCast(p0, &str)
	return str.String()
}

func (s *sqlGenerator) VisitExpressionCollate(p0 *parse.ExpressionCollate) any {
	str := strings.Builder{}
	str.WriteString(p0.Expression.Accept(s).(string))
	str.WriteString(" COLLATE ")
	str.WriteString(p0.Collation)
	return str.String()
}

func (s *sqlGenerator) VisitExpressionStringComparison(p0 *parse.ExpressionStringComparison) any {
	str := strings.Builder{}
	str.WriteString(p0.Left.Accept(s).(string))
	str.WriteString(" ")
	if p0.Not {
		str.WriteString("NOT ")
	}
	str.WriteString(string(p0.Operator))
	str.WriteString(" ")
	str.WriteString(p0.Right.Accept(s).(string))
	// compare cannot be typecasted
	return str.String()
}

func (s *sqlGenerator) VisitExpressionIs(p0 *parse.ExpressionIs) any {
	str := strings.Builder{}
	str.WriteString(p0.Left.Accept(s).(string))
	str.WriteString(" IS ")
	if p0.Not {
		str.WriteString("NOT ")
	}
	if p0.Distinct {
		str.WriteString("DISTINCT FROM ")
	}
	str.WriteString(p0.Right.Accept(s).(string))
	// cannot be typecasted
	return str.String()
}

func (s *sqlGenerator) VisitExpressionIn(p0 *parse.ExpressionIn) any {
	str := strings.Builder{}
	str.WriteString(p0.Expression.Accept(s).(string))
	if p0.Not {
		str.WriteString(" NOT")
	}
	str.WriteString(" IN (")
	if len(p0.List) > 0 {
		for i, arg := range p0.List {
			if i > 0 {
				str.WriteString(", ")
			}

			str.WriteString(arg.Accept(s).(string))
		}
	} else if p0.Subquery != nil {
		str.WriteString(p0.Subquery.Accept(s).(string))
	} else {
		panic("IN must specify list or subquery")
	}
	str.WriteString(")")

	return str.String()
}

func (s *sqlGenerator) VisitExpressionBetween(p0 *parse.ExpressionBetween) any {
	str := strings.Builder{}
	str.WriteString(p0.Expression.Accept(s).(string))
	if p0.Not {
		str.WriteString(" NOT")
	}
	str.WriteString(" BETWEEN ")

	str.WriteString(p0.Lower.Accept(s).(string))
	str.WriteString(" AND ")
	str.WriteString(p0.Upper.Accept(s).(string))

	return str.String()
}

func (s *sqlGenerator) VisitExpressionSubquery(p0 *parse.ExpressionSubquery) any {
	str := strings.Builder{}
	if p0.Exists {
		if p0.Not {
			str.WriteString("NOT ")
		}
		str.WriteString("EXISTS ")
	}

	str.WriteString("(")
	str.WriteString(p0.Subquery.Accept(s).(string))
	str.WriteString(")")
	typeCast(p0, &str)
	return str.String()
}

func (s *sqlGenerator) VisitExpressionCase(p0 *parse.ExpressionCase) any {
	str := strings.Builder{}
	str.WriteString("CASE")
	if p0.Case != nil {
		str.WriteString(" ")
		str.WriteString(p0.Case.Accept(s).(string))
	}
	for _, whenThen := range p0.WhenThen {
		str.WriteString("\n WHEN ")
		str.WriteString(whenThen[0].Accept(s).(string))
		str.WriteString("\n THEN ")
		str.WriteString(whenThen[1].Accept(s).(string))
	}
	if p0.Else != nil {
		str.WriteString("\n ELSE ")
		str.WriteString(p0.Else.Accept(s).(string))
	}
	str.WriteString("\n END")
	return str.String()
}

func (s *sqlGenerator) VisitCommonTableExpression(p0 *parse.CommonTableExpression) any {
	str := strings.Builder{}
	str.WriteString(p0.Name)
	if p0.Columns != nil {
		str.WriteString(" (")
		for i, col := range p0.Columns {
			if i > 0 {
				str.WriteString(", ")
			}
			str.WriteString(col)
		}
		str.WriteString(")")
	}
	str.WriteString(" AS (")
	str.WriteString(p0.Query.Accept(s).(string))
	str.WriteString(")")
	return str.String()
}

func (s *sqlGenerator) VisitSQLStatement(p0 *parse.SQLStatement) any {
	str := strings.Builder{}
	for i, cte := range p0.CTEs {
		if i > 0 {
			str.WriteString(", ")
		}
		if i == 0 {
			str.WriteString("WITH ")
		}
		str.WriteString(cte.Accept(s).(string))
	}
	str.WriteString("\n")

	str.WriteString(p0.SQL.Accept(s).(string))

	return str.String()
}

func (s *sqlGenerator) VisitSelectStatement(p0 *parse.SelectStatement) any {
	str := strings.Builder{}
	for i, core := range p0.SelectCores {
		if i > 0 {
			str.WriteString(" ")
			str.WriteString(string(p0.CompoundOperators[i-1]))
			str.WriteString(" ")
		}
		str.WriteString(core.Accept(s).(string))
		str.WriteString("\n")
	}

	for i, order := range p0.Ordering {
		if i == 0 {
			str.WriteString("ORDER BY ")
		} else {
			str.WriteString(", ")
		}

		str.WriteString(order.Accept(s).(string))
	}

	if p0.Limit != nil {
		str.WriteString(" LIMIT ")
		str.WriteString(p0.Limit.Accept(s).(string))
	}

	if p0.Offset != nil {
		str.WriteString(" OFFSET ")
		str.WriteString(p0.Offset.Accept(s).(string))
	}

	return str.String()
}

func (s *sqlGenerator) VisitSelectCore(p0 *parse.SelectCore) any {
	str := strings.Builder{}
	str.WriteString("SELECT ")
	if p0.Distinct {
		str.WriteString("DISTINCT ")
	}

	for i, resultColumn := range p0.Columns {
		if i > 0 {
			str.WriteString(", ")
		}
		str.WriteString(resultColumn.Accept(s).(string))
	}

	if p0.From != nil {
		str.WriteString("\nFROM ")
		str.WriteString(p0.From.Accept(s).(string))
	}

	for _, join := range p0.Joins {
		str.WriteString("\n")
		str.WriteString(join.Accept(s).(string))
	}

	if p0.Where != nil {
		str.WriteString("\nWHERE ")
		str.WriteString(p0.Where.Accept(s).(string))
	}

	if p0.GroupBy != nil {
		str.WriteString("\nGROUP BY ")
		for i, groupBy := range p0.GroupBy {
			if i > 0 {
				str.WriteString(", ")
			}
			str.WriteString(groupBy.Accept(s).(string))
		}

		if p0.Having != nil {
			str.WriteString("\nHAVING ")
			str.WriteString(p0.Having.Accept(s).(string))
		}
	}

	return str.String()
}

func (s *sqlGenerator) VisitResultColumnExpression(p0 *parse.ResultColumnExpression) any {
	str := strings.Builder{}
	str.WriteString(p0.Expression.Accept(s).(string))
	if p0.Alias != "" {
		str.WriteString(" AS ")
		str.WriteString(p0.Alias)
	}
	return str.String()
}

func (s *sqlGenerator) VisitResultColumnWildcard(p0 *parse.ResultColumnWildcard) any {
	str := strings.Builder{}
	if p0.Table != "" {
		str.WriteString(p0.Table)
		str.WriteString(".")
	}
	str.WriteString("*")
	return str.String()
}

func (s *sqlGenerator) VisitRelationTable(p0 *parse.RelationTable) any {
	str := strings.Builder{}
	if s.pgSchema != "" {
		str.WriteString(s.pgSchema)
		str.WriteString(".")
	}
	str.WriteString(p0.Table)
	if p0.Alias != "" {
		str.WriteString(" AS ")
		str.WriteString(p0.Alias)
	}
	return str.String()
}

func (s *sqlGenerator) VisitRelationSubquery(p0 *parse.RelationSubquery) any {
	str := strings.Builder{}
	str.WriteString("(")
	str.WriteString(p0.Subquery.Accept(s).(string))
	str.WriteString(") ")
	if p0.Alias != "" {
		str.WriteString("AS ")
		str.WriteString(p0.Alias)
	}
	return str.String()
}

func (s *sqlGenerator) VisitRelationFunctionCall(p0 *parse.RelationFunctionCall) any {
	str := strings.Builder{}
	str.WriteString(p0.FunctionCall.Accept(s).(string))
	str.WriteString(" ")
	if p0.Alias != "" {
		str.WriteString("AS ")
		str.WriteString(p0.Alias)
	}
	return str.String()
}

func (s *sqlGenerator) VisitJoin(p0 *parse.Join) any {
	str := strings.Builder{}
	str.WriteString(string(p0.Type))
	str.WriteString(" JOIN ")
	str.WriteString(p0.Relation.Accept(s).(string))
	// we do not worry about on being nil, since Kwil
	// forces the user to specify the join condition
	// to prevent cartesian products
	str.WriteString(" ON ")
	str.WriteString(p0.On.Accept(s).(string))
	return str.String()
}

func (s *sqlGenerator) VisitUpdateStatement(p0 *parse.UpdateStatement) any {
	str := strings.Builder{}
	str.WriteString("UPDATE ")
	if s.pgSchema != "" {
		str.WriteString(s.pgSchema)
		str.WriteString(".")
	}
	str.WriteString(p0.Table)
	if p0.Alias != "" {
		str.WriteString(" AS ")
		str.WriteString(p0.Alias)
	}
	str.WriteString("\nSET ")
	for i, set := range p0.SetClause {
		if i > 0 {
			str.WriteString(",\n")
		}
		str.WriteString(set.Accept(s).(string))
	}

	if p0.From != nil {
		str.WriteString("\nFROM ")
		str.WriteString(p0.From.Accept(s).(string))
	}

	for _, join := range p0.Joins {
		str.WriteString("\n")
		str.WriteString(join.Accept(s).(string))
	}

	if p0.Where != nil {
		str.WriteString("\nWHERE ")
		str.WriteString(p0.Where.Accept(s).(string))
	}

	return str.String()
}

func (s *sqlGenerator) VisitUpdateSetClause(p0 *parse.UpdateSetClause) any {
	str := strings.Builder{}
	str.WriteString(p0.Column)
	str.WriteString(" = ")
	str.WriteString(p0.Value.Accept(s).(string))
	return str.String()
}

func (s *sqlGenerator) VisitDeleteStatement(p0 *parse.DeleteStatement) any {
	str := strings.Builder{}
	str.WriteString("DELETE FROM ")

	if s.pgSchema != "" {
		str.WriteString(s.pgSchema)
		str.WriteString(".")
	}

	str.WriteString(p0.Table)
	if p0.Alias != "" {
		str.WriteString(" AS ")
		str.WriteString(p0.Alias)
	}

	if p0.From != nil {
		str.WriteString("\nFROM ")
		str.WriteString(p0.From.Accept(s).(string))
	}

	for _, join := range p0.Joins {
		str.WriteString("\n")
		str.WriteString(join.Accept(s).(string))
	}

	if p0.Where != nil {
		str.WriteString("\nWHERE ")
		str.WriteString(p0.Where.Accept(s).(string))
	}

	return str.String()
}

func (s *sqlGenerator) VisitInsertStatement(p0 *parse.InsertStatement) any {
	str := strings.Builder{}
	str.WriteString("INSERT INTO ")
	if s.pgSchema != "" {
		str.WriteString(s.pgSchema)
		str.WriteString(".")
	}

	str.WriteString(p0.Table)
	if p0.Alias != "" {
		str.WriteString(" AS ")
		str.WriteString(p0.Alias)
	}
	if len(p0.Columns) > 0 {
		str.WriteString(" (")

		for i, col := range p0.Columns {
			if i > 0 {
				str.WriteString(", ")
			}
			str.WriteString(col)
		}

		str.WriteString(") ")
	}
	str.WriteString("\nVALUES ")

	for i, val := range p0.Values {
		if i > 0 {
			str.WriteString(",")
		}
		str.WriteString("\n(")
		for j, v := range val {
			if j > 0 {
				str.WriteString(", ")
			}
			str.WriteString(v.Accept(s).(string))
		}
		str.WriteString(")")
	}

	if p0.Upsert != nil {
		str.WriteString("\n")
		str.WriteString(p0.Upsert.Accept(s).(string))
	}

	return str.String()
}

func (s *sqlGenerator) VisitUpsertClause(p0 *parse.UpsertClause) any {
	str := strings.Builder{}
	str.WriteString("ON CONFLICT ")
	if len(p0.ConflictColumns) > 0 {
		str.WriteString("(")
		for i, col := range p0.ConflictColumns {
			if i > 0 {
				str.WriteString(", ")
			}
			str.WriteString(col)
		}
		str.WriteString(")\n")

		if p0.ConflictWhere != nil {
			str.WriteString("WHERE ")
			str.WriteString(p0.ConflictWhere.Accept(s).(string))
			str.WriteString("\n")
		}
	}

	str.WriteString("DO ")
	if p0.DoUpdate == nil {
		str.WriteString("NOTHING")
	} else {
		str.WriteString("UPDATE SET")
		for i, set := range p0.DoUpdate {
			if i > 0 {
				str.WriteString(",")
			}
			str.WriteString("\n	")
			str.WriteString(set.Accept(s).(string))
		}

		if p0.UpdateWhere != nil {
			str.WriteString("\nWHERE ")
			str.WriteString(p0.UpdateWhere.Accept(s).(string))
		}
	}

	return str.String()
}

func (s *sqlGenerator) VisitOrderingTerm(p0 *parse.OrderingTerm) any {
	str := strings.Builder{}
	str.WriteString(p0.Expression.Accept(s).(string))

	if p0.Order != "" {
		str.WriteString(" ")
		str.WriteString(string(p0.Order))
	}

	if p0.Nulls != "" {
		str.WriteString(" NULLS ")
		str.WriteString(string(p0.Nulls))
	}

	return str.String()
}

// procedureGenerator is a visitor that generates plpgsql code.
type procedureGenerator struct {
	sqlGenerator
	// anonymousReceivers counts the amount of anonymous receivers
	// we should declare. This will be cross-referenced with the
	// analyzer to ensure we declare the correct amount.
	anonymousReceivers int
	// procedure is the procedure we are generating code for
	procedure *types.Procedure
}

var _ parse.ProcedureVisitor = &procedureGenerator{}

func (p *procedureGenerator) VisitProcedureStmtDeclaration(p0 *parse.ProcedureStmtDeclaration) any {
	// plpgsql declares variables at the top of the procedure
	return ""
}

func (p *procedureGenerator) VisitProcedureStmtAssignment(p0 *parse.ProcedureStmtAssign) any {
	varName := p0.Variable.Accept(p).(string)
	return varName + " := " + p0.Value.Accept(p).(string) + ";\n"
}

func (p *procedureGenerator) VisitProcedureStmtCall(p0 *parse.ProcedureStmtCall) any {
	call := p0.Call.Accept(p).(string)

	if len(p0.Receivers) == 0 {
		return "PERFORM " + call + ";\n"
	}

	s := strings.Builder{}
	s.WriteString("SELECT * INTO ")

	for i, rec := range p0.Receivers {
		if i > 0 {
			s.WriteString(", ")
		}
		if rec == nil {
			s.WriteString(formatAnonymousReceiver(p.anonymousReceivers))
			p.anonymousReceivers++
		} else {
			s.WriteString(rec.Accept(p).(string))
		}
	}

	s.WriteString(" FROM ")
	s.WriteString(call)
	s.WriteString(";\n")
	return s.String()
}

func (p *procedureGenerator) VisitProcedureStmtForLoop(p0 *parse.ProcedureStmtForLoop) any {
	s := strings.Builder{}
	// if we are iterating over an array, the syntax is different
	switch v := p0.LoopTerm.(type) {
	case *parse.LoopTermRange, *parse.LoopTermSQL:
		s.WriteString("FOR ")
		s.WriteString(p0.Receiver.Accept(p).(string))
		s.WriteString(" IN ")
		s.WriteString(p0.LoopTerm.Accept(p).(string))
	case *parse.LoopTermVariable:
		s.WriteString("FOREACH ")
		s.WriteString(p0.Receiver.Accept(p).(string))
		s.WriteString(" IN ")
		s.WriteString(p0.LoopTerm.Accept(p).(string))
	default:
		panic("unknown loop term type: " + fmt.Sprintf("%T", v))
	}

	s.WriteString(" LOOP\n")

	for _, stmt := range p0.Body {
		s.WriteString(stmt.Accept(p).(string))
	}

	s.WriteString("END LOOP;\n")

	return s.String()
}

func (p *procedureGenerator) VisitLoopTermRange(p0 *parse.LoopTermRange) any {
	s := strings.Builder{}
	s.WriteString(p0.Start.Accept(p).(string))
	s.WriteString("..")
	s.WriteString(p0.End.Accept(p).(string))
	return s.String()
}

func (p *procedureGenerator) VisitLoopTermSQL(p0 *parse.LoopTermSQL) any {
	return p0.Statement.Accept(p).(string)
}

func (p *procedureGenerator) VisitLoopTermVariable(p0 *parse.LoopTermVariable) any {
	// we use coalesce here so that we do not error when looping on null arrays
	return fmt.Sprintf("ARRAY COALESCE(%s, '{}')", p0.Variable.Accept(p).(string))
}

func (p *procedureGenerator) VisitProcedureStmtIf(p0 *parse.ProcedureStmtIf) any {
	s := strings.Builder{}
	for i, clause := range p0.IfThens {
		if i == 0 {
			s.WriteString("IF ")
		} else {
			s.WriteString("ELSIF ")
		}

		s.WriteString(clause.Accept(p).(string))
	}

	if p0.Else != nil {
		s.WriteString("ELSE\n")
		for _, stmt := range p0.Else {

			s.WriteString(stmt.Accept(p).(string))
		}
	}

	s.WriteString("END IF;\n")
	return s.String()
}

func (p *procedureGenerator) VisitIfThen(p0 *parse.IfThen) any {
	s := strings.Builder{}
	s.WriteString(p0.If.Accept(p).(string))
	s.WriteString(" THEN\n")
	for _, stmt := range p0.Then {
		s.WriteString(stmt.Accept(p).(string))
	}

	return s.String()
}

func (p *procedureGenerator) VisitProcedureStmtSQL(p0 *parse.ProcedureStmtSQL) any {
	return p0.SQL.Accept(p).(string) + ";\n"
}

func (p *procedureGenerator) VisitProcedureStmtBreak(p0 *parse.ProcedureStmtBreak) any {
	return "EXIT;\n"
}

func (p *procedureGenerator) VisitProcedureStmtReturn(p0 *parse.ProcedureStmtReturn) any {
	if p0.SQL != nil {
		return "RETURN QUERY " + p0.SQL.Accept(p).(string) + ";\n"
	}

	s := strings.Builder{}
	for i, expr := range p0.Values {
		s.WriteString(formatReturnVar(i))
		s.WriteString(" := ")
		s.WriteString(expr.Accept(p).(string))
		s.WriteString(";\n")
	}

	s.WriteString("RETURN;")
	return s.String()
}

func (p *procedureGenerator) VisitProcedureStmtReturnNext(p0 *parse.ProcedureStmtReturnNext) any {
	s := strings.Builder{}
	for i, expr := range p0.Values {
		// we do not format the return var for return next, but instead
		// assign it to the column name directly
		s.WriteString(p.procedure.Returns.Fields[i].Name)
		s.WriteString(" := ")
		s.WriteString(expr.Accept(p).(string))
		s.WriteString(";\n")
	}

	s.WriteString("RETURN NEXT;\n")
	return s.String()
}

// formatPGLiteral formats a literal for user in postgres.
func formatPGLiteral(value any) (string, error) {
	str := strings.Builder{}
	switch v := value.(type) {
	case string: // for text type
		// escape single quotes
		str.WriteString("'")
		str.WriteString(strings.ReplaceAll(v, "'", "''"))
		str.WriteString("'")
	case int64, int, int32: // for int type
		str.WriteString(fmt.Sprint(v))
	case types.UUID:
		str.WriteString(v.String())
	case *types.Uint256:
		str.WriteString(v.String())
	case *decimal.Decimal:
		str.WriteString(v.String())
	case bool: // for bool type
		if v {
			str.WriteString("true")
		} else {
			str.WriteString("false")
		}
	case []byte: // for blob type: https://dba.stackexchange.com/questions/203358/how-do-i-write-a-hex-literal-in-postgresql
		str.WriteString(fmt.Sprintf("E'\\\\x%x'", v))
	case nil:
		str.WriteString("NULL")
	case fmt.Stringer:
		str.WriteString(v.String())
	default:
		return "", fmt.Errorf("unsupported literal type: %T", v)
	}

	return str.String(), nil
}

// FormatProcedureName formats a procedure name for usage in postgres. This
// simply prepends the name with _fp_
func formatForeignProcedureName(name string) string {
	return "_fp_" + name
}

// formatAnonymousReceiver creates a plpgsql variable name for anonymous receivers.
func formatAnonymousReceiver(index int) string {
	return fmt.Sprintf("_anon_%d", index)
}

// formatReturnVar formats a return variable name for usage in postgres.
func formatReturnVar(i int) string {
	return fmt.Sprintf("_out_%d", i)
}

// formatVariable formats an expression variable for usage in postgres.
func formatVariable(e *parse.ExpressionVariable) string {
	switch e.Prefix {
	case parse.VariablePrefixDollar:
		return formatParameterName(e.Name)
	case parse.VariablePrefixAt:
		return formatContextualVariableName(e.Name)
	default:
		// should never happen
		panic("invalid variable prefix: " + string(e.Prefix))
	}
}

// formatParameterName formats a parameter name for usage in postgres. This
// simply prepends the name with _param_. It expects the name does not have
// the $ prefix
func formatParameterName(name string) string {
	return "_param_" + name
}

// formatContextualVariableName formats a contextual variable name for usage in postgres.
// This uses the current_setting function to get the value of the variable. It also
// removes the @ prefix. If the type is not a text type, it will also type cast it.
// The type casting is necessary since current_setting returns all values as text.
func formatContextualVariableName(name string) string {
	str := fmt.Sprintf("current_setting('%s.%s')", PgSessionPrefix, name)

	dataType, ok := parse.SessionVars[name]
	if !ok {
		panic("unknown contextual variable: " + name)
	}

	switch dataType {
	case types.BlobType:
		return fmt.Sprintf("decode(%s, 'base64')", str)
	case types.IntType:
		return fmt.Sprintf("%s::int8", str)
	case types.BoolType:
		return fmt.Sprintf("%s::bool", str)
	case types.UUIDType:
		return fmt.Sprintf("%s::uuid", str)
	case types.TextType:
		return str
	default:
		panic("unallowed contextual variable type: " + dataType.String())
	}
}
