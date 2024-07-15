package pg

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/jackc/pglogrepl"
	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/serialize"
)

type changesetIoWriter struct {
	metadata  *changesetMetadata
	oidToType map[uint32]*datatype

	writer io.Writer
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
	if c == nil || c.writer == nil { // !c.writable.Load()
		return nil
	}

	idx := c.registerMetadata(relation)

	tup, err := convertPgxTuple(insert.Tuple, relation, c.oidToType)
	if err != nil {
		return err
	}
	tup.RelationIdx = idx

	// write changesetInsertByte
	_, err = c.writer.Write([]byte{changesetInsertByte})
	if err != nil {
		return err
	}

	return tup.serialize(c.writer)
}

func (c *changesetIoWriter) decodeUpdate(update *pglogrepl.UpdateMessageV2, relation *pglogrepl.RelationMessageV2) error {
	if c == nil || c.writer == nil {
		return nil
	}

	idx := c.registerMetadata(relation)

	// write changesetUpdateByte
	_, err := c.writer.Write([]byte{changesetUpdateByte})
	if err != nil {
		return err
	}

	// write old tuple
	tup, err := convertPgxTuple(update.OldTuple, relation, c.oidToType)
	if err != nil {
		return err
	}
	tup.RelationIdx = idx

	err = tup.serialize(c.writer)
	if err != nil {
		return err
	}

	// write new tuple
	tup, err = convertPgxTuple(update.NewTuple, relation, c.oidToType)
	if err != nil {
		return err
	}
	tup.RelationIdx = idx

	return tup.serialize(c.writer)
}

func (c *changesetIoWriter) decodeDelete(delete *pglogrepl.DeleteMessageV2, relation *pglogrepl.RelationMessageV2) error {
	if c == nil || c.writer == nil {
		return nil
	}

	idx := c.registerMetadata(relation)

	// write changesetDeleteByte
	_, err := c.writer.Write([]byte{changesetDeleteByte})
	if err != nil {
		return err
	}

	// write old tuple
	tup, err := convertPgxTuple(delete.OldTuple, relation, c.oidToType)
	if err != nil {
		return err
	}
	tup.RelationIdx = idx

	return tup.serialize(c.writer)
}

// commit is called when the changeset is complete.
// It exports the metadata to the writer.
// It zeroes the metadata, so that the changeset can be reused,
// and send a finish signal to the writer.
func (c *changesetIoWriter) commit() error {
	if c == nil || c.writer == nil {
		return nil
	}

	// write changesetMetadataByte
	_, err := c.writer.Write([]byte{changesetMetadataByte})
	if err != nil {
		return err
	}

	// serialize metadata
	err = c.metadata.serialize(c.writer)
	if err != nil {
		return err
	}

	c.metadata = &changesetMetadata{
		relationIdx: map[[2]string]int{},
	}
	c.writer = nil

	return err
}

// fail is called when the changeset is incomplete.
// It zeroes the metadata and writer, so that another changeset may be collected.
func (c *changesetIoWriter) fail() {
	// if !c.writable.Load() {
	// 	return
	// }

	c.metadata = &changesetMetadata{
		relationIdx: map[[2]string]int{},
	}
	c.writer = nil
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

// applyInserts applies all inserts in the changeset to the database.
// any conflicts during migration will be ignored in favor of whatever already exists on the new network.
func (c *Changeset) applyInserts(ctx context.Context, tx sql.DB) error {
	// If no inserts, return
	if len(c.Inserts) == 0 {
		return nil
	}

	var columnStr, placeholderStr string
	if len(c.Columns) > 0 {
		columnStr = c.Columns[0].Name
		placeholderStr = "$1"
		for i := 1; i < len(c.Columns); i++ {
			columnStr += ", " + c.Columns[i].Name
			placeholderStr += ", $" + fmt.Sprint(i+1)
		}
	}

	// Conflict resolution: DO NOTHING
	// Any conflicts will be ignored, in favor of whatever already exists on the new network.
	insertSql := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s) ON CONFLICT DO NOTHING", c.Schema, c.Table, columnStr, placeholderStr)

	for _, insertOp := range c.Inserts {
		values, err := c.DecodeTuple(insertOp)
		if err != nil {
			return err
		}

		_, err = tx.Execute(ctx, insertSql, values...)
		if err != nil {
			return err
		}
		// Insert a row
	}

	return nil
}

