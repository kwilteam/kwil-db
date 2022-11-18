package nodes

func IsNumericValue(e Expression) bool {
	_, ok := e.(*Number)
	return ok
}

func IsStringValue(e Expression) bool {
	_, ok := e.(*String)
	return ok
}

func IsHeredocValue(e Expression) bool {
	_, ok := e.(*Heredoc)
	return ok
}

func IsLiteralValue(e Expression) bool {
	_, ok := e.(*Literal)
	return ok
}

func IsObjectValue(e Expression) bool {
	_, ok := e.(*Object)
	return ok
}

func IsFunctionValue(e Expression) bool {
	_, ok := e.(*Function)
	return ok
}

func IsListValue(e Expression) bool {
	_, ok := e.(*List)
	return ok
}

func IsVariableValue(e Expression) bool {
	_, ok := e.(*Variable)
	return ok
}
