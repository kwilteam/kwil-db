package sqlite

import (
	"os"
)

// OpenDbWithTearDown opens a connection to the database with the given name.
// It will create a real database, and return a function that can be used to
// delete the database.
// It will delete any database previously created with this function.
func OpenDbWithTearDown(name string) (*Connection, func() error, error) {
	path := "./tmp/"
	conn, err := OpenConn(name, WithPath(path))
	if err != nil {
		return nil, nil, err
	}

	closeFunc := func() error {

		err = conn.Delete()
		if err != nil {
			return err
		}

		// delete the temp directory
		return os.RemoveAll(path)
	}

	return conn, closeFunc, nil
}