// applyUpdates applies all updates in the changeset to the database.
// Apply updates only if the oldValues in the old network are same as the current record in the new network.
// If not, discard the update in favor of whatever data exists on the new network
func (c *Changeset) applyUpdates(ctx context.Context, tx sql.DB) error {
	// If no updates, return
	if len(c.Updates) == 0 {
		return nil
	}

	updateSql := fmt.Sprintf("UPDATE %s.%s SET ", c.Schema, c.Table)
	for i, col := range c.Columns {
		if i > 0 {
			updateSql += ", "
		}
		updateSql += col.Name + " = $" + fmt.Sprint(i+1)
	}

	// Conflict resolution:
	// If new network's current record is same as the oldValues in the old network, then update the record
	// Else, discard the update in favor of whatever data exists on the new network
	for _, updateOp := range c.Updates {
		newValues, err := c.DecodeTuple(updateOp[1])
		if err != nil {
			return err
		}

		oldValues, err := c.DecodeTuple(updateOp[0])
		if err != nil {
			return err
		}

		var oldArgs []any
		cnt := 1
		whereClause := ""
		for i, v := range oldValues {
			if i > 0 {
				whereClause += " AND "
			}
			if v == nil {
				whereClause += fmt.Sprintf("%s IS NULL", c.Columns[i].Name)
			} else {
				whereClause += fmt.Sprintf("%s = $%d", c.Columns[i].Name, cnt+len(newValues))
				oldArgs = append(oldArgs, v)
				cnt++
			}
		}

		_, err = tx.Execute(ctx, updateSql+" WHERE "+whereClause, append(newValues, oldArgs...)...)
		if err != nil {
			return err
		}
	}

	return nil
}

