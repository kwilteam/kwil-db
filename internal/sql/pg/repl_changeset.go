package pg

import (
	"encoding/binary"
	"fmt"
	"io"
	"sync/atomic"

	"github.com/jackc/pglogrepl"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/serialize"
)

type changesetIoWriter struct {
	// writable is an atomic boolean that is true if the changeset is writable.
	writable atomic.Bool

	writer io.Writer

	metadata *changesetMetadata

	oidToType map[uint32]*datatype
}

var (
	changesetInsertByte   = byte(0x01)
	changesetUpdateByte   = byte(0x02)
	changesetDeleteByte   = byte(0x03)
	changesetMetadataByte = byte(0x04)
)

// registerMetadata registers a relation with the changeset metadata.
// it returns the index of the relation in the metadata.
func (c *changesetIoWriter) registerMetadata(relation *pglogrepl.RelationMessageV2) uint32 {
	idx, ok := c.metadata.relationIdx[[2]string{relation.Namespace, relation.RelationName}]
	if ok {
		return uint32(idx)
	}

	c.metadata.relationIdx[[2]string{relation.Namespace, relation.RelationName}] = len(c.metadata.Relations)
	rel := &Relation{
		Schema: relation.Namespace,
		Name:   relation.RelationName,
		Cols:   make([]*Column, len(relation.Columns)),
	}

	for i, col := range relation.Columns {
		dt, ok := c.oidToType[col.DataType]
		if !ok {
			panic(fmt.Sprintf("unknown data type OID %d", col.DataType))
		}

		rel.Cols[i] = &Column{
			Name: col.Name,
			Type: dt.KwilType,
		}
	}

	c.metadata.Relations = append(c.metadata.Relations, rel)

	return uint32(len(c.metadata.Relations) - 1)
}

func (c *changesetIoWriter) decodeInsert(insert *pglogrepl.InsertMessageV2, relation *pglogrepl.RelationMessageV2) error {
	if !c.writable.Load() {
		return nil
	}

	idx := c.registerMetadata(relation)

	tup, err := convertPgxTuple(insert.Tuple, relation, c.oidToType)
	if err != nil {
		return err
	}
	tup.RelationIdx = idx

	bts, err := tup.serialize()
	if err != nil {
		return err
	}

	_, err = c.writer.Write(append([]byte{changesetInsertByte}, bts...))
	return err
}

func (c *changesetIoWriter) decodeUpdate(update *pglogrepl.UpdateMessageV2, relation *pglogrepl.RelationMessageV2) error {
	if !c.writable.Load() {
		return nil
	}

	idx := c.registerMetadata(relation)

	tup, err := convertPgxTuple(update.OldTuple, relation, c.oidToType)
	if err != nil {
		return err
	}
	tup.RelationIdx = idx

	bts, err := tup.serialize()
	if err != nil {
		return err
	}

	tup, err = convertPgxTuple(update.NewTuple, relation, c.oidToType)
	if err != nil {
		return err
	}
	tup.RelationIdx = idx

	bts2, err := tup.serialize()
	if err != nil {
		return err
	}

	_, err = c.writer.Write(append([]byte{changesetUpdateByte}, append(bts, bts2...)...))
	return err
}

func (c *changesetIoWriter) decodeDelete(delete *pglogrepl.DeleteMessageV2, relation *pglogrepl.RelationMessageV2) error {
	if !c.writable.Load() {
		return nil
	}

	idx := c.registerMetadata(relation)

	tup, err := convertPgxTuple(delete.OldTuple, relation, c.oidToType)
	if err != nil {
		return err
	}
	tup.RelationIdx = idx

	bts, err := tup.serialize()
	if err != nil {
		return err
	}

	_, err = c.writer.Write(append([]byte{changesetDeleteByte}, bts...))
	return err
}

// commit is called when the changeset is complete.
// It exports the metadata to the writer.
// It zeroes the metadata, so that the changeset can be reused,
// and send a finish signal to the writer.
func (c *changesetIoWriter) commit() error {
	if !c.writable.Load() {
		return nil
	}

	bts, err := c.metadata.serialize()
	if err != nil {
		return err
	}

	_, err = c.writer.Write(append([]byte{changesetMetadataByte}, bts...))
	if err != nil {
		return err
	}

	c.metadata = &changesetMetadata{
		relationIdx: map[[2]string]int{},
	}

	return nil
}

// fail is called when the changeset is incomplete.
// It zeroes the metadata, so that the changeset can be reused.
func (c *changesetIoWriter) fail() {
	if !c.writable.Load() {
		return
	}

	c.metadata = &changesetMetadata{
		relationIdx: map[[2]string]int{},
	}
}

