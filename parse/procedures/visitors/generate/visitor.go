package generate

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/parse/metadata"
	parser "github.com/kwilteam/kwil-db/parse/procedures/parser"
	"github.com/kwilteam/kwil-db/parse/sql/tree"
	"github.com/kwilteam/kwil-db/parse/util"
)

type generatorVisitor struct {
	// variables tracks all variables that the visitor comes across.
	// it will assign the variables a unique name.
	variables map[string]*types.DataType
	// currentProcedure is the current procedure being generated.
	currentProcedure *types.Procedure

	// anonymousReceiverCount is the count of anonymous receivers.
	// It is used to ensure unique names for anonymous receivers.
	// Anonymoud receivers are used when a procedure returns like:
	// $return1, _, $return3 := ...
	anonymousReceiverCount int
	// anonymousTypes are the types of the anonymous receivers.
	anonymousTypes []*types.DataType

	// returnedVariables holds the variables that are returned by the procedure,
	// if any.
	returnedVariables []*types.NamedType
	// loopTargets are the variables that are loop targets.
	loopTargets []string
	// inLoop is true if the visitor is currently in a loop.
	inLoop bool
	// pgSchemaName is the name of the postgres schema.
	pgSchemaName string
}

var _ parser.Visitor = &generatorVisitor{}

func (g *generatorVisitor) VisitExpressionArithmetic(p0 *parser.ExpressionArithmetic) any {
	str := strings.Builder{}

	str.WriteString(p0.Left.Accept(g).(string))
	str.WriteString(" ")
	str.WriteString(string(p0.Operator))
	str.WriteString(" ")
	str.WriteString(p0.Right.Accept(g).(string))

	return str.String()
}

func (g *generatorVisitor) VisitExpressionArrayAccess(p0 *parser.ExpressionArrayAccess) any {
	return fmt.Sprintf("%s[%s]", p0.Target.Accept(g).(string), p0.Index.Accept(g).(string))
}

func (g *generatorVisitor) VisitExpressionBlobLiteral(p0 *parser.ExpressionBlobLiteral) any {
	return fmt.Sprintf("'\\%s'", hex.EncodeToString(p0.Value))
}

func (g *generatorVisitor) VisitExpressionBooleanLiteral(p0 *parser.ExpressionBooleanLiteral) any {
	return fmt.Sprintf("%t", p0.Value)
}

func (g *generatorVisitor) VisitExpressionCall(p0 *parser.ExpressionCall) any {

	inputs := make([]string, len(p0.Arguments))
	for i, arg := range p0.Arguments {
		inputs[i] = arg.Accept(g).(string)
	}

	// if it is not a function, it is a procedure,
	// and we need to prefix it with the schema name.
	funcDef, ok := metadata.Functions[p0.Name]
	if ok {
		return funcDef.PGFormat(inputs)
	}

	str := strings.Builder{}
	str.WriteString(g.pgSchemaName)
	str.WriteString(".")
	str.WriteString(p0.Name)

	str.WriteString("(")
	for i, in := range inputs {
		if i > 0 {
			str.WriteString(", ")
		}

		str.WriteString(in)
	}

	str.WriteString(")")

	return str.String()
}

func (g *generatorVisitor) VisitExpressionForeignCall(p0 *parser.ExpressionForeignCall) any {
	// foreign procedure calls are simply locally defined functions.
	// They are prefixed with _fp_ to hide them from the user.
	// They take the two contextual variables as the first two arguments.
	// The rest of the arguments are the same as the procedure definition.

	str := strings.Builder{}

	str.WriteString(util.FormatForeignProcedureName(p0.Name))

	str.WriteString("(")
	// appending the context args and regular args.
	for i, in := range append(p0.ContextArgs, p0.Arguments...) {
		if i > 0 {
			str.WriteString(", ")
		}

		str.WriteString(in.Accept(g).(string))
	}

	str.WriteString(")")

	return str.String()
}

