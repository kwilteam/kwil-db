package engine

// TODO: This logic is PostgreSQL-specific and needs to be refactored to support MySQL

func isComparisonOperator(s string) bool {
	switch s {
	case ">":
	case "<":
	case "<=":
	case ">=":
	case "=":
	case "<>":
	case "!=":
	default:
		return false
	}
	return true
}

func isMathematicalOperator(s string) bool {
	switch s {
	case "+":
	case "-":
	case "*":
	case "/":
	case "%":
	case "^":
	case "|/":
	case "||/":
	case "!":
	case "!!":
	case "@":
	case "&":
	case "|":
	case "#":
	case "~":
	case "<<":
	case ">>":
	default:
		return false
	}
	return true
}
