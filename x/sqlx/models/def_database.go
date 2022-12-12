package models

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	types "kwil/x/sqlx/spec"
	"strings"

	"gopkg.in/yaml.v2"
)

type Database struct {
	Owner       string      `json:"owner"`
	Name        string      `json:"name"`
	DefaultRole string      `json:"default_role"`
	Tables      []*Table    `json:"tables"`
	Roles       []*Role     `json:"roles"`
	SQLQueries  []*SQLQuery `json:"sql_queries"`
	Indexes     []*Index    `json:"indexes"`
}

func (d *Database) ToYAML() ([]byte, error) {
	return yaml.Marshal(d)
}

func (d *Database) FromYAML(bts []byte) error {
	db := &Database{}
	err := yaml.Unmarshal(bts, db)
	if err != nil {
		return err
	}
	*d = *db
	return nil
}

func (d *Database) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

func (d *Database) FromJSON(bts []byte) error {
	db := &Database{}
	err := json.Unmarshal(bts, db)
	if err != nil {
		return err
	}
	*d = *db
	return nil
}

func (d *Database) EncodeGOB() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := gob.NewEncoder(buf).Encode(d)
	return buf.Bytes(), err
}

func (d *Database) DecodeGOB(b []byte) error {
	var db Database
	buf := bytes.NewBuffer(b)
	err := gob.NewDecoder(buf).Decode(&db)
	if err != nil {
		return err
	}
	*d = db
	return nil
}

// Validation
func (db *Database) Validate() error {
	// check general database properties
	err := CheckName(db.Name, types.DATABASE)
	if err != nil {
		return err
	}

	err = CheckAddress(db.Owner)
	if err != nil {
		return err
	}

	err = CheckName(db.DefaultRole, types.ROLE)
	if err != nil {
		return err
	}

	// check if default role exists
	if db.GetRole(db.DefaultRole) == nil {
		return fmt.Errorf(`default role "%s" does not exist`, db.DefaultRole)
	}

	// check tables
	tblMap := make(map[string]struct{})
	for _, tbl := range db.Tables {
		// check if table name is unique
		if _, ok := tblMap[tbl.Name]; ok {
			return fmt.Errorf(`duplicate table name "%s"`, tbl.Name)
		}
		tblMap[tbl.Name] = struct{}{}

		err := tbl.Validate(db)
		if err != nil {
			return fmt.Errorf(`error on table "%s": %w`, tbl.Name, err)
		}
	}

	//check queries
	qryMap := make(map[string]struct{})
	for _, qry := range db.SQLQueries {
		// check if query name is unique
		if _, ok := qryMap[qry.Name]; ok {
			return fmt.Errorf(`duplicate query name "%s"`, qry.Name)
		}
		qryMap[qry.Name] = struct{}{}

		err := qry.Validate(db)
		if err != nil {
			return fmt.Errorf(`error on query "%s": %w`, qry.Name, err)
		}
	}

	// check roles
	roleMap := make(map[string]struct{})
	for _, role := range db.Roles {
		// check if role name is unique
		if _, ok := roleMap[role.Name]; ok {
			return fmt.Errorf(`duplicate role name "%s"`, role.Name)
		}
		roleMap[role.Name] = struct{}{}

		err := role.Validate(db)
		if err != nil {
			return fmt.Errorf(`error on role "%s": %w`, role.Name, err)
		}
	}

	// check indexes
	idxMap := make(map[string]struct{})
	for _, idx := range db.Indexes {
		// check if index name is unique
		if _, ok := idxMap[idx.Name]; ok {
			return fmt.Errorf(`duplicate index name "%s"`, idx.Name)
		}
		idxMap[idx.Name] = struct{}{}

		err := idx.Validate(db)
		if err != nil {
			return fmt.Errorf(`error on index "%s": %w`, idx.Name, err)
		}
	}

	return nil
}

func (db *Database) GetTable(name string) *Table {
	for _, t := range db.Tables {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func (db *Database) GetQuery(name string) *SQLQuery {
	for _, q := range db.SQLQueries {
		if q.Name == name {
			return q
		}
	}
	return nil
}

func (db *Database) GetRole(name string) *Role {
	for _, r := range db.Roles {
		if r.Name == name {
			return r
		}
	}
	return nil
}

func (db *Database) GetIndex(name string) *Index {
	for _, i := range db.Indexes {
		if i.Name == name {
			return i
		}
	}
	return nil
}

func (db *Database) GenerateDDL() (string, error) {
	schema_name := db.Name + "_" + db.Owner

	stmts := []string{}
	for _, t := range db.Tables {
		stmt, err := t.GenerateDDL(schema_name)
		if err != nil {
			return "", err
		}
		stmts = append(stmts, stmt...)
	}

	for _, ind := range db.Indexes {
		stmt := ind.GenerateDDL(schema_name)
		stmts = append(stmts, stmt)
	}

	return strings.Join(stmts, "\n "), nil
}

func (db *Database) GetSchemaName() string {
	return db.Name + "_" + db.Owner
}

func (db *Database) PrepareQueries() ([]*ExecutableQuery, error) {
	var queries []*ExecutableQuery
	schemaName := db.GetSchemaName()
	for _, q := range db.SQLQueries {
		eq, err := q.Prepare(schemaName)
		if err != nil {
			return nil, err
		}
		if eq == nil {
			continue
		}
		queries = append(queries, eq)
	}
	return queries, nil
}
