package client

import (
	// "github.com/kwilteam/kwil-db/pkg/sql"
	"bytes"
	"fmt"
	"sort"

	"github.com/kwilteam/kwil-db/pkg/crypto"
	"github.com/kwilteam/kwil-db/pkg/sql"
	"github.com/kwilteam/kwil-db/pkg/sql/sqlite"
)

type SqliteSession struct {
	sess *sqlite.Session
}

func (s *SqliteSession) GenerateChangeset() (sql.Changeset, error) {
	cs, err := s.sess.GenerateChangeset()
	if err != nil {
		return nil, err
	}

	return &Changeset{
		changeset: cs,
	}, nil

}

func (s *SqliteSession) Delete() error {
	return s.sess.Delete()
}

type Changeset struct {
	changeset *sqlite.Changeset
	id        []byte
}

// Export gets the changeset as a byte array.
func (c *Changeset) Export() ([]byte, error) {
	bts := c.changeset.Export()

	err := c.changeset.Reset()
	if err != nil {
		return nil, err
	}

	return bts, nil
}

func (c *Changeset) Close() error {
	return c.changeset.Close()
}

/*
ID generates a determinstic ID for the changeset.
It does this by:
1. ordering each record by primary key
if there are multiple primary keys, it will concatenate them alphabetically
ex: if the primary keys are (id, name), and the identified row has id=1 and name="John",
the primary key will be "1John".  this is case sensitive.  below, it is referred to as the "primary key identifier"

2. extracting an identifier for each "change" (this is made to be compatible with patchsets)
For inserts, it will be the primary key and the new values, ordered by column index, concatenated as bytes.
For updates, it will be the primary key and the new values for changed values, ordered by column index, concatenated as bytes.
For deletes, it will be the primary key identifier and the old values, ordered by column index, concatenated as bytes.

Each concatenated identifier is then hashed using sha224, and the hashes are concatenated as bytes, and then hashed again using sha224.
*/
func (c *Changeset) ID() ([]byte, error) {
	if c.id != nil {
		return c.id, nil
	}

	idents := &identifiers{
		data: make([]*identifiedRecord, 0),
	}

	for {
		rowReturned, err := c.changeset.Next()
		if err != nil {
			return nil, err
		}
		if !rowReturned {
			break
		}

		primaryKeys, err := c.changeset.PrimaryKey()
		if err != nil {
			return nil, err
		}

		recordValues, err := getRecordValues(c.changeset)
		if err != nil {
			return nil, err
		}

		record := &identifiedRecord{
			primaryKey: concatenatePrimaryKeys(primaryKeys),
			values:     RecordValues(recordValues),
		}

		idents.Add(record)
	}

	hash := []byte{}
	for _, id := range idents.data {
		hash = append(hash, id.Hash()...)
	}

	c.id = crypto.Sha224(hash)
	err := c.changeset.Reset()
	if err != nil {
		return nil, err
	}

	return c.id, nil
}

func RecordValues(records [][]byte) [][]byte {
	RecordValues := make([][]byte, len(records))
	for i, value := range records {
		Val := make([]byte, len(value))
		copy(Val, value)
		RecordValues[i] = Val
	}
	return RecordValues
}

/*
getRecordValues gets the values for the current record.
It does this depending on the operation type:
- for inserts, it gets the new values, ordered by column index
- for updates, it gets the new values for changed values, ordered by column index
- for deletes, it does not get any values
*/
func getRecordValues(c *sqlite.Changeset) ([][]byte, error) {
	operation, err := c.Operation()
	if err != nil {
		return nil, err
	}

	switch operation.Type {
	case sqlite.OpInsert:
		bts := make([][]byte, 0)

		for i := 0; i < operation.NumColumns; i++ {
			value, err := c.New(i)
			if err != nil {
				return nil, err
			}

			bts = append(bts, value.Blob())
		}

		return bts, nil
	case sqlite.OpUpdate:
		bts := make([][]byte, 0)
		for i := 0; i < operation.NumColumns; i++ {
			value, err := c.New(i)
			if err != nil {
				return nil, err
			}
			if !value.Changed() {
				continue
			}

			bts = append(bts, value.Blob())
		}

		return bts, nil
	case sqlite.OpDelete:
		return make([][]byte, 0), nil
	default:
		return nil, fmt.Errorf("unknown operation type received from database engine: %v", operation.Type)
	}
}

type identifiedRecord struct {
	primaryKey []byte
	values     [][]byte
}

// Identify returns the primary key identifier
func (i *identifiedRecord) Identify() []byte {
	return i.primaryKey
}

// Hash returns the hash of the record
func (i *identifiedRecord) Hash() []byte {
	return crypto.Sha224(append(i.primaryKey, bytes.Join(i.values, nil)...))
}

// concatenatePrimaryKeys concatenates the primary keys into a byte array
func concatenatePrimaryKeys(primaryKeys []*sqlite.Value) []byte {
	pks := make([][]byte, len(primaryKeys))
	for _, primaryKey := range primaryKeys {
		pks = append(pks, primaryKey.Blob())
	}

	sortBytesLexicographically(pks)

	return bytes.Join(pks, nil)
}

// identifiers is used for logn ordering times for identifiers
type identifiers struct {
	data []*identifiedRecord
}

// Add method to add a new Identifier alphabetically
func (i *identifiers) Add(id *identifiedRecord) {
	index := sort.Search(len(i.data), func(j int) bool { return bytes.Compare(i.data[j].Identify(), id.Identify()) >= 0 })
	i.data = append(i.data, nil)
	copy(i.data[index+1:], i.data[index:])
	i.data[index] = id
}

// sorting:

// byteSlices is a 2D byte slice
type byteSlices [][]byte

// Implementing sort.Interface for ByteSlices
func (b byteSlices) Len() int           { return len(b) }
func (b byteSlices) Less(i, j int) bool { return bytes.Compare(b[i], b[j]) < 0 }
func (b byteSlices) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// sortBytesLexicographically sorts a 2D byte slice lexicographically
func sortBytesLexicographically(bts [][]byte) {
	sort.Sort(byteSlices(bts))
}
