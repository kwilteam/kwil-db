package util

// ExtractSQLName remove surrounding lexical token(pair) of an identifier(name).
// Those tokens are: `"` and `[` `]` and "`".
// In sqlparser identifiers are used for: table name, table alias name, column name,
// column alias name, collation name, index name, function name.
func ExtractSQLName(name string) string {
	// remove surrounding token pairs
	if len(name) > 1 {
		if name[0] == '"' && name[len(name)-1] == '"' {
			name = name[1 : len(name)-1]
		}

		if name[0] == '[' && name[len(name)-1] == ']' {
			name = name[1 : len(name)-1]
		}

		if name[0] == '`' && name[len(name)-1] == '`' {
			name = name[1 : len(name)-1]
		}
	}

	return name
}