// ChangesetGroup is a group of changesets.
type ChangesetGroup struct {
	// Changesets is a list of changesets, as they were
	// encountered in the WAL stream.
	// It is meant to be RLP encoded.
	Changesets []*Changeset
}

// convertPgxTuple converts a pgx TupleData to a Tuple.
func convertPgxTuple(pgxTuple *pglogrepl.TupleData, relation *pglogrepl.RelationMessageV2, oidToType map[uint32]*datatype) (*Tuple, error) {
	tuple := &Tuple{
		Columns: make([]*TupleColumn, len(pgxTuple.Columns)),
	}

	for i, col := range pgxTuple.Columns {
		tupleCol := &TupleColumn{}

		dataType, ok := oidToType[relation.Columns[i].DataType]
		if !ok {
			return nil, fmt.Errorf("unknown data type OID %d", relation.Columns[i].DataType)
		}

		switch col.DataType {
		case pglogrepl.TupleDataTypeText:
			tupleCol.ValueType = SerializedValue
			encoded, err := dataType.SerializeChangeset(string(col.Data))
			if err != nil {
				return nil, err
			}

			tupleCol.Data = encoded
		case pglogrepl.TupleDataTypeBinary:
			panic("per pgx docs, we should never actually get this type")
		case pglogrepl.TupleDataTypeNull:
			tupleCol.ValueType = NullValue
		case pglogrepl.TupleDataTypeToast:
			tupleCol.ValueType = ToastValue
		default:
			panic(fmt.Sprintf("unknown tuple data type %d", col.DataType))
		}

		tuple.Columns[i] = tupleCol
	}

	return tuple, nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
// It serializes the ChangesetGroup using RLP. We could probably make
// a custom encoding format that is faster and more compact, but for this
// initial implementation, we will use RLP.
func (c *ChangesetGroup) MarshalBinary() ([]byte, error) {
	return serialize.Encode(c.Changesets)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (c *ChangesetGroup) UnmarshalBinary(data []byte) error {
	err := serialize.Decode(data, &c.Changesets)
	if err != nil {
		return err
	}

	return nil
}

// DeserializeChangeset deserializes a changeset a serialized changeset stream.
func DeserializeChangeset(data []byte) (*ChangesetGroup, error) {
	var inserts []*Tuple
	var updates [][2]*Tuple
	var deletes []*Tuple
	metadata := &changesetMetadata{}
	var err error
	for {
		switch data[0] {
		case changesetInsertByte:
			tup := &Tuple{}
			data, err = tup.deserialize(data[1:])
			if err != nil {
				return nil, err
			}
			inserts = append(inserts, tup)
		case changesetUpdateByte:
			tup1 := &Tuple{}
			data, err = tup1.deserialize(data[1:])
			if err != nil {
				return nil, err
			}

			tup2 := &Tuple{}
			data, err = tup2.deserialize(data)
			if err != nil {
				return nil, err
			}

			updates = append(updates, [2]*Tuple{tup1, tup2})
		case changesetDeleteByte:
			tup := &Tuple{}
			data, err = tup.deserialize(data[1:])
			if err != nil {
				return nil, err
			}
			deletes = append(deletes, tup)
		case changesetMetadataByte:
			data, err = metadata.deserialize(data[1:])
			if err != nil {
				return nil, err
			}

			// this is the end of the changeset
			if len(data) != 0 {
				return nil, fmt.Errorf("unexpected data after metadata: %v", data)
			}
		default:
			return nil, fmt.Errorf("unknown changeset byte %d", data[0])
		}

		if len(data) == 0 {
			break
		}
	}

	group := &ChangesetGroup{
		Changesets: make([]*Changeset, len(metadata.Relations)),
	}

	for i, rel := range metadata.Relations {
		group.Changesets[i] = &Changeset{
			Schema:  rel.Schema,
			Table:   rel.Name,
			Columns: rel.Cols,
		}
	}

	for _, tup := range inserts {
		group.Changesets[tup.RelationIdx].Inserts = append(group.Changesets[tup.RelationIdx].Inserts, tup)
	}

	for _, tup := range updates {
		group.Changesets[tup[0].RelationIdx].Updates = append(group.Changesets[tup[0].RelationIdx].Updates, tup)
	}

	for _, tup := range deletes {
		group.Changesets[tup.RelationIdx].Deletes = append(group.Changesets[tup.RelationIdx].Deletes, tup)
	}

	return group, nil
}

// Changeset is a set of changes to a table.
// It is meant to be RLP encoded, to be compact and easy to send over the wire,
// while also being deterministic. It is meant to translate a lot of the internal
// implementation details into changesets that are understood by higher-level
// Kwil components.
type Changeset struct {
	// Schema is the PostgreSQL schema name.
	Schema string
	// Table is the name of the table.
	Table string
	// Columns is a list of column names and their values.
	Columns []*Column
	// Inserts is a list of tuples to insert.
	Inserts []*Tuple
	// Updates is a list of tuples pairs to update.
	// The first tuple is the old tuple, the second is the new tuple.
	Updates [][2]*Tuple
	// Deletes is a list of tuples to delete.
	// It is the values of each tuple before it was deleted.
	Deletes []*Tuple
}

// DecodeTuple decodes serialized tuple column values into their native types.
// Any value may be nil, depending on the ValueType.
func (c *Changeset) DecodeTuple(tuple *Tuple) ([]any, error) {
	values := make([]any, len(tuple.Columns))
	for i, col := range tuple.Columns {
		switch col.ValueType {
		case NullValue:
			values[i] = nil
		case ToastValue:
			values[i] = nil
		case SerializedValue:
			dt, ok := kwilTypeToDataType[*c.Columns[i].Type]
			if !ok {
				return nil, fmt.Errorf("unknown data type %s", c.Columns[i].Type)
			}
			val, err := dt.DeserializeChangeset(col.Data)
			if err != nil {
				return nil, err
			}

			values[i] = val
		}
	}

	return values, nil
}

// changesetMetadata contains metadata about a changeset.
type changesetMetadata struct {
	// Relation is the schema and table name of the changeset.
	// It is used as a key in the changeset map.
	Relations []*Relation
	// relationIdx is a map of relations, indexed by the schema and table name.
	// it points to the index of the relation in the Relations list.
	relationIdx map[[2]string]int
}

// serialize serializes the metadata with the length of the serialized data
// as a 4-byte prefix.
func (m *changesetMetadata) serialize() ([]byte, error) {
	bts, err := serialize.Encode(m.Relations)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(bts)))

	return append(buf, bts...), nil
}

