package types

type Extension struct {
	Name           string            `json:"name"`
	Initialization map[string]string `json:"initialization"`
	Alias          string            `json:"alias"`
}
