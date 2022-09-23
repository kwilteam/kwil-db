package db

import (
	"github.com/osamingo/boolconv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"kwil/pkg/types/db"
)

type DB struct {
	log    *zerolog.Logger
	prefix []byte
	conf   db.DatabaseConfig
	loader *DBLoader
}

// This could probably use a rename
// The point of this is so that I can use either
// a transaction or a regular kv store
// in the store role and paramaterized query functions
type KVBasic interface {
	Set(k, v []byte) error
	Get(k []byte) ([]byte, error)
}

// Make sure to pass this a pointer!
func NewDB(dbConf db.DatabaseConfig, l *DBLoader) *DB {
	pref := getDBPrefix(dbConf)
	logger := log.With().Str("module", "dba").Str("db_name", string(pref)).Logger()

	return &DB{log: &logger, prefix: pref, conf: dbConf, loader: l}
}

// StoreAll stores the name, owner, dbType, and defaultRole of the database.
// This does not check if values already exist
func (d *DB) StoreAll(isTx bool) error {
	var k KVBasic
	var com func() error
	if isTx {
		// Make a tx
		tx := txn{
			btx: d.loader.kv.NewTransaction(true),
		}
		defer tx.Discard()
		com = tx.Commit

		// Store the name
		err := tx.Set(append(d.prefix, []byte("name")...), []byte(*d.conf.GetName()))
		if err != nil {
			return err
		}

		// Store the owner
		err = tx.Set(append(d.prefix, []byte("owner")...), []byte(*d.conf.GetOwner()))
		if err != nil {
			return err
		}

		// Store the dbType
		err = tx.Set(append(d.prefix, []byte("dbType")...), []byte(*d.conf.GetDBType()))
		if err != nil {
			return err
		}

		// Store the defaultRole
		err = tx.Set(append(d.prefix, []byte("defRole")...), []byte(*d.conf.GetDefaultRole()))
		if err != nil {
			return err
		}

		// Set k
		k = tx
	} else {

		// Store the name
		err := d.Set([]byte("name"), []byte(*d.conf.GetName()))
		if err != nil {
			return err
		}
		// Store the owner
		err = d.Set([]byte("owner"), []byte(*d.conf.GetOwner()))
		if err != nil {
			return err
		}
		// Store the dbType
		err = d.Set([]byte("dbType"), []byte(*d.conf.GetDBType()))
		if err != nil {
			return err
		}
		// Store the defaultRole
		err = d.Set([]byte("defRole"), []byte(*d.conf.GetDefaultRole()))
		if err != nil {
			return err
		}

		// Set k
		k = d.loader.kv
	}

	// We also have to store the roles and paramaterized_queries
	//rolePref := append(d.prefix, []byte("roles")...)
	structure := d.conf.GetStructure()
	for _, role := range *structure.GetRoles() {
		d.log.Info().Msgf("storing role %s", role.Name)
		err := StoreRole(&role, k)
		if err != nil {
			return err
		}
	}

	// Store the paramaterized queries
	for _, pq := range *structure.GetQueries() {
		d.log.Info().Msgf("storing paramaterized query %s", pq.Name)
		err := StoreParQuer(&pq, k)
		if err != nil {
			return err
		}
	}

	if isTx { // if it a tx, commit.  Otherwise just return nil
		return com()
	}
	return nil
}

// Not yet used
/*
func (d *DB) StoreAllIfNotExists(isTx bool) error {
	exists, err := d.loader.kv.Exists(d.prefix)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return d.StoreAll(isTx)
}*/

func (d *DB) DeleteEntireDB() error {
	return d.loader.kv.DeleteByPrefix(d.prefix)
}

// Set function is a wrapper that adds the DB prefix
func (d *DB) Set(k, v []byte) error {
	_ = d.loader.kv.Set(append(d.prefix, k...), v)
	return nil
}

func (d *DB) Get(k []byte) ([]byte, error) {
	return d.loader.kv.Get(append(d.prefix, k...))
}

func (d *DB) Delete(k []byte) error {
	return d.loader.kv.Delete(append(d.prefix, k...))
}

func (d *DB) Close() error {
	return d.loader.kv.Close()
}

func (d *DB) GetAllByPrefix(prefix []byte) ([][]byte, [][]byte, error) {
	return d.loader.kv.GetAllByPrefix(append(d.prefix, prefix...))
}

func StoreRole(role *db.Role, d KVBasic) error {
	perms := role.GetPermissions()

	ddlKey := getRoleDDLKey(role)
	err := d.Set(ddlKey, bool2ByteArr(perms.DDL))
	if err != nil {
		return err
	}

	// Now we loop through thr ParamaterizedQueries and store them
	for _, pq := range perms.ParamaterizedQueries {
		pqKey := getRolePQKey(role, pq)
		err := d.Set(pqKey, bool2ByteArr(true))
		if err != nil {
			return err
		}
	}

	return nil
}

func getRoleDDLKey(role *db.Role) []byte {
	return append([]byte(role.GetName()), []byte("ddl")...)
}

func getRolePQKey(role *db.Role, pq string) []byte {
	k := append([]byte(role.GetName()), []byte("pq")...)
	return append(k, []byte(pq)...)
}

func (d *DB) GetRole(roleName string) (*db.Role, error) {
	var role db.Role
	role.Name = roleName

	// Get the ddl permission
	ddlKey := getRoleDDLKey(&role)
	ddl, err := d.Get(ddlKey)
	if err != nil {
		return nil, err
	}

	// Get the parameterized queries
	// We can't use getRolePQKey here because we need to get all the pq keys for this role
	pqPref := append([]byte(roleName), []byte("pq")...)
	pqKeys, _, err := d.GetAllByPrefix(pqPref)
	if err != nil {
		return nil, err
	}

	// Now loop through the keys and convert them to a string slice
	var pqSlice []string
	// These pqKey values are in the format /<rolename>pq<pq>, so we need to strip /<rolename>pq
	stripLen := len([]byte(roleName+"pq")) + 1 // adding 1 for the '/' at the beginning
	for _, pqKey := range pqKeys {
		// strip the first stripLen bytes and append to the slice
		pqSlice = append(pqSlice, string(pqKey[stripLen:]))
	}

	role.Permissions = db.Permissions{
		DDL:                  bytes2Bool(ddl),
		ParamaterizedQueries: pqSlice,
	}

	return &role, nil
}

func bool2ByteArr(b bool) []byte {
	nb := boolconv.NewBool(b)
	return nb.Bytes()
}

func bytes2Bool(b []byte) bool {
	nb := boolconv.BtoB(b)
	return nb.Tob()
}

/* Roles should be stored as followed:
<role_name(string)> ddl : <can_ddl(bool)>
<role_name(string)> queries <query_name(string)> : <can_execute(bool)>

the above should be concataneted as bytes into a single key
*/
