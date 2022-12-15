package cache

import (
	"fmt"
	"kwil/x/sqlx/models"
)

type Database struct {
	Name        string
	Owner       string
	DefaultRole string
	Tables      map[string]*Table
	Queries     map[string]*Executable
	Roles       map[string]*Role
	Indexes     map[string]*Index
}

func (c *Database) From(db *models.Database) error {
	c.Name = db.Name
	c.Owner = db.Owner
	c.DefaultRole = db.DefaultRole

	// Tables
	c.Tables = make(map[string]*Table)
	for _, table := range db.Tables {
		tbl := &Table{}
		err := tbl.From(table)
		if err != nil {
			return fmt.Errorf("failed to convert table %s: %s", table.Name, err.Error())
		}
		c.Tables[table.Name] = tbl
	}

	// Set Queries
	c.Queries = make(map[string]*Executable)
	for _, query := range db.SQLQueries {
		exec, err := query.Prepare(db)
		if err != nil {
			return fmt.Errorf("failed to convert query %s: %s", query.Name, err.Error())
		}

		qry := &Executable{}
		err = qry.From(exec)
		if err != nil {
			return fmt.Errorf("failed to convert query %s: %s", query.Name, err.Error())
		}
		c.Queries[query.Name] = qry
	}

	// Set Roles
	c.Roles = make(map[string]*Role)
	for _, role := range db.Roles {
		rol := &Role{}
		err := rol.From(role)
		if err != nil {
			return fmt.Errorf("failed to convert role %s: %s", role.Name, err.Error())
		}
		c.Roles[role.Name] = rol
	}

	// Set Indexes
	c.Indexes = make(map[string]*Index)
	for _, index := range db.Indexes {
		ind := &Index{}
		err := ind.From(index)
		if err != nil {
			return fmt.Errorf("failed to convert index %s: %s", index.Name, err.Error())
		}
		c.Indexes[index.Name] = ind
	}
	return nil
}

func (c *Database) GetTable(name string) (*Table, bool) {
	tbl, ok := c.Tables[name]
	return tbl, ok
}

func (c *Database) GetQuery(name string) (*Executable, bool) {
	qry, ok := c.Queries[name]
	return qry, ok
}

func (c *Database) GetRole(name string) (*Role, bool) {
	rol, ok := c.Roles[name]
	return rol, ok
}

func (c *Database) GetIndex(name string) (*Index, bool) {
	ind, ok := c.Indexes[name]
	return ind, ok
}

func (d *Database) GetSchemaName() string {
	return d.Name + "_" + d.Owner
}
