package schema

func (d *Database) ListQueries() []string {
	return d.Queries.ListAll()
}
