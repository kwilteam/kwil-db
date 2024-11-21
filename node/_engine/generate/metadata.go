package generate

var (
	// PgSessionPrefix is the prefix for all session variables.
	// It is used in combination with Postgre's current_setting function
	// to set contextual variables.
	PgSessionPrefix = "ctx"
)