func (g *generatorVisitor) VisitExpressionComparison(p0 *parser.ExpressionComparison) any {
	str := strings.Builder{}

	str.WriteString(p0.Left.Accept(g).(string))
	str.WriteString(" ")
	str.WriteString(string(p0.Operator))
	str.WriteString(" ")
	str.WriteString(p0.Right.Accept(g).(string))

	return str.String()
}

func (g *generatorVisitor) VisitExpressionFieldAccess(p0 *parser.ExpressionFieldAccess) any {
	return fmt.Sprintf("%s.%s", p0.Target.Accept(g).(string), p0.Field)
}

func (g *generatorVisitor) VisitExpressionIntLiteral(p0 *parser.ExpressionIntLiteral) any {
	return fmt.Sprintf("%d", p0.Value)
}

func (g *generatorVisitor) VisitExpressionMakeArray(p0 *parser.ExpressionMakeArray) any {
	str := strings.Builder{}
	str.WriteString("ARRAY[")
	for i, val := range p0.Values {
		if i > 0 {
			str.WriteString(", ")
		}

		str.WriteString(val.Accept(g).(string))
	}

	str.WriteString("]")

	return str.String()
}

func (g *generatorVisitor) VisitExpressionNullLiteral(p0 *parser.ExpressionNullLiteral) any {
	return "NULL"
}

func (g *generatorVisitor) VisitExpressionParenthesized(p0 *parser.ExpressionParenthesized) any {
	return fmt.Sprintf("(%s)", p0.Expression.Accept(g).(string))
}

func (g *generatorVisitor) VisitExpressionTextLiteral(p0 *parser.ExpressionTextLiteral) any {

	// escape single quotes
	p0.Value = strings.ReplaceAll(p0.Value, "'", "''")

	return fmt.Sprintf("'%s'", p0.Value)
}

func (g *generatorVisitor) VisitExpressionVariable(p0 *parser.ExpressionVariable) any {
	return p0.Name
}

func (g *generatorVisitor) VisitLoopTargetCall(p0 *parser.LoopTargetCall) any {
	return fmt.Sprintf("SELECT * FROM %s", p0.Call.Accept(g).(string))
}

func (g *generatorVisitor) VisitLoopTargetRange(p0 *parser.LoopTargetRange) any {
	return fmt.Sprintf("%s..%s", p0.Start.Accept(g).(string), p0.End.Accept(g).(string))
}

func (g *generatorVisitor) VisitLoopTargetVariable(p0 *parser.LoopTargetVariable) any {
	return fmt.Sprintf("ARRAY %s", p0.Variable.Accept(g).(string))
}

func (g *generatorVisitor) VisitLoopTargetSQL(p0 *parser.LoopTargetSQL) any {

	stmt, err := tree.SafeToSQL(p0.Statement)
	if err != nil {
		panic(fmt.Sprintf("error converting statement to sql: %v", err))
	}

	// since our ToSQL returns a semi-colon, we need to remove it.
	return strings.TrimSuffix(stmt, "; ")
}

func (g *generatorVisitor) VisitStatementProcedureCall(p0 *parser.StatementProcedureCall) any {
	// in order to handle variadic returns, we will need
	// to trim the last paren from the call, and add
	// our out variables.
	call := p0.Call.Accept(g).(string)
	if len(p0.Variables) == 0 {
		return fmt.Sprintf("PERFORM %s;", call)
	}

	if !strings.HasSuffix(call, ")") {
		panic("internal error generating procedure call with return values")
	}

	selectInto := strings.Builder{}
	for i, v := range p0.Variables {
		if i > 0 {
			selectInto.WriteString(", ")
		}

		// if v is nil, it is an anonymous receiver.
		if v == nil {
			// if we do not have enough anonymous types, we should panic.
			// this is an internal bug.
			if len(g.anonymousTypes) <= g.anonymousReceiverCount {
				panic("internal error: not enough anonymous types")
			}

			// use double underscore for collision avoidance
			ident := fmt.Sprintf("__anon_%d", g.anonymousReceiverCount)
			g.anonymousReceiverCount++

			g.variables[ident] = g.anonymousTypes[g.anonymousReceiverCount-1]
			selectInto.WriteString(ident)
		} else {
			selectInto.WriteString(*v)
		}
	}

	return fmt.Sprintf("SELECT * INTO %s FROM %s;", selectInto.String(), call)
}

