package specifications

var (
	SchemaLoader DatabaseSchemaLoader = &FileDatabaseSchemaLoader{}
)

func SetSchemaLoader(loader DatabaseSchemaLoader) {
	SchemaLoader = loader
}
