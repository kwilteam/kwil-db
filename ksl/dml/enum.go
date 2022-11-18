package dml

type Enum struct {
	Name          string
	Values        []EnumValue
	Documentation string
	DatabaseName  string
}

type EnumValue struct {
	Name          string
	DatabaseName  string
	Documentation string
}
