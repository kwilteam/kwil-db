package tree

import (
	sqlwriter "github.com/kwilteam/kwil-db/parse/sql/tree/sql-writer"
)

type distinctable interface {
	SQLFunction
	stringDistinct(exprs ...Expression) string
}

type AggregateFunc struct {
	*node

	AnySQLFunction
}

func (s *AggregateFunc) Accept(v AstVisitor) any {
	return v.VisitAggregateFunc(s)
}

func (s *AggregateFunc) ToSQL() string {
	return s.ToString()
}

func (s *AggregateFunc) Walk(w AstWalker) error {
	return run(
		w.EnterAggregateFunc(s),
		w.ExitAggregateFunc(s),
	)
}

// stringDistinct returns the string representation of the function with the
// given arguments, prepended by the DISTINCT keyword.
func (s *AggregateFunc) stringDistinct(exprs ...Expression) string {
	if s.Min > 0 && len(exprs) < int(s.Min) {
		panic("too few arguments for function " + s.FunctionName)
	}
	if s.Max > 0 && len(exprs) > int(s.Max) {
		panic("too many arguments for function " + s.FunctionName)
	}

	if len(exprs) == 0 {
		return s.stringAll()
	}

	return s.buildFunctionString(func(stmt *sqlwriter.SqlWriter) {
		stmt.Token.Distinct()

		stmt.WriteList(len(exprs), func(i int) {
			stmt.WriteString(exprs[i].ToSQL())
		})
	})
}

func (s *AggregateFunc) ToString(exprs ...Expression) string {
	if s.distinct {
		return s.stringDistinct(exprs...)
	}
	return s.string(exprs...)
}

func NewAggregateFunctionWithGetter(name string, min uint8, max uint8, distinct bool) SQLFunctionGetter {
	return func(pos *Position) SQLFunction {
		return &AggregateFunc{
			AnySQLFunction: AnySQLFunction{
				FunctionName: name,
				Min:          min,
				Max:          max,
				distinct:     distinct,
			},
		}
	}
}

var (
	FunctionCOUNTGetter       = NewAggregateFunctionWithGetter("count", 0, 1, false)
	FunctionGROUPCONCATGetter = NewAggregateFunctionWithGetter("group_concat", 1, 2, false)
	// If MAX/MIN has a single argument, it returns the maximum/minimum value
	// of all values in the group. If it has two arguments, it returns the
	// maximum/minimum value of the set of arguments.
	// The first one is an aggregate function, the second one is a scalar function
	FunctionMAXGetter = NewAggregateFunctionWithGetter("max", 1, 0, false)
	FunctionMINGetter = NewAggregateFunctionWithGetter("min", 1, 0, false)
)

var (
	FunctionCOUNT = AggregateFunc{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "count",
			Max:          1,
		},
	}

	FunctionGROUPCONCAT = AggregateFunc{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "group_concat",
			Min:          1,
			Max:          2,
		},
	}

	// If MAX/MIN has a single argument, it returns the maximum/minimum value
	// of all values in the group. If it has two arguments, it returns the
	// maximum/minimum value of the set of arguments.
	// The first one is an aggregate function, the second one is a scalar function
	FunctionMAX = AggregateFunc{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "max",
			Min:          1,
		},
	}

	FunctionMIN = AggregateFunc{
		AnySQLFunction: AnySQLFunction{
			FunctionName: "min",
			Min:          1,
		},
	}
)
