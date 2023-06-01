package tree

import sqlwriter "github.com/kwilteam/kwil-db/pkg/engine/tree/sql-writer"

type distinctable interface {
	SQLFunction
	StringDistinct(exprs ...Expression) string
}

type AggregateFunc struct {
	AnySQLFunction
}

// StringDistinct returns the string representation of the function with the
// given arguments, prepended by the DISTINCT keyword.
func (s *AggregateFunc) StringDistinct(exprs ...Expression) string {
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

var (
	FunctionCOUNT = AggregateFunc{AnySQLFunction: AnySQLFunction{
		FunctionName: "count",
		Max:          1,
	},
	}

	FunctionGROUPCONCAT = AggregateFunc{AnySQLFunction: AnySQLFunction{
		FunctionName: "group_concat",
		Min:          1,
		Max:          2,
	},
	}

	// If MAX/MIN has a single argument, it returns the maximum/minimum value
	// of all values in the group. If it has two arguments, it returns the
	// maximum/minimum value of the set of arguments.
	// The first one is an aggregate function, the second one is a scalar function
	FunctionMAX = AggregateFunc{AnySQLFunction: AnySQLFunction{
		FunctionName: "max",
		Min:          1,
	},
	}

	FunctionMIN = AggregateFunc{AnySQLFunction: AnySQLFunction{
		FunctionName: "min",
		Min:          1,
	},
	}
)
