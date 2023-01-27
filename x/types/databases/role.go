package databases

type Role struct {
	Name        string   `json:"name" clean:"lower"`
	Default     bool     `json:"default"`
	Permissions []string `json:"permissions" clean:"lower"`
}
