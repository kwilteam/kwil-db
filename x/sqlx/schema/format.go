package schema

func FormatOwner(owner string) string {
	return "o" + owner[2:]
}

func FormatTable(table string) string {
	return "o" + table[2:]
}

func FormatConstraint(constraint string) string {
	return "c" + constraint[2:]
}