// deserialize deserializes the metadata.
// It returns the remaining data after the metadata.
func (m *changesetMetadata) deserialize(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short")
	}

	size := binary.LittleEndian.Uint32(data[:4])
	if len(data) < int(size)+4 {
		return nil, fmt.Errorf("data too short")
	}

	err := serialize.Decode(data[4:4+size], &m.Relations)
	if err != nil {
		return nil, err
	}

	m.relationIdx = make(map[[2]string]int)
	for i, rel := range m.Relations {
		m.relationIdx[[2]string{rel.Schema, rel.Name}] = i
	}

	return data[4+size:], nil
}

// Relation is a table in a schema.
type Relation struct {
	Schema string
	Name   string
	Cols   []*Column
}

// Column is a column name and value.
type Column struct {
	Name string
	Type *types.DataType
}

// Tuple is a tuple of values.
type Tuple struct {
	// relationIdx is the index of the relation in the changeset metadata struct.
	RelationIdx uint32
	// Columns is a list of columns and their values.
	Columns []*TupleColumn
}

// serialize serializes the tuple with the length of the serialized data
// as a 4-byte prefix.
func (t *Tuple) serialize() ([]byte, error) {
	bts, err := serialize.Encode(t)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(bts)))

	return append(buf, bts...), nil
}

// deserialize deserializes the tuple.
// It returns the remaining data after the tuple.
func (t *Tuple) deserialize(data []byte) ([]byte, error) {
	if len(data) < 4 {
		return nil, fmt.Errorf("data too short")
	}

	size := binary.LittleEndian.Uint32(data[:4])
	if len(data) < int(size)+4 {
		return nil, fmt.Errorf("data too short")
	}

	err := serialize.Decode(data[4:4+size], &t)
	if err != nil {
		return nil, err
	}

	return data[4+size:], nil
}

// TupleColumn is a column within a tuple.
type TupleColumn struct {
	// ValueType gives information on the type of data in the column.
	// If the type is of type Null or Toast, the Data field will be nil.
	ValueType ValueType
	// Data is the actual data in the column.
	Data []byte
}

// ValueType gives information on the type of data in a tuple column.
type ValueType uint8

const (
	// NullValue indicates a NULL value
	// (as opposed to something like an empty string).
	NullValue ValueType = iota
	// ToastValue indicates a column is a TOAST pointer,
	// and that the actual value is stored elsewhere and
	// was unchanged.
	ToastValue
	// SerializedValue indicates a column is a non-nil value
	// and can be deserialized.
	SerializedValue
)