// applyDeletes applies all deletes in the changeset to the database.
// If the record in the new network is same as the oldValues in the old network, then delete the record
// Else, discard the delete in favor of whatever data exists on the new network
func (c *Changeset) applyDeletes(ctx context.Context, tx sql.DB) error {
	// If no deletes, return
	if len(c.Deletes) == 0 {
		return nil
	}

	deleteSql := fmt.Sprintf("DELETE FROM %s.%s WHERE ", c.Schema, c.Table)

	for _, deleteOp := range c.Deletes {
		record, err := c.DecodeTuple(deleteOp)
		if err != nil {
			return err
		}

		whereClause := ""
		var args []any
		cnt := 1
		for i, v := range record {
			if i > 0 {
				whereClause += " AND "
			}
			if v == nil {
				whereClause += fmt.Sprintf("%s IS NULL", c.Columns[i].Name)
			} else {
				whereClause += fmt.Sprintf("%s = $%d", c.Columns[i].Name, cnt)
				args = append(args, v)
				cnt++
			}
		}

		_, err = tx.Execute(ctx, deleteSql+whereClause, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Changeset) ApplyChangeset(ctx context.Context, tx sql.DB) error {
	// Apply Inserts
	err := c.applyInserts(ctx, tx)
	if err != nil {
		return err
	}

	// Apply Updates
	err = c.applyUpdates(ctx, tx)
	if err != nil {
		return err
	}

	// 	Apply Deletes
	err = c.applyDeletes(ctx, tx)
	if err != nil {
		return err
	}

	return nil
}

// ApplyChangesets applies all changesets in the group to the database.
func (c *ChangesetGroup) ApplyChangesets(ctx context.Context, tx sql.DB) error {
	for _, changeset := range c.Changesets {
		err := changeset.ApplyChangeset(ctx, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

// DeserializeChangeset deserializes a changeset stream.
func DeserializeChangeset(data io.Reader) (*ChangesetGroup, error) {
	var inserts []*Tuple
	var updates [][2]*Tuple
	var deletes []*Tuple
	metadata := &changesetMetadata{}
	var err error
	buf := make([]byte, 1)

	for {
		_, err = data.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		switch buf[0] {
		case changesetInsertByte:
			tup := &Tuple{}
			err = tup.deserialize(data)
			if err != nil {
				return nil, err
			}
			inserts = append(inserts, tup)
		case changesetUpdateByte:
			tup1 := &Tuple{}
			err = tup1.deserialize(data)
			if err != nil {
				return nil, err
			}

			tup2 := &Tuple{}
			err = tup2.deserialize(data)
			if err != nil {
				return nil, err
			}

			updates = append(updates, [2]*Tuple{tup1, tup2})
		case changesetDeleteByte:
			tup := &Tuple{}
			err = tup.deserialize(data)
			if err != nil {
				return nil, err
			}
			deletes = append(deletes, tup)
		case changesetMetadataByte:
			err = metadata.deserialize(data)
			if err != nil {
				return nil, err
			}

			// this is the end of the changeset, check that there is no more data
			_, err = data.Read(buf)
			if err != io.EOF {
				return nil, fmt.Errorf("expected end of changeset, got %d", buf[0])
			}
		default:
			return nil, fmt.Errorf("unknown changeset byte %d", buf[0])
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
func (m *changesetMetadata) serialize(w io.Writer) error {
	bts, err := serialize.Encode(m.Relations)
	if err != nil {
		return err
	}

	size := uint32(len(bts))
	err = binary.Write(w, binary.LittleEndian, size)
	if err != nil {
		return fmt.Errorf("failed to write size: %w", err)
	}

	_, err = w.Write(bts)
	if err != nil {
		return fmt.Errorf("failed to write tuple data: %w", err)
	}

	return nil
}

// deserialize deserializes the metadata.
// It returns the remaining data after the metadata.
func (m *changesetMetadata) deserialize(data io.Reader) error {
	var size uint32
	err := binary.Read(data, binary.LittleEndian, &size)
	if err != nil {
		return fmt.Errorf("failed to read size: %w", err)
	}

	buf := make([]byte, size)
	_, err = io.ReadFull(data, buf)
	if err != nil {
		return fmt.Errorf("failed to read tuple data: %w", err)
	}

	err = serialize.Decode(buf, &m.Relations)
	if err != nil {
		return err
	}

	m.relationIdx = make(map[[2]string]int)
	for i, rel := range m.Relations {
		m.relationIdx[[2]string{rel.Schema, rel.Name}] = i
	}

	return nil
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
func (t *Tuple) serialize(w io.Writer) error {
	bts, err := serialize.Encode(t)
	if err != nil {
		return err
	}

	size := uint32(len(bts))
	err = binary.Write(w, binary.LittleEndian, size)
	if err != nil {
		return fmt.Errorf("failed to write size: %w", err)
	}

	_, err = w.Write(bts)
	if err != nil {
		return fmt.Errorf("failed to write tuple data: %w", err)
	}

	return nil
}

// deserialize deserializes the tuple.
// It returns the remaining data after the tuple.
func (t *Tuple) deserialize(data io.Reader) error {
	var size uint32
	err := binary.Read(data, binary.LittleEndian, &size)
	if err != nil {
		return fmt.Errorf("failed to read size: %w", err)
	}

	buf := make([]byte, size)
	_, err = io.ReadFull(data, buf)
	if err != nil {
		return fmt.Errorf("failed to read tuple data: %w", err)
	}

	return serialize.Decode(buf, &t)
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
