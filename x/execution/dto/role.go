package dto

type Role struct {
	Name        string   `json:"name"`
	Default     bool     `json:"default"`
	Permissions []string `json:"permissions"`
}