func (g *generatorVisitor) VisitStatementBreak(p0 *parser.StatementBreak) any {
	if !g.inLoop {
		panic("break statement outside of loop")
	}
	return "EXIT;"
}

func (g *generatorVisitor) VisitStatementForLoop(p0 *parser.StatementForLoop) any {

	_, ok := g.variables[p0.Variable]
	if ok {
		panic(fmt.Sprintf("variable %s already declared", p0.Variable))
	}

	// we need to declare the loop targets as RECORD,
	// unless they are looping over a range or array.
	str := strings.Builder{}
	switch target := p0.Target.(type) {
	case *parser.LoopTargetVariable:
		// get the target variable type
		varType, ok := g.variables[target.Variable.Name]
		if !ok {
			panic(fmt.Sprintf("variable %s not found", p0.Variable))
		}

		// varType is an array, we declare a variable
		// of the same scalar type.
		if !varType.IsArray {
			// this should be redundant with the type checker, I think I can delete?
			panic(fmt.Sprintf("expected array type, got %s", varType))
		}

		g.variables[p0.Variable] = &types.DataType{
			Name: varType.Name,
		}

		str.WriteString("FOREACH ")
		str.WriteString(p0.Variable)
		str.WriteString(" IN ")
		str.WriteString(p0.Target.Accept(g).(string))
	case *parser.LoopTargetCall:
		g.loopTargets = append(g.loopTargets, p0.Variable)

		str.WriteString("FOR ")
		str.WriteString(p0.Variable)
		str.WriteString(" IN ")
		str.WriteString(p0.Target.Accept(g).(string))
	case *parser.LoopTargetRange:
		// declare an int type for the receiver
		g.variables[p0.Variable] = types.IntType

		str.WriteString("FOR ")
		str.WriteString(p0.Variable)
		str.WriteString(" IN ")
		str.WriteString(p0.Target.Accept(g).(string))
	case *parser.LoopTargetSQL:
		g.loopTargets = append(g.loopTargets, p0.Variable)

		str.WriteString("FOR ")
		str.WriteString(p0.Variable)
		str.WriteString(" IN ")
		str.WriteString(p0.Target.Accept(g).(string))
	default:
		panic(fmt.Sprintf("unexpected loop target type: %T", p0.Target))
	}

	str.WriteString(" LOOP\n")

	for _, stmt := range p0.Body {
		// we need to check if we are in a loop.
		// if we are the outermost loop, we need
		// to set inLoop to true, and set it back
		// to false when we are done.
		if !g.inLoop {
			g.inLoop = true

			defer func() {
				g.inLoop = false
			}()
		}

		str.WriteString(stmt.Accept(g).(string))
		str.WriteString("\n")
	}

	str.WriteString("END LOOP;")

	return str.String()
}

func (g *generatorVisitor) VisitStatementIf(p0 *parser.StatementIf) any {
	str := strings.Builder{}

	mainIf := p0.IfThens[0]
	str.WriteString("IF ")
	str.WriteString(mainIf.If.Accept(g).(string))
	str.WriteString(" THEN\n")
	for _, stmt := range mainIf.Then {
		str.WriteString(stmt.Accept(g).(string))
		str.WriteString("\n")
	}

	for _, ifthen := range p0.IfThens[1:] {
		str.WriteString("ELSIF ")
		str.WriteString(ifthen.If.Accept(g).(string))
		str.WriteString(" THEN\n")
		for _, stmt := range ifthen.Then {
			str.WriteString(stmt.Accept(g).(string))
			str.WriteString("\n")
		}
	}

	if p0.Else != nil {
		str.WriteString("ELSE\n")
		for _, stmt := range p0.Else {
			str.WriteString(stmt.Accept(g).(string))
			str.WriteString("\n")
		}
	}

	str.WriteString("END IF;")

	return str.String()
}

