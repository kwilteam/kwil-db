package types

type CreateDatabase struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	DBType    string `json:"type"`
	Fee       string `json:"fee"`
	Signature string `json:"signature"`
	From      string `json:"from"`
}

type DDL struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Owner     string `json:"owner"`
	DBType    string `json:"type"`
	DDLType   string `json:"ddl_type"`
	DDL       string `json:"ddl"`
	Fee       string `json:"fee"`
	Signature string `json:"signature"`
	From      string `json:"from"`
}
