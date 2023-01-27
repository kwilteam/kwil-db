package specifications

var (
	SchemaLoader DatabaseSchemaLoader = &FileDatabaseSchemaLoader{FilePath: ""}
)

func SetSchemaLoader(loader DatabaseSchemaLoader) {
	SchemaLoader = loader
}
