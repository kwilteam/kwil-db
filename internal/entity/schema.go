package entity

type DatasetIdentifier struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

type Schema struct {
	Owner   string    `json:"owner"`
	Name    string    `json:"name"`
	Tables  []*Table  `json:"tables"`
	Actions []*Action `json:"actions"`
}

type Table struct {
	Name    string    `json:"name"`
	Columns []*Column `json:"columns"`
	Indexes []*Index  `json:"indexes"`
}

type Column struct {
	Name       string       `json:"name"`
	Type       string       `json:"type"`
	Attributes []*Attribute `json:"attributes,omitempty"`
}

type Attribute struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type Action struct {
	Name       string   `json:"name"`
	Inputs     []string `json:"inputs"`
	Public     bool     `json:"public"`
	Statements []string `json:"statements"`
}

type Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Type    string   `json:"type"`
}
