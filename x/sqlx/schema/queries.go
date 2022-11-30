package schema

func (d *Database) ListQueries() []string {
	var queries []string
	for k := range d.Queries {
		queries = append(queries, k)
	}
	return queries
}
