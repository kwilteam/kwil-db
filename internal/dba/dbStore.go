package dba

import (
	"github.com/kwilteam/kwil-db/pkg/types/dba"
)

type DB struct {
	prefix  []byte
	name    *string
	owner   *string
	dbType  *string
	defRole *string
	loader  *DBLoader
}

func NewDB(dbConf dba.DatabaseConfig, l *DBLoader) *DB {
	pref := getDBPrefix(dbConf)

	return &DB{prefix: pref, name: dbConf.GetName(), owner: dbConf.GetOwner(), dbType: dbConf.GetDBType(), defRole: dbConf.GetDefaultRole(), loader: l}
}

// StoreAll stores the name, owner, dbType, and defaultRole of the database.
// This does not check if values already exist
func (d *DB) StoreAll() error {
	// Store the name
	err := d.Set([]byte("name"), []byte(*d.name))
	if err != nil {
		return err
	}
	// Store the owner
	err = d.Set([]byte("owner"), []byte(*d.owner))
	if err != nil {
		return err
	}
	// Store the dbType
	err = d.Set([]byte("dbType"), []byte(*d.dbType))
	if err != nil {
		return err
	}
	// Store the defaultRole
	err = d.Set([]byte("defRole"), []byte(*d.defRole))
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) StoreAllIfNotExists() error {
	exists, err := d.loader.kv.Exists(d.prefix)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return d.StoreAll()
}

// Set function is a wrapper that adds the prefix
func (d *DB) Set(k, v []byte) error {
	d.loader.kv.Set(append(d.prefix, k...), v)
	return nil
}
