package catalog

// Catalog describes a database instance consisting of metadata in which database objects are defined
type Catalog struct {
	Comment       string
	DefaultSchema string
	Name          string
	Schemas       []*Schema
	SearchPath    []string
	LoadExtension func(string) *Schema

	extensions map[string]struct{}
}

// New creates a new catalog
func New(defaultSchema string) *Catalog {
	newCatalog := &Catalog{
		DefaultSchema: defaultSchema,
		Schemas:       make([]*Schema, 0),
		extensions:    make(map[string]struct{}),
	}

	if newCatalog.DefaultSchema != "" {
		newCatalog.Schemas = append(newCatalog.Schemas, &Schema{Name: defaultSchema})
	}

	return newCatalog
}
