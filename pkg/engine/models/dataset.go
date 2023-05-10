package models

import (
	"github.com/kwilteam/kwil-db/pkg/crypto"
	"strings"
)

type Dataset struct {
	Owner   string    `json:"owner" clean:"lower"`
	Name    string    `json:"name" clean:"lower"`
	Tables  []*Table  `json:"tables"`
	Actions []*Action `json:"actions"`
}

// hashes the lower-cased name and owner and prepends an x
func (d *Dataset) ID() string {
	return GenerateSchemaId(d.Owner, d.Name)
}

func (d *Dataset) GetTable(t string) *Table {
	for _, tbl := range d.Tables {
		if tbl.Name == t {
			return tbl
		}
	}
	return nil
}

// GetTableMapping returns a map of table names to table objects
func (d *Dataset) GetTableMapping() map[string]*Table {
	mapping := make(map[string]*Table)
	for _, tbl := range d.Tables {
		mapping[tbl.Name] = tbl
	}
	return mapping
}

func (d *Dataset) GetIdentifier() *DatasetIdentifier {
	return &DatasetIdentifier{
		Owner: d.Owner,
		Name:  d.Name,
	}
}

func (d *Dataset) GetAction(a string) *Action {
	for _, act := range d.Actions {
		if act.Name == a {
			return act
		}
	}
	return nil
}

type DatasetIdentifier struct {
	Owner string `json:"owner" clean:"lower"`
	Name  string `json:"name" clean:"lower"`
}

func (d *DatasetIdentifier) ID() string {
	return GenerateSchemaId(d.Owner, d.Name)
}

func GenerateSchemaId(owner, name string) string {
	return "x" + crypto.Sha224Hex(joinBytes([]byte(strings.ToLower(name)), []byte(strings.ToLower(owner))))
}

// joinBytes is a helper function to join multiple byte slices into one
func joinBytes(s ...[]byte) []byte {
	n := 0
	for _, v := range s {
		n += len(v)
	}

	b, i := make([]byte, n), 0
	for _, v := range s {
		i += copy(b[i:], v)
	}
	return b
}
