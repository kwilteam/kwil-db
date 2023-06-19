package dataset2

type Cache struct {
	preparedStatements    map[string]PreparedStatement
	initializedExtensions map[string]InitializedExtension
	extensionInitializers map[string]Initializer
	procedures            map[string]*StoredProcedure
}

func newCache() *Cache {
	return &Cache{
		preparedStatements:    make(map[string]PreparedStatement),
		initializedExtensions: make(map[string]InitializedExtension),
		extensionInitializers: make(map[string]Initializer),
		procedures:            make(map[string]*StoredProcedure),
	}
}
