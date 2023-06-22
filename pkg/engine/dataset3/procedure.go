package dataset3

type Procedure struct {
	Name       string   `json:"name"`
	Args       []string `json:"args"`
	Public     bool     `json:"public"`
	Statements []string `json:"statements"`
}
