package dto

type Result interface {
	Records() []map[string]any
}
