package sqlite

import (
	"fmt"
	"os"
)

const (
	testdbName = "test_temp_db_DELETE_ME"
)

// OpenDbWithTearDown opens a connection to the database with the given name.
// It will create a real database, and return a function that can be used to
// delete the database.
// It will delete any database previously created with this function.
func OpenDbWithTearDown() (*Connection, func() error, error) {
	path := "./tmp/"
	conn, err := OpenConn(testdbName, WithPath(path))
	if err != nil {
		return nil, nil, err
	}

	err = conn.Delete()
	if err != nil {
		fmt.Printf("failed to delete temp database while opening %v\n", err)
	}

	conn2, err := OpenConn(testdbName, WithPath("./tmp/"))
	if err != nil {
		return nil, nil, err
	}

	closeFunc := func() error {

		err = conn2.Delete()
		if err != nil {
			return err
		}

		// delete the temp directory
		return os.RemoveAll(path)
	}

	return conn2, closeFunc, nil
}
