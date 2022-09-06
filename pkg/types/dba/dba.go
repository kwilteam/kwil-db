package dba

type DatabaseConfig interface {
	GetName() *string
	GetOwner() *string
	GetDBType() *string
	GetDefaultRole() *string
	GetStructure() Structure
}

type Structure interface {
	GetRoles() *[]Role
	GetQueries() *[]ParameterizedQuery
}
