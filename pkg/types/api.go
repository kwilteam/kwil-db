package types

type CreateDatabase struct {
	Name      string `json:"name"`
	DBType    string `json:"type"`
	Signature string `json:"signature"`
	From      string `json:"from"`
}
