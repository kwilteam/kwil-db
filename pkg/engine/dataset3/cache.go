package dataset2

type Cache struct {
	preparedStatements    map[string]PreparedStatement
	initializedExtensions map[string]InitializedExtension
	extensionInitializers map[string]Initializer
	procedures            map[string]*StoredProcedure
}