func (g *generatorVisitor) VisitStatementReturn(p0 *parser.StatementReturn) any {
	if p0.SQL != nil {
		stmt, err := tree.SafeToSQL(p0.SQL)
		if err != nil {
			panic(fmt.Sprintf("error converting statement to sql: %v", err))
		}

		return fmt.Sprintf("RETURN QUERY %s", stmt)
	}

	str := strings.Builder{}

	if p0.Values != nil {
		if len(p0.Values) != len(g.currentProcedure.Returns.Fields) {
			panic(fmt.Sprintf("expected %d return values, got %d", len(g.currentProcedure.Returns.Fields), len(p0.Values)))
		}

		// redeclare returned variables in case there
		// are multiple return paths due to if statements.
		g.returnedVariables = make([]*types.NamedType, 0, len(p0.Values))

		outVars := make([]*types.NamedType, len(p0.Values))
		for i, val := range p0.Values {
			s, ok := val.Accept(g).(string)
			if !ok {
				panic(fmt.Sprintf("unexpected return value type: %T", p0.Values))
			}

			// we need to declare out vars to be able to return them
			outVars[i] = &types.NamedType{
				Name: fmt.Sprintf("_out_%d", i),
				Type: g.currentProcedure.Returns.Fields[i].Type,
			}

			// we will assign the value to the out var
			str.WriteString(fmt.Sprintf("%s := %s;\n", outVars[i].Name, s))
			g.returnedVariables = append(g.returnedVariables, &types.NamedType{
				Name: outVars[i].Name,
				Type: outVars[i].Type,
			})
		}
	}

	str.WriteString("RETURN;")
	return str.String()
}

func (g *generatorVisitor) VisitStatementReturnNext(p0 *parser.StatementReturnNext) interface{} {
	if !g.inLoop {
		panic("RETURN NEXT statement outside of loop")
	}

	str := strings.Builder{}
	for i, expr := range p0.Returns {
		str.WriteString(fmt.Sprintf("%s := %s;\n", g.currentProcedure.Returns.Fields[i].Name, expr.Accept(g).(string)))
	}

	str.WriteString("RETURN NEXT;")

	return str.String()
}

func (g *generatorVisitor) VisitStatementSQL(p0 *parser.StatementSQL) any {
	stmt, err := tree.SafeToSQL(p0.Statement)
	if err != nil {
		panic(fmt.Sprintf("error converting statement to sql: %v", err))
	}

	return stmt
}

func (g *generatorVisitor) VisitStatementVariableAssignment(p0 *parser.StatementVariableAssignment) any {

	// we will ensure the variable was already declared,
	// and then add 1 statement.

	_, ok := g.variables[p0.Name]
	if !ok {
		panic(fmt.Sprintf("variable %s not found", p0.Name))
	}

	res := p0.Value.Accept(g).(string)

	return fmt.Sprintf("%s := %s;", p0.Name, res)
}

func (g *generatorVisitor) VisitStatementVariableAssignmentWithDeclaration(p0 *parser.StatementVariableAssignmentWithDeclaration) any {

	// we will add 1 statement, since this is a declaration and an assignment.

	_, ok := g.variables[p0.Name]
	if ok {
		panic(fmt.Sprintf("variable %s already exists", p0.Name))
	}

	g.variables[p0.Name] = p0.Type

	res := p0.Value.Accept(g).(string)

	return fmt.Sprintf("%s := %s;", p0.Name, res)
}

func (g *generatorVisitor) VisitStatementVariableDeclaration(p0 *parser.StatementVariableDeclaration) any {
	// we will not add any statements, since this is just a declaration.
	_, ok := g.variables[p0.Name]
	if ok {
		panic(fmt.Sprintf("variable %s already exists", p0.Name))
	}

	g.variables[p0.Name] = p0.Type

	return ""
}
