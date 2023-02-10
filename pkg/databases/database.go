package databases

import (
	"kwil/pkg/crypto"
	"kwil/pkg/databases/spec"
	"strings"
)

type Database[T spec.AnyValue] struct {
	Owner      string         `json:"owner" clean:"lower"`
	Name       string         `json:"name" clean:"lower"`
	Tables     []*Table[T]    `json:"tables"`
	Roles      []*Role        `json:"roles"`
	SQLQueries []*SQLQuery[T] `json:"sql_queries"`
	Indexes    []*Index       `json:"indexes"`
}

// hashes the lower-cased name and owner and prepends an x
func (d *Database[T]) GetSchemaName() string {
	return GenerateSchemaName(d.Owner, d.Name)
}

func (d *Database[T]) GetQuery(q string) *SQLQuery[T] {
	for _, qry := range d.SQLQueries {
		if qry.Name == q {
			return qry
		}
	}
	return nil
}

func (d *Database[T]) GetTable(t string) *Table[T] {
	for _, tbl := range d.Tables {
		if tbl.Name == t {
			return tbl
		}
	}
	return nil
}

func (d *Database[T]) GetRole(r string) *Role {
	for _, role := range d.Roles {
		if role.Name == r {
			return role
		}
	}
	return nil
}

func (d *Database[T]) GetIndex(i string) *Index {
	for _, idx := range d.Indexes {
		if idx.Name == i {
			return idx
		}
	}
	return nil
}

func (d *Database[T]) GetDefaultRoles() []string {
	var roles []string
	for _, role := range d.Roles {
		if role.Default {
			roles = append(roles, role.Name)
		}
	}
	return roles
}

func (d *Database[T]) GetIdentifier() *DatabaseIdentifier {
	return &DatabaseIdentifier{
		Owner: d.Owner,
		Name:  d.Name,
	}
}

type DatabaseIdentifier struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
}

func (d *DatabaseIdentifier) GetSchemaName() string {
	return GenerateSchemaName(d.Owner, d.Name)
}

func GenerateSchemaName(owner, name string) string {
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
