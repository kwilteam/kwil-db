package schema

// Database is defined in models but I am keeping methods here
func NewEmptyDatabase() *Database {
	return &Database{
		Tables:  make(Tables),
		Indexes: make(Indices),
		Queries: DefinedQueries{
			Inserts: make(map[string]*InsertDef),
			Updates: make(map[string]*UpdateDef),
			Deletes: make(map[string]*DeleteDef),
		},
		Roles: make(Roles),
	}
}
