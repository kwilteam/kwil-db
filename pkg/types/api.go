package types

type CreateDatabase struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	DBType    string `json:"type"`
	Fee       string `json:"fee"`
	Signature string `json:"signature"`
	From      string `json:"from"`
}
