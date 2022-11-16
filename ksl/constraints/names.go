package constraints

import "strings"

func PrimaryKeyName(tableName string, maxLength int) string {
	suffix := "_pkey"

	name := tableName
	if len(tableName) >= maxLength-len(suffix) {
		split := floorCharBoundary(tableName, maxLength-len(suffix))
		name = tableName[:split]
	}
	return name + suffix
}

func UniqueIndexName(tableName string, columns []string, maxLength int) string {
	suffix := "_key"

	name := tableName + "_" + strings.Join(columns, "_")
	if len(name) >= maxLength-len(suffix) {
		split := floorCharBoundary(name, maxLength-len(suffix))
		name = name[:split]
	}
	return name + suffix
}

func NonUniqueIndexName(tableName string, columns []string, maxLength int) string {
	suffix := "_idx"

	name := tableName + "_" + strings.Join(columns, "_")
	if len(name) >= maxLength-len(suffix) {
		split := floorCharBoundary(name, maxLength-len(suffix))
		name = name[:split]
	}
	return name + suffix
}

func DefaultName(tableName string, columnName string, maxLength int) string {
	suffix := "_df"

	name := tableName + "_" + columnName
	if len(name) >= maxLength-len(suffix) {
		split := floorCharBoundary(name, maxLength-len(suffix))
		name = name[:split]
	}
	return name + suffix
}

func ForeignKeyConstraintName(tableName string, columns []string, maxLength int) string {
	suffix := "_fkey"

	name := tableName + "_" + strings.Join(columns, "_")
	if len(name) >= maxLength-len(suffix) {
		split := floorCharBoundary(name, maxLength-len(suffix))
		name = name[:split]
	}
	return name + suffix
}

func floorCharBoundary(s string, idx int) int {
	if idx >= len(s) {
		return len(s)
	} else {
		for !isCharBoundary(s, idx) {
			idx--
		}

		return idx
	}
}

func isCharBoundary(s string, idx int) bool {
	return idx == 0 || idx == len(s) || s[idx] < 128 || s[idx] >= 192
}
