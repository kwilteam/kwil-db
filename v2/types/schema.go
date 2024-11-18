package types

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"kwil/types/decimal"
	"kwil/types/validation"
	"kwil/utils"
)

// Schema is a database schema that contains tables, procedures, and extensions.
type Schema struct {
	// Name is the name of the schema given by the deployer.
	Name string `json:"name"`
	// Owner is the identifier (generally an address in bytes or public key) of the owner of the schema
	Owner             HexBytes            `json:"owner"`
	Extensions        []*Extension        `json:"extensions"`
	Tables            []*Table            `json:"tables"`
	Actions           []*Action           `json:"actions"`
	Procedures        []*Procedure        `json:"procedures"`
	ForeignProcedures []*ForeignProcedure `json:"foreign_calls"`
}

func (s Schema) SerializeSize() int {
	// uint16 version + uint32 name length + name + owner length + owner +
	// uint32 extensions length + extensions + uint32 tables length + tables +
	// uint32 actions length + actions + uint32 procedures length + procedures +
	// uint32 foreign_procedures length + foreign_procedures
	size := 2 + 4 + len(s.Name) + 4 + len(s.Owner) + 4 + 4 + 4 + 4 + 4

	for _, ext := range s.Extensions {
		size += ext.SerializeSize()
	}
	for _, table := range s.Tables {
		size += table.SerializeSize()
	}
	for _, action := range s.Actions {
		size += action.SerializeSize()
	}
	for _, proc := range s.Procedures {
		size += proc.SerializeSize()
	}
	for _, fp := range s.ForeignProcedures {
		size += fp.SerializeSize()
	}
	return size
}

func (s Schema) MarshalBinary() ([]byte, error) {
	b := make([]byte, s.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(s.Name)))
	offset += 4
	copy(b[offset:], s.Name)
	offset += len(s.Name)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(s.Owner)))
	offset += 4
	copy(b[offset:], s.Owner)
	offset += len(s.Owner)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(s.Extensions)))
	offset += 4
	for _, ext := range s.Extensions {
		extData, err := ext.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], extData)
		offset += len(extData)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(s.Tables)))
	offset += 4
	for _, table := range s.Tables {
		tableData, err := table.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], tableData)
		offset += len(tableData)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(s.Actions)))
	offset += 4
	for _, action := range s.Actions {
		actionData, err := action.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], actionData)
		offset += len(actionData)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(s.Procedures)))
	offset += 4
	for _, proc := range s.Procedures {
		procData, err := proc.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], procData)
		offset += len(procData)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(s.ForeignProcedures)))
	offset += 4
	for _, fp := range s.ForeignProcedures {
		fpData, err := fp.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], fpData)
		offset += len(fpData)
	}

	return b, nil
}

func (s *Schema) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid schema data, unknown version %d", ver)
	}

	offset := 2
	nameLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+nameLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	s.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	ownerLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+ownerLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	s.Owner = make([]byte, ownerLen)
	copy(s.Owner, data[offset:offset+ownerLen])
	offset += ownerLen

	extCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if extCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	s.Extensions = make([]*Extension, extCount)
	for i := range extCount {
		ext := &Extension{}
		if err := ext.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		s.Extensions[i] = ext
		offset += ext.SerializeSize()
	}

	tableCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if tableCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	s.Tables = make([]*Table, tableCount)
	for i := range tableCount {
		table := &Table{}
		if err := table.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		s.Tables[i] = table
		offset += table.SerializeSize()
	}

	actionCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if actionCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	s.Actions = make([]*Action, actionCount)
	for i := range actionCount {
		action := &Action{}
		if err := action.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		s.Actions[i] = action
		offset += action.SerializeSize()
	}

	procCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if procCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	s.Procedures = make([]*Procedure, procCount)
	for i := range procCount {
		proc := &Procedure{}
		if err := proc.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		s.Procedures[i] = proc
		offset += proc.SerializeSize()
	}

	fpCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if fpCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	s.ForeignProcedures = make([]*ForeignProcedure, fpCount)
	for i := range fpCount {
		fp := &ForeignProcedure{}
		if err := fp.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		s.ForeignProcedures[i] = fp
		offset += fp.SerializeSize()
	}

	if offset != s.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (s *Schema) Clean() (err error) {
	err = cleanIdent(&s.Name)
	if err != nil {
		return err
	}

	// nameSet is used to check for duplicate names
	// among tables, procedures, and extensions
	nameSet := make(map[string]struct{})
	checkName := func(name string) error {
		_, ok := nameSet[name]
		if ok {
			return fmt.Errorf(`duplicate name: "%s"`, name)
		}

		nameSet[name] = struct{}{}
		return nil
	}

	for _, table := range s.Tables {
		err := table.Clean(s.Tables)
		if err != nil {
			return err
		}

		err = checkName(table.Name)
		if err != nil {
			return err
		}
	}

	for _, action := range s.Actions {
		err := action.Clean()
		if err != nil {
			return err
		}

		err = checkName(action.Name)
		if err != nil {
			return err
		}
	}

	for _, procedure := range s.Procedures {
		err := procedure.Clean()
		if err != nil {
			return err
		}

		err = checkName(procedure.Name)
		if err != nil {
			return err
		}
	}

	for _, extension := range s.Extensions {
		err := extension.Clean()
		if err != nil {
			return err
		}

		err = checkName(extension.Alias)
		if err != nil {
			return err
		}
	}

	for _, foreignCall := range s.ForeignProcedures {
		err := foreignCall.Clean()
		if err != nil {
			return err
		}

		err = checkName(foreignCall.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

// FindTable finds a table based on its name.
// It returns false if the table is not found.
func (s *Schema) FindTable(name string) (table *Table, found bool) {
	for _, tbl := range s.Tables {
		if strings.EqualFold(tbl.Name, name) {
			return tbl, true
		}
	}

	return nil, false
}

// FindAction finds an action based on its name.
// It returns false if the action is not found.
func (s *Schema) FindAction(name string) (action *Action, found bool) {
	for _, act := range s.Actions {
		if strings.EqualFold(act.Name, name) {
			return act, true
		}
	}

	return nil, false
}

// FindProcedure finds a procedure based on its name.
// It returns false if the procedure is not found.
func (s *Schema) FindProcedure(name string) (procedure *Procedure, found bool) {
	for _, proc := range s.Procedures {
		if strings.EqualFold(proc.Name, name) {
			return proc, true
		}
	}

	return nil, false
}

// FindForeignProcedure finds a foreign procedure based on its name.
// It returns false if the procedure is not found.
func (s *Schema) FindForeignProcedure(name string) (procedure *ForeignProcedure, found bool) {
	for _, proc := range s.ForeignProcedures {
		if strings.EqualFold(proc.Name, name) {
			return proc, true
		}
	}

	return nil, false
}

// FindExtensionImport finds an extension based on its alias.
// It returns false if the extension is not found.
func (s *Schema) FindExtensionImport(alias string) (extension *Extension, found bool) {
	for _, ext := range s.Extensions {
		if strings.EqualFold(ext.Alias, alias) {
			return ext, true
		}
	}

	return nil, false
}

func (s *Schema) DBID() string {
	return utils.GenerateDBID(s.Name, s.Owner)
}

// Table is a table in a database schema.
type Table struct {
	Name        string        `json:"name"`
	Columns     []*Column     `json:"columns"`
	Indexes     []*Index      `json:"indexes"`
	ForeignKeys []*ForeignKey `json:"foreign_keys"`
}

func (t Table) SerializeSize() int {
	// uint16 version + uint32 name length + name + uint32 columns length + columns + uint32 indexes length + indexes + uint32 foreignKeys length + foreignKeys
	size := 2 + 4 + len(t.Name) + 4 + 4 + 4
	for _, col := range t.Columns {
		size += col.SerializeSize()
	}
	for _, idx := range t.Indexes {
		size += idx.SerializeSize()
	}
	for _, fk := range t.ForeignKeys {
		size += fk.SerializeSize()
	}
	return size
}

func (t Table) MarshalBinary() ([]byte, error) {
	b := make([]byte, t.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(t.Name)))
	offset += 4
	copy(b[offset:], t.Name)
	offset += len(t.Name)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(t.Columns)))
	offset += 4
	for _, col := range t.Columns {
		colData, err := col.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], colData)
		offset += len(colData)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(t.Indexes)))
	offset += 4
	for _, idx := range t.Indexes {
		idxData, err := idx.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], idxData)
		offset += len(idxData)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(t.ForeignKeys)))
	offset += 4
	for _, fk := range t.ForeignKeys {
		fkData, err := fk.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], fkData)
		offset += len(fkData)
	}

	return b, nil
}

func (t *Table) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid table data, unknown version %d", ver)
	}

	offset := 2
	nameLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+nameLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	t.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	colCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if colCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	t.Columns = make([]*Column, colCount)
	for i := range colCount {
		col := &Column{}
		if err := col.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		t.Columns[i] = col
		offset += col.SerializeSize()
	}

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	idxCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if idxCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	t.Indexes = make([]*Index, idxCount)
	for i := range idxCount {
		idx := &Index{}
		if err := idx.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		t.Indexes[i] = idx
		offset += idx.SerializeSize()
	}

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	fkCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if fkCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	t.ForeignKeys = make([]*ForeignKey, fkCount)
	for i := range fkCount {
		fk := &ForeignKey{}
		if err := fk.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		t.ForeignKeys[i] = fk
		offset += fk.SerializeSize()
	}

	if offset != t.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
// It takes a slice of all tables in the schema, which is used to check for foreign key references.
func (t *Table) Clean(tables []*Table) error {
	hasPrimaryAttribute := false
	for _, col := range t.Columns {
		if err := col.Clean(); err != nil {
			return err
		}
		if col.hasPrimary() {
			if hasPrimaryAttribute {
				return fmt.Errorf("table %s has multiple primary attributes", t.Name)
			}
			hasPrimaryAttribute = true
		}
	}

	hasPrimaryIndex := false
	idxNames := make(map[string]struct{})
	for _, idx := range t.Indexes {
		if err := idx.Clean(t); err != nil {
			return err
		}

		_, ok := idxNames[idx.Name]
		if ok {
			return fmt.Errorf("table %s has multiple indexes with the same name: %s", t.Name, idx.Name)
		}
		idxNames[idx.Name] = struct{}{}

		if idx.Type == PRIMARY {
			if hasPrimaryIndex {
				return fmt.Errorf("table %s has multiple primary indexes", t.Name)
			}
			hasPrimaryIndex = true
		}
	}

	if !hasPrimaryAttribute && !hasPrimaryIndex {
		return fmt.Errorf("table %s has no primary key", t.Name)
	}

	if hasPrimaryAttribute && hasPrimaryIndex {
		return fmt.Errorf("table %s has both primary attribute and primary index", t.Name)
	}

	_, err := t.GetPrimaryKey()
	if err != nil {
		return err
	}

	for _, fk := range t.ForeignKeys {
		if err := fk.Clean(t, tables); err != nil {
			return err
		}
	}

	return cleanIdent(&t.Name)

}

// GetPrimaryKey returns the names of the column(s) that make up the primary key.
// If there is more than one, or no primary key, an error is returned.
func (t *Table) GetPrimaryKey() ([]string, error) {
	var primaryKey []string

	hasAttributePrimaryKey := false
	for _, col := range t.Columns {
		for _, attr := range col.Attributes {
			if attr.Type == PRIMARY_KEY {
				if hasAttributePrimaryKey {
					return nil, fmt.Errorf("table %s has multiple primary attributes", t.Name)
				}
				hasAttributePrimaryKey = true
				primaryKey = []string{col.Name}
			}
		}
	}

	hasIndexPrimaryKey := false
	for _, idx := range t.Indexes {
		if idx.Type == PRIMARY {
			if hasIndexPrimaryKey {
				return nil, fmt.Errorf("table %s has multiple primary indexes", t.Name)
			}
			hasIndexPrimaryKey = true

			// copy
			// if we do not copy, then the returned slice will allow modification of the index
			primaryKey = make([]string, len(idx.Columns))
			copy(primaryKey, idx.Columns)
		}
	}

	if !hasAttributePrimaryKey && !hasIndexPrimaryKey {
		return nil, fmt.Errorf("table %s has no primary key", t.Name)
	}

	if hasAttributePrimaryKey && hasIndexPrimaryKey {
		return nil, fmt.Errorf("table %s has both primary attribute and primary index", t.Name)
	}

	return primaryKey, nil
}

// Copy returns a copy of the table
func (t *Table) Copy() *Table {
	res := &Table{
		Name: t.Name,
	}

	for _, col := range t.Columns {
		res.Columns = append(res.Columns, col.Copy())
	}

	for _, idx := range t.Indexes {
		res.Indexes = append(res.Indexes, idx.Copy())
	}

	for _, fk := range t.ForeignKeys {
		res.ForeignKeys = append(res.ForeignKeys, fk.Copy())
	}

	return res
}

// FindColumn finds a column based on its name.
// It returns false if the column is not found.
func (t *Table) FindColumn(name string) (column *Column, found bool) {
	for _, col := range t.Columns {
		if strings.EqualFold(col.Name, name) {
			return col, true
		}
	}

	return nil, false
}

// Column is a column in a table.
type Column struct {
	Name       string       `json:"name"`
	Type       *DataType    `json:"type"`
	Attributes []*Attribute `json:"attributes"`
}

func (c Column) SerializeSize() int {
	// uint16 version + uint32 name length + name + type size + uint32 attributes length + attributes
	size := 2 + 4 + len(c.Name) + c.Type.SerializeSize() + 4
	for _, attr := range c.Attributes {
		size += attr.SerializeSize()
	}
	return size
}

func (c Column) MarshalBinary() ([]byte, error) {
	b := make([]byte, c.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(c.Name)))
	offset += 4
	copy(b[offset:], c.Name)
	offset += len(c.Name)

	typeData, err := c.Type.MarshalBinary()
	if err != nil {
		return nil, err
	}
	copy(b[offset:], typeData)
	offset += len(typeData)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(c.Attributes)))
	offset += 4

	for _, attr := range c.Attributes {
		attrData, err := attr.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], attrData)
		offset += len(attrData)
	}

	return b, nil
}

func (c *Column) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid column data, unknown version %d", ver)
	}

	offset := 2
	nameLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+nameLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	c.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	c.Type = &DataType{}
	if err := c.Type.UnmarshalBinary(data[offset:]); err != nil {
		return err
	}
	offset += c.Type.SerializeSize()

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	attrCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if attrCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	c.Attributes = make([]*Attribute, attrCount)
	for i := range attrCount {
		attr := &Attribute{}
		if err := attr.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		c.Attributes[i] = attr
		offset += attr.SerializeSize()
	}

	if offset != c.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

func (c *Column) Clean() error {
	for _, attr := range c.Attributes {
		if err := attr.Clean(c); err != nil {
			return err
		}
	}

	return errors.Join(
		cleanIdent(&c.Name),
		c.Type.Clean(),
	)
}

// Copy returns a copy of the column
func (c *Column) Copy() *Column {
	res := &Column{
		Name: c.Name,
		Type: c.Type.Copy(),
	}

	for _, attr := range c.Attributes {
		res.Attributes = append(res.Attributes, attr.Copy())
	}

	return res
}

// HasAttribute returns true if the column has the given attribute.
func (c *Column) HasAttribute(attr AttributeType) bool {
	for _, a := range c.Attributes {
		if a.Type == attr {
			return true
		}
	}

	return false
}

func (c *Column) hasPrimary() bool {
	for _, attr := range c.Attributes {
		if attr.Type == PRIMARY_KEY {
			return true
		}
	}
	return false
}

// Attribute is a column attribute.
// These are constraints and default values.
type Attribute struct {
	Type  AttributeType `json:"type"`
	Value string        `json:"value"`
}

func (a Attribute) SerializeSize() int {
	// uint16 version + uint32 type length + type + uint32 value length + value
	return 2 + 4 + len(a.Type) + 4 + len(a.Value)
}

func (a Attribute) MarshalBinary() ([]byte, error) {
	b := make([]byte, a.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(a.Type)))
	offset += 4
	copy(b[offset:], a.Type)
	offset += len(a.Type)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(a.Value)))
	offset += 4
	copy(b[offset:], a.Value)

	return b, nil
}

func (a *Attribute) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid attribute data, unknown version %d", ver)
	}

	offset := 2
	typeLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+typeLen+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	a.Type = AttributeType(data[offset : offset+typeLen])
	offset += typeLen

	valueLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+valueLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	a.Value = string(data[offset : offset+valueLen])
	offset += valueLen

	if offset != a.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (a *Attribute) Clean(col *Column) error {
	switch a.Type {
	case MIN, MAX:
		if !col.Type.EqualsStrict(IntType) && !col.Type.EqualsStrict(Uint256Type) && col.Type.Name != DecimalStr {
			return fmt.Errorf("attribute %s is only valid for int columns", a.Type)
		}
	case MIN_LENGTH, MAX_LENGTH:
		if !col.Type.EqualsStrict(TextType) && !col.Type.EqualsStrict(BlobType) {
			return fmt.Errorf("attribute %s is only valid for text and blob columns", a.Type)
		}
	}

	return a.Type.Clean()
}

// Copy returns a copy of the attribute
func (a *Attribute) Copy() *Attribute {
	return &Attribute{
		Type:  a.Type,
		Value: a.Value,
	}
}

// IndexType is a type of index (e.g. BTREE, UNIQUE_BTREE, PRIMARY)
type IndexType string

// Index is an index on a table.
type Index struct {
	Name    string    `json:"name"`
	Columns []string  `json:"columns"`
	Type    IndexType `json:"type"`
}

func (i Index) SerializeSize() int {
	// uint16 version + uint32 name length + name + uint32 columns length + columns + uint32 type length + type
	size := 2 + 4 + len(i.Name) + 4
	for _, col := range i.Columns {
		size += 4 + len(col)
	}
	size += 4 + len(i.Type)
	return size
}

func (i Index) MarshalBinary() ([]byte, error) {
	b := make([]byte, i.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(i.Name)))
	offset += 4
	copy(b[offset:], i.Name)
	offset += len(i.Name)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(i.Columns)))
	offset += 4

	for _, col := range i.Columns {
		binary.BigEndian.PutUint32(b[offset:], uint32(len(col)))
		offset += 4
		copy(b[offset:], col)
		offset += len(col)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(i.Type)))
	offset += 4
	copy(b[offset:], i.Type)

	return b, nil
}

func (i *Index) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid index data, unknown version %d", ver)
	}

	offset := 2
	nameLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+nameLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	i.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	colCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if colCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	i.Columns = make([]string, colCount)
	for j := range colCount {
		if len(data) < offset+4 {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		colLen := int(binary.BigEndian.Uint32(data[offset:]))
		offset += 4
		if len(data) < offset+colLen {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		i.Columns[j] = string(data[offset : offset+colLen])
		offset += colLen
	}

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	typeLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+typeLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	i.Type = IndexType(data[offset : offset+typeLen])
	offset += typeLen

	if offset != i.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (i *Index) Clean(tbl *Table) error {
	for _, col := range i.Columns {
		if !hasColumn(tbl, col) {
			return fmt.Errorf("column %s not found in table %s", col, tbl.Name)
		}
	}

	return errors.Join(
		cleanIdent(&i.Name),
		cleanIdents(&i.Columns),
		i.Type.Clean(),
	)
}

// Copy returns a copy of the index.
func (i *Index) Copy() *Index {
	return &Index{
		Name:    i.Name,
		Columns: i.Columns,
		Type:    i.Type,
	}
}

// index types
const (
	// BTREE is the default index type.
	BTREE IndexType = "BTREE"
	// UNIQUE_BTREE is a unique BTREE index.
	UNIQUE_BTREE IndexType = "UNIQUE_BTREE"
	// PRIMARY is a primary index.
	// Only one primary index is allowed per table.
	// A primary index cannot exist on a table that also has a primary key.
	PRIMARY IndexType = "PRIMARY"
)

func (i IndexType) String() string {
	return string(i)
}

func (i *IndexType) IsValid() bool {
	upper := strings.ToUpper(i.String())

	return upper == BTREE.String() ||
		upper == UNIQUE_BTREE.String() ||
		upper == PRIMARY.String()
}

func (i *IndexType) Clean() error {
	if !i.IsValid() {
		return fmt.Errorf("invalid index type: %s", i.String())
	}

	*i = IndexType(strings.ToUpper(i.String()))

	return nil
}

// ForeignKey is a foreign key in a table.
type ForeignKey struct {
	// ChildKeys are the columns that are referencing another.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "a" is the child key
	ChildKeys []string `json:"child_keys"`

	// ParentKeys are the columns that are being referred to.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "b" is the parent key
	ParentKeys []string `json:"parent_keys"`

	// ParentTable is the table that holds the parent columns.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "tbl2" is the parent table
	ParentTable string `json:"parent_table"`

	// Do we need parent schema stored with meta data or should assume and
	// enforce same schema when creating the dataset with generated DDL.
	// ParentSchema string `json:"parent_schema"`

	// Action refers to what the foreign key should do when the parent is altered.
	// This is NOT the same as a database action;
	// however sqlite's docs refer to these as actions,
	// so we should be consistent with that.
	// For example, ON DELETE CASCADE is a foreign key action
	Actions []*ForeignKeyAction `json:"actions"`
}

func (f ForeignKey) SerializeSize() int {
	// uint16 version +
	// uint32 childKeys length + childKeys +
	// uint32 parentKeys length + parentKeys +
	// uint32 parentTable length + parentTable +
	// uint32 actions length + actions
	size := 2 + 4 + 4 + 4 + len(f.ParentTable) + 4
	for _, key := range f.ChildKeys {
		size += 4 + len(key)
	}
	for _, key := range f.ParentKeys {
		size += 4 + len(key)
	}
	for _, action := range f.Actions {
		size += action.SerializeSize()
	}
	return size
}

func (f ForeignKey) MarshalBinary() ([]byte, error) {
	b := make([]byte, f.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(f.ChildKeys)))
	offset += 4
	for _, key := range f.ChildKeys {
		binary.BigEndian.PutUint32(b[offset:], uint32(len(key)))
		offset += 4
		copy(b[offset:], key)
		offset += len(key)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(f.ParentKeys)))
	offset += 4
	for _, key := range f.ParentKeys {
		binary.BigEndian.PutUint32(b[offset:], uint32(len(key)))
		offset += 4
		copy(b[offset:], key)
		offset += len(key)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(f.ParentTable)))
	offset += 4
	copy(b[offset:], f.ParentTable)
	offset += len(f.ParentTable)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(f.Actions)))
	offset += 4
	for _, action := range f.Actions {
		actionData, err := action.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], actionData)
		offset += len(actionData)
	}

	return b, nil
}

func (f *ForeignKey) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid foreign key data, unknown version %d", ver)
	}

	offset := 2
	childKeyCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if childKeyCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	f.ChildKeys = make([]string, childKeyCount)
	for i := range childKeyCount {
		if len(data) < offset+4 {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		keyLen := int(binary.BigEndian.Uint32(data[offset:]))
		offset += 4
		if len(data) < offset+keyLen {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		f.ChildKeys[i] = string(data[offset : offset+keyLen])
		offset += keyLen
	}

	if offset+4 > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	parentKeyCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if parentKeyCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	f.ParentKeys = make([]string, parentKeyCount)
	for i := range parentKeyCount {
		if len(data) < offset+4 {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		keyLen := int(binary.BigEndian.Uint32(data[offset:]))
		offset += 4
		if len(data) < offset+keyLen {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		f.ParentKeys[i] = string(data[offset : offset+keyLen])
		offset += keyLen
	}

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	tableLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+tableLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	f.ParentTable = string(data[offset : offset+tableLen])
	offset += tableLen

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	actionCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if actionCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	f.Actions = make([]*ForeignKeyAction, actionCount)
	for i := range actionCount {
		action := &ForeignKeyAction{}
		if err := action.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		f.Actions[i] = action
		offset += action.SerializeSize()
	}

	if offset != f.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

// Clean runs a set of validations and cleans the foreign key
func (f *ForeignKey) Clean(currentTable *Table, allTables []*Table) error {
	if len(f.ChildKeys) != len(f.ParentKeys) {
		return errors.New("foreign key must have same number of child and parent keys")
	}

	for _, action := range f.Actions {
		err := action.Clean()
		if err != nil {
			return err
		}
	}

	for _, childKey := range f.ChildKeys {
		if !hasColumn(currentTable, childKey) {
			return fmt.Errorf("column %s not found in table %s", childKey, currentTable.Name)
		}
	}

	found := false
	for _, table := range allTables {
		// we need to use equal fold since this can be used
		// in a case insensitive context
		if strings.EqualFold(table.Name, f.ParentTable) {
			found = true
			for _, parentKey := range f.ParentKeys {
				if !hasColumn(table, parentKey) {
					return fmt.Errorf("column %s not found in table %s", parentKey, table.Name)
				}
			}
			break
		}
	}

	if !found {
		return fmt.Errorf("parent table %s not found", f.ParentTable)
	}

	return errors.Join(
		cleanIdents(&f.ChildKeys),
		cleanIdents(&f.ParentKeys),
		// cleanIdent(&f.ParentSchema),
		cleanIdent(&f.ParentTable),
	)
}

func hasColumn(table *Table, colName string) bool {
	return slices.ContainsFunc(table.Columns, func(col *Column) bool {
		return strings.EqualFold(col.Name, colName)
	})
}

// Copy returns a copy of the foreign key
func (f *ForeignKey) Copy() *ForeignKey {
	actions := make([]*ForeignKeyAction, len(f.Actions))
	for i, action := range f.Actions {
		actions[i] = action.Copy()
	}

	return &ForeignKey{
		ChildKeys:   f.ChildKeys,
		ParentKeys:  f.ParentKeys,
		ParentTable: f.ParentTable,
		// ParentSchema: f.ParentSchema,
		Actions: actions,
	}
}

// ForeignKeyAction is used to specify what should occur
// if a parent key is updated or deleted
type ForeignKeyAction struct {
	// On can be either "UPDATE" or "DELETE"
	On ForeignKeyActionOn `json:"on"`

	// Do specifies what a foreign key action should do
	Do ForeignKeyActionDo `json:"do"`
}

func (f ForeignKeyAction) SerializeSize() int {
	// uint16 version + uint32 on length + on + uint32 do length + do
	return 2 + 4 + len(f.On) + 4 + len(f.Do)
}

func (f ForeignKeyAction) MarshalBinary() ([]byte, error) {
	b := make([]byte, f.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(f.On)))
	offset += 4
	copy(b[offset:], f.On)
	offset += len(f.On)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(f.Do)))
	offset += 4
	copy(b[offset:], f.Do)

	return b, nil
}

func (f *ForeignKeyAction) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid foreign key action data, unknown version %d", ver)
	}

	offset := 2
	onLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+onLen+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	f.On = ForeignKeyActionOn(data[offset : offset+onLen])
	offset += onLen

	doLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+doLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	f.Do = ForeignKeyActionDo(data[offset : offset+doLen])
	offset += doLen

	if offset != f.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

// Clean runs a set of validations and cleans the attributes in ForeignKeyAction
func (f *ForeignKeyAction) Clean() error {
	return errors.Join(
		f.On.Clean(),
		f.Do.Clean(),
	)
}

// Copy returns a copy of the foreign key action
func (f *ForeignKeyAction) Copy() *ForeignKeyAction {
	return &ForeignKeyAction{
		On: f.On,
		Do: f.Do,
	}
}

// ForeignKeyActionOn specifies when a foreign key action should occur.
// It can be either "UPDATE" or "DELETE".
type ForeignKeyActionOn string

// ForeignKeyActionOn types
const (
	// ON_UPDATE is used to specify an action should occur when a parent key is updated
	ON_UPDATE ForeignKeyActionOn = "UPDATE"
	// ON_DELETE is used to specify an action should occur when a parent key is deleted
	ON_DELETE ForeignKeyActionOn = "DELETE"
)

// IsValid checks whether or not the string is a valid ForeignKeyActionOn
func (f *ForeignKeyActionOn) IsValid() bool {
	upper := strings.ToUpper(f.String())
	return upper == ON_UPDATE.String() ||
		upper == ON_DELETE.String()
}

// Clean checks whether the string is valid, and will convert it to the correct case.
func (f *ForeignKeyActionOn) Clean() error {
	upper := strings.ToUpper(f.String())

	if !f.IsValid() {
		return fmt.Errorf("invalid ForeignKeyActionOn. received: %s", f.String())
	}

	*f = ForeignKeyActionOn(upper)

	return nil
}

// String returns the ForeignKeyActionOn as a string
func (f ForeignKeyActionOn) String() string {
	return string(f)
}

// ForeignKeyActionDo specifies what should be done when a foreign key action is triggered.
type ForeignKeyActionDo string

// ForeignKeyActionDo types
const (
	// DO_NO_ACTION does nothing when a parent key is altered
	DO_NO_ACTION ForeignKeyActionDo = "NO ACTION"

	// DO_RESTRICT prevents the parent key from being altered
	DO_RESTRICT ForeignKeyActionDo = "RESTRICT"

	// DO_SET_NULL sets the child key(s) to NULL
	DO_SET_NULL ForeignKeyActionDo = "SET NULL"

	// DO_SET_DEFAULT sets the child key(s) to their default values
	DO_SET_DEFAULT ForeignKeyActionDo = "SET DEFAULT"

	// DO_CASCADE updates the child key(s) or deletes the records (depending on the action type)
	DO_CASCADE ForeignKeyActionDo = "CASCADE"
)

// String returns the ForeignKeyActionDo as a string
func (f ForeignKeyActionDo) String() string {
	return string(f)
}

// IsValid checks if the string is a valid ForeignKeyActionDo
func (f *ForeignKeyActionDo) IsValid() bool {
	upper := strings.ToUpper(f.String())

	return upper == DO_NO_ACTION.String() ||
		upper == DO_RESTRICT.String() ||
		upper == DO_SET_NULL.String() ||
		upper == DO_SET_DEFAULT.String() ||
		upper == DO_CASCADE.String()
}

// Clean checks the validity or the string, and converts it to the correct case
func (f *ForeignKeyActionDo) Clean() error {
	upper := strings.ToUpper(f.String())

	if !f.IsValid() {
		return fmt.Errorf("invalid ForeignKeyActionDo. received: %s", upper)
	}

	*f = ForeignKeyActionDo(upper)

	return nil
}

// Extension defines what extensions the schema uses, and how they are initialized.
type Extension struct {
	// Name is the name of the extension registered in the node
	Name string `json:"name"`
	// Initialization is a list of key value pairs that are used to initialize the extension
	Initialization []*ExtensionConfig `json:"initialization"`
	// Alias is the alias of the extension, which is how its instance is referred to in the schema
	Alias string `json:"alias"`
}

func (e Extension) SerializeSize() int {
	// uint16 version + uint32 name length + name + uint32 initialization length + initialization + uint32 alias length + alias
	size := 2 + 4 + len(e.Name) + 4 + 4 + len(e.Alias)
	for _, init := range e.Initialization {
		size += init.SerializeSize()
	}
	return size
}

func (e Extension) MarshalBinary() ([]byte, error) {
	b := make([]byte, e.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(e.Name)))
	offset += 4
	copy(b[offset:], e.Name)
	offset += len(e.Name)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(e.Initialization)))
	offset += 4
	for _, init := range e.Initialization {
		initData, err := init.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], initData)
		offset += len(initData)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(e.Alias)))
	offset += 4
	copy(b[offset:], e.Alias)

	return b, nil
}

func (e *Extension) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid extension data, unknown version %d", ver)
	}

	offset := 2
	nameLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+nameLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	e.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	initCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if initCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	e.Initialization = make([]*ExtensionConfig, initCount)
	for i := range initCount {
		config := &ExtensionConfig{}
		if err := config.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		e.Initialization[i] = config
		offset += config.SerializeSize()
	}

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	aliasLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+aliasLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	e.Alias = string(data[offset : offset+aliasLen])
	offset += aliasLen

	if offset != e.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (e *Extension) Clean() error {
	keys := make(map[string]struct{})
	for _, config := range e.Initialization {
		_, ok := keys[config.Key]
		if ok {
			return fmt.Errorf("duplicate key %s in extension %s", config.Key, e.Name)
		}

		keys[config.Key] = struct{}{}
	}

	return errors.Join(
		cleanIdent(&e.Name),
		cleanIdent(&e.Alias),
	)
}

// CleanMap returns a map of the config values for the extension.
// Since the Kuneiform parser parses all values as strings, it cleans
// the single quotes from the values.
func (e *Extension) CleanMap() map[string]string {
	config := make(map[string]string)
	for _, c := range e.Initialization {
		config[c.Key] = strings.Trim(c.Value, "'")
	}

	return config
}

// ExtensionConfig is a key value pair that represents a configuration value for an extension
type ExtensionConfig struct {
	Key   string `json:"name"`
	Value string `json:"value"`
}

func (e ExtensionConfig) SerializeSize() int {
	// uint16 version + uint32 key length + key + uint32 value length + value
	return 2 + 4 + len(e.Key) + 4 + len(e.Value)
}

func (e ExtensionConfig) MarshalBinary() ([]byte, error) {
	b := make([]byte, e.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(e.Key)))
	offset += 4
	copy(b[offset:], e.Key)
	offset += len(e.Key)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(e.Value)))
	offset += 4
	copy(b[offset:], e.Value)

	return b, nil
}

func (e *ExtensionConfig) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid extension config data, unknown version %d", ver)
	}

	offset := 2
	keyLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+keyLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	e.Key = string(data[offset : offset+keyLen])
	offset += keyLen

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	valueLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+valueLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	e.Value = string(data[offset : offset+valueLen])
	offset += valueLen

	if offset != e.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

func cleanIdent(ident *string) error {
	if ident == nil {
		return errors.New("ident cannot be nil")
	}

	*ident = strings.TrimSpace(*ident)
	*ident = strings.ToLower(*ident)

	err := validation.ValidateIdentifier(*ident)
	if err != nil {
		return err
	}

	return nil
}

func cleanIdents(idents *[]string) error {
	if idents == nil {
		return errors.New("identifiers cannot be nil")
	}

	for i := range *idents {
		err := cleanIdent(&(*idents)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func cleanActionParameters(inputs *[]string) error {
	if inputs == nil {
		return nil
	}

	for i := range *inputs {
		err := cleanParameter(&(*inputs)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// cleanParameter applies only to the unparsed instructions/statements.
func cleanParameter(input *string) error {
	if len(*input) < 2 {
		return errors.New("parameter cannot be empty")
	}

	if len(*input) > validation.MAX_IDENT_NAME_LENGTH {
		return fmt.Errorf("parameter cannot be longer than %d characters", validation.MAX_IDENT_NAME_LENGTH)
	}

	if !strings.HasPrefix(*input, "$") {
		return errors.New("parameter must start with $")
	}

	*input = strings.ToLower(*input)

	return nil
}

// AttributeType is a type of attribute (e.g. PRIMARY_KEY, UNIQUE, NOT_NULL, DEFAULT, MIN, MAX, MIN_LENGTH, MAX_LENGTH)
type AttributeType string

// Attribute Types
const (
	PRIMARY_KEY AttributeType = "PRIMARY_KEY"
	UNIQUE      AttributeType = "UNIQUE"
	NOT_NULL    AttributeType = "NOT_NULL"
	DEFAULT     AttributeType = "DEFAULT"
	MIN         AttributeType = "MIN"
	MAX         AttributeType = "MAX"
	MIN_LENGTH  AttributeType = "MIN_LENGTH"
	MAX_LENGTH  AttributeType = "MAX_LENGTH" // is this kwil custom?
)

func (a AttributeType) String() string {
	return string(a)
}

func (a *AttributeType) IsValid() bool {
	upper := strings.ToUpper(a.String())

	return upper == PRIMARY_KEY.String() ||
		upper == UNIQUE.String() ||
		upper == NOT_NULL.String() ||
		upper == DEFAULT.String() ||
		upper == MIN.String() ||
		upper == MAX.String() ||
		upper == MIN_LENGTH.String() ||
		upper == MAX_LENGTH.String()
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (a *AttributeType) Clean() error {
	if !a.IsValid() {
		return fmt.Errorf("invalid attribute type: %s", a.String())
	}

	*a = AttributeType(strings.ToUpper(a.String()))

	return nil
}

// Action is a procedure in a database schema.
// These are defined by Kuneiform's `action` keyword.
type Action struct {
	Name        string     `json:"name"`
	Annotations []string   `json:"annotations"`
	Parameters  []string   `json:"parameters"`
	Public      bool       `json:"public"`
	Modifiers   []Modifier `json:"modifiers"`
	Body        string     `json:"body"`
}

func (a Action) SerializeSize() int {
	// uint16 version + uint32 name length + name + uint32 annotations length + annotations + uint32 parameters length + parameters + uint8 public + uint32 modifiers length + modifiers + uint32 body length + body
	size := 2 + 4 + len(a.Name) + 4 + 4 + 1 + 4 + 4 + len(a.Body)
	for _, ann := range a.Annotations {
		size += 4 + len(ann)
	}
	for _, param := range a.Parameters {
		size += 4 + len(param)
	}
	for _, mod := range a.Modifiers {
		size += 4 + len(mod)
	}
	return size
}

func (a Action) MarshalBinary() ([]byte, error) {
	b := make([]byte, a.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(a.Name)))
	offset += 4
	copy(b[offset:], a.Name)
	offset += len(a.Name)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(a.Annotations)))
	offset += 4
	for _, ann := range a.Annotations {
		binary.BigEndian.PutUint32(b[offset:], uint32(len(ann)))
		offset += 4
		copy(b[offset:], ann)
		offset += len(ann)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(a.Parameters)))
	offset += 4
	for _, param := range a.Parameters {
		binary.BigEndian.PutUint32(b[offset:], uint32(len(param)))
		offset += 4
		copy(b[offset:], param)
		offset += len(param)
	}

	if a.Public {
		b[offset] = 1
	}
	offset++

	binary.BigEndian.PutUint32(b[offset:], uint32(len(a.Modifiers)))
	offset += 4
	for _, mod := range a.Modifiers {
		binary.BigEndian.PutUint32(b[offset:], uint32(len(mod)))
		offset += 4
		copy(b[offset:], mod)
		offset += len(mod)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(a.Body)))
	offset += 4
	copy(b[offset:], a.Body)

	return b, nil
}

func (a *Action) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid action data, unknown version %d", ver)
	}

	offset := 2
	nameLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+nameLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	a.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	annCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if annCount > len(data) { // do no over allocate
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	a.Annotations = make([]string, annCount)
	for i := range annCount {
		if len(data) < offset+4 {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		annLen := int(binary.BigEndian.Uint32(data[offset:]))
		offset += 4
		if len(data) < offset+annLen {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		a.Annotations[i] = string(data[offset : offset+annLen])
		offset += annLen
	}

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	paramCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if paramCount > len(data) { // do no over allocate
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	a.Parameters = make([]string, paramCount)
	for i := range paramCount {
		if len(data) < offset+4 {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		paramLen := int(binary.BigEndian.Uint32(data[offset:]))
		offset += 4
		if len(data) < offset+paramLen {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		a.Parameters[i] = string(data[offset : offset+paramLen])
		offset += paramLen
	}

	if len(data) < offset+1 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	switch data[offset] {
	case 0:
	case 1:
		a.Public = true
	default:
		return fmt.Errorf("invalid is-public flag: %d", data[offset])
	}
	offset++

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	modCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if modCount > len(data) { // do no over allocate
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	a.Modifiers = make([]Modifier, modCount)
	for i := range modCount {
		if len(data) < offset+4 {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		modLen := int(binary.BigEndian.Uint32(data[offset:]))
		offset += 4
		if len(data) < offset+modLen {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		a.Modifiers[i] = Modifier(data[offset : offset+modLen])
		offset += modLen
	}

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	bodyLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+bodyLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	a.Body = string(data[offset : offset+bodyLen])
	offset += bodyLen

	if offset != a.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (p *Action) Clean() error {
	for _, m := range p.Modifiers {
		if err := m.Clean(); err != nil {
			return err
		}
	}

	p.Body = strings.TrimSpace(p.Body)

	return errors.Join(
		cleanIdent(&p.Name),
		cleanActionParameters(&p.Parameters),
	)
}

// IsView returns true if the procedure has a view modifier.
func (p *Action) IsView() bool {
	for _, m := range p.Modifiers {
		if m == ModifierView {
			return true
		}
	}

	return false
}

// IsOwnerOnly returns true if the procedure has an owner modifier.
func (p *Action) IsOwnerOnly() bool {
	for _, m := range p.Modifiers {
		if m == ModifierOwner {
			return true
		}
	}

	return false
}

// Modifier modifies the access to a procedure.
type Modifier string

const (
	// View means that an action does not modify the database.
	ModifierView Modifier = "VIEW"

	// Authenticated requires that the caller is identified.
	ModifierAuthenticated Modifier = "AUTHENTICATED"

	// Owner requires that the caller is the owner of the database.
	ModifierOwner Modifier = "OWNER"
)

func (m *Modifier) IsValid() bool {
	upper := strings.ToUpper(m.String())

	return upper == ModifierView.String() ||
		upper == ModifierAuthenticated.String() ||
		upper == ModifierOwner.String()
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (m *Modifier) Clean() error {
	if !m.IsValid() {
		return fmt.Errorf("invalid modifier: %s", m.String())
	}

	*m = Modifier(strings.ToUpper(m.String()))

	return nil
}

func (m Modifier) String() string {
	return string(m)
}

type Procedure struct {
	// Name is the name of the procedure.
	// It should always be lower case.
	Name string `json:"name"`

	// Parameters are the parameters of the procedure.
	Parameters []*ProcedureParameter `json:"parameters"`

	// Public is true if the procedure is public.
	Public bool `json:"public"`

	// Modifiers are the modifiers of the procedure.
	Modifiers []Modifier `json:"modifiers"`

	// Body is the body of the procedure.
	Body string `json:"body"`

	// Returns is the return type of the procedure. This may be nil.
	Returns *ProcedureReturn `json:"return_types"`
	// Annotations are the annotations of the procedure.
	Annotations []string `json:"annotations"`
}

func (p Procedure) SerializeSize() int {
	// uint16 version + uint32 name length + name + uint32 parameters length +
	//   parameters + uint8 public + uint32 modifiers length + modifiers +
	//   uint32 body length + body + return types size + uint32 annotations length + annotations
	size := 2 + 4 + len(p.Name) + 4 + 1 + 4 + 4 + len(p.Body) + 4
	for _, param := range p.Parameters {
		size += param.SerializeSize()
	}
	for _, mod := range p.Modifiers {
		size += 4 + len(mod)
	}
	if p.Returns != nil {
		size += p.Returns.SerializeSize()
	} else {
		size += 4
	}
	for _, ann := range p.Annotations {
		size += 4 + len(ann)
	}
	return size
}

func (p Procedure) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(p.Name)))
	offset += 4
	copy(b[offset:], p.Name)
	offset += len(p.Name)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(p.Parameters)))
	offset += 4
	for _, param := range p.Parameters {
		paramData, err := param.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], paramData)
		offset += len(paramData)
	}

	if p.Public {
		b[offset] = 1
	}
	offset++

	binary.BigEndian.PutUint32(b[offset:], uint32(len(p.Modifiers)))
	offset += 4
	for _, mod := range p.Modifiers {
		binary.BigEndian.PutUint32(b[offset:], uint32(len(mod)))
		offset += 4
		copy(b[offset:], mod)
		offset += len(mod)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(p.Body)))
	offset += 4
	copy(b[offset:], p.Body)
	offset += len(p.Body)

	if p.Returns == nil {
		// write math.MaxUint32 to indicate that there is no return type
		binary.BigEndian.PutUint32(b[offset:], math.MaxUint32)
		offset += 4
	} else {
		returnData, err := p.Returns.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], returnData)
		offset += len(returnData)
	}

	binary.BigEndian.PutUint32(b[offset:], uint32(len(p.Annotations)))
	offset += 4
	for _, ann := range p.Annotations {
		binary.BigEndian.PutUint32(b[offset:], uint32(len(ann)))
		offset += 4
		copy(b[offset:], ann)
		offset += len(ann)
	}

	return b, nil
}

func (p *Procedure) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid procedure data, unknown version %d", ver)
	}

	offset := 2
	nameLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+nameLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	p.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	paramCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if paramCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	p.Parameters = make([]*ProcedureParameter, paramCount)
	for i := range paramCount {
		param := &ProcedureParameter{}
		if err := param.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		p.Parameters[i] = param
		offset += param.SerializeSize()
	}

	if len(data) < offset+1 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	switch data[offset] {
	case 0:
	case 1:
		p.Public = true
	default:
		return fmt.Errorf("invalid is-public flag: %d", data[offset])
	}
	offset++

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	modCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if modCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	p.Modifiers = make([]Modifier, modCount)
	for i := range modCount {
		if len(data) < offset+4 {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		modLen := int(binary.BigEndian.Uint32(data[offset:]))
		offset += 4
		if len(data) < offset+modLen {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		p.Modifiers[i] = Modifier(data[offset : offset+modLen])
		offset += modLen
	}

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	bodyLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+bodyLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	p.Body = string(data[offset : offset+bodyLen])
	offset += bodyLen

	// compare to math.MaxUint32 to determine if there is a return type
	if binary.BigEndian.Uint32(data[offset:]) == math.MaxUint32 {
		offset += 4
	} else {
		p.Returns = &ProcedureReturn{}
		if err := p.Returns.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		offset += p.Returns.SerializeSize()
	}

	if len(data) < offset+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	annCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if annCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	p.Annotations = make([]string, annCount)
	for i := range annCount {
		if len(data) < offset+4 {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		annLen := int(binary.BigEndian.Uint32(data[offset:]))
		offset += 4
		if len(data) < offset+annLen {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		p.Annotations[i] = string(data[offset : offset+annLen])
		offset += annLen
	}

	if offset != p.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

func (p *Procedure) Clean() error {
	params := make(map[string]struct{})
	for _, param := range p.Parameters {
		err := param.Clean()
		if err != nil {
			return err
		}

		_, ok := params[param.Name]
		if ok {
			return fmt.Errorf(`duplicate parameter name: "%s"`, param.Name)
		}

		params[param.Name] = struct{}{}
	}

	if p.Returns != nil {
		err := p.Returns.Clean()
		if err != nil {
			return err
		}
	}

	p.Body = strings.TrimSpace(p.Body)

	return cleanIdent(&p.Name)
}

// IsView returns true if the procedure has a view modifier.
func (p *Procedure) IsView() bool {
	for _, m := range p.Modifiers {
		if m == ModifierView {
			return true
		}
	}

	return false
}

// IsOwnerOnly returns true if the procedure has an owner modifier.
func (p *Procedure) IsOwnerOnly() bool {
	for _, m := range p.Modifiers {
		if m == ModifierOwner {
			return true
		}
	}

	return false
}

// ProcedureReturn holds the return type of a procedure.
type ProcedureReturn struct {
	IsTable bool         `json:"is_table"`
	Fields  []*NamedType `json:"fields"`
}

func (p ProcedureReturn) SerializeSize() int {
	// uint16 version + uint8 isTable + uint32 fields length + fields
	size := 2 + 1 + 4
	for _, field := range p.Fields {
		size += field.SerializeSize()
	}
	return size
}

func (p ProcedureReturn) MarshalBinary() ([]byte, error) {
	b := make([]byte, p.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	if p.IsTable {
		b[offset] = 1
	}
	offset++

	binary.BigEndian.PutUint32(b[offset:], uint32(len(p.Fields)))
	offset += 4
	for _, field := range p.Fields {
		fieldData, err := field.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], fieldData)
		offset += len(fieldData)
	}

	return b, nil
}

func (p *ProcedureReturn) UnmarshalBinary(data []byte) error {
	if len(data) < 7 { // version(2) + isTable(1) + fieldsLen(4)
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid procedure return data, unknown version %d", ver)
	}

	offset := 2
	switch data[offset] {
	case 0:
	case 1:
		p.IsTable = true
	default:
		return fmt.Errorf("invalid is-table flag: %d", data[offset])
	}
	offset++

	fieldCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4

	if fieldCount > len(data) { // don't over-allocate
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	p.Fields = make([]*NamedType, fieldCount)
	for i := range fieldCount {
		field := &NamedType{}
		if err := field.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		p.Fields[i] = field
		offset += field.SerializeSize()
	}

	if offset != p.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

func (p *ProcedureReturn) Clean() error {
	for _, t := range p.Fields {
		return t.Clean()
	}

	return nil
}

func (p *ProcedureReturn) Copy() *ProcedureReturn {
	fields := make([]*NamedType, len(p.Fields))
	for i, field := range p.Fields {
		fields[i] = field.Copy()
	}

	return &ProcedureReturn{
		IsTable: p.IsTable,
		Fields:  fields,
	}
}

// NamedType is a single column in a
// RETURN TABLE(...) statement in a procedure.
type NamedType struct {
	// Name is the name of the column.
	Name string `json:"name"`
	// Type is the type of the column.
	Type *DataType `json:"type"`
}

func (n NamedType) SerializeSize() int {
	// uint16 version + uint32 name length + name + type size
	if n.Type == nil {
		return 0
	}
	return 2 + 4 + len(n.Name) + n.Type.SerializeSize()
}

func (n NamedType) MarshalBinary() ([]byte, error) {
	if n.Type == nil {
		return nil, errors.New("invalid procedure parameter, type is nil")
	}
	b := make([]byte, n.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(n.Name)))
	offset += 4
	copy(b[offset:], n.Name)
	offset += len(n.Name)

	if n.Type == nil {
		return nil, errors.New("invalid procedure parameter, type is nil")
	}

	typeData, err := n.Type.MarshalBinary()
	if err != nil {
		return nil, err
	}
	copy(b[offset:], typeData)

	return b, nil
}

func (n *NamedType) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid procedure parameter data, unknown version %d", ver)
	}

	offset := 2
	nameLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+nameLen {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	n.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	n.Type = &DataType{}
	if err := n.Type.UnmarshalBinary(data[offset:]); err != nil {
		return err
	}
	offset += n.Type.SerializeSize()

	if offset != n.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

func (p *NamedType) Clean() error {
	return errors.Join(
		cleanIdent(&p.Name),
		p.Type.Clean(),
	)
}

func (p *NamedType) Copy() *NamedType {
	return &NamedType{
		Name: p.Name,
		Type: p.Type.Copy(),
	}
}

// ProcedureParameter is a parameter in a procedure.
type ProcedureParameter struct {
	// Name is the name of the parameter.
	// It should always be lower case.
	Name string `json:"name"`
	// Type is the type of the parameter.
	Type *DataType `json:"type"`
}

func (p ProcedureParameter) SerializeSize() int {
	return NamedType(p).SerializeSize()
}

func (p ProcedureParameter) MarshalBinary() ([]byte, error) {
	return NamedType(p).MarshalBinary()
}

func (p *ProcedureParameter) UnmarshalBinary(data []byte) error {
	n := (*NamedType)(p)
	return n.UnmarshalBinary(data)
}

func (c *ProcedureParameter) Clean() error {
	return errors.Join(
		cleanParameter(&c.Name),
		c.Type.Clean(),
	)
}

// ForeignProcedure is used to define foreign procedures that can be
// dynamically called by the procedure.
type ForeignProcedure struct {
	// Name is the name of the foreign procedure.
	Name string `json:"name"`
	// Parameters are the parameters of the foreign procedure.
	Parameters []*DataType `json:"parameters"`
	// Returns specifies what the foreign procedure returns.
	// If it does not return a table, the names of the return
	// values are not needed, and should be left empty.
	Returns *ProcedureReturn `json:"return_types"`
}

func (f ForeignProcedure) SerializeSize() int {
	// uint16 version + uint32 name length + name + uint32 parameters length +
	//    parameters + returns nil flag + return types size
	size := 2 + 4 + len(f.Name) + 4 + 1
	for _, param := range f.Parameters {
		size += param.SerializeSize()
	}
	if f.Returns != nil {
		size += f.Returns.SerializeSize()
	}
	return size
}

func (f ForeignProcedure) MarshalBinary() ([]byte, error) {
	b := make([]byte, f.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2

	binary.BigEndian.PutUint32(b[offset:], uint32(len(f.Name)))
	offset += 4
	copy(b[offset:], f.Name)
	offset += len(f.Name)

	binary.BigEndian.PutUint32(b[offset:], uint32(len(f.Parameters)))
	offset += 4
	for _, param := range f.Parameters {
		paramData, err := param.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], paramData)
		offset += len(paramData)
	}

	if f.Returns == nil {
		b[offset] = 0
		// offset++
	} else {
		b[offset] = 1
		offset++
		returnData, err := f.Returns.MarshalBinary()
		if err != nil {
			return nil, err
		}
		copy(b[offset:], returnData)
	}

	return b, nil
}

func (f *ForeignProcedure) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid foreign procedure data, unknown version %d", ver)
	}

	offset := 2
	nameLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+nameLen+4 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	f.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	paramCount := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if paramCount > len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	f.Parameters = make([]*DataType, paramCount)
	for i := range paramCount {
		param := &DataType{}
		if err := param.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		f.Parameters[i] = param
		offset += param.SerializeSize()
	}

	if offset >= len(data) {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	var haveReturns bool
	switch data[offset] {
	case 0:
	case 1:
		haveReturns = true
	default:
		return fmt.Errorf("invalid have-returns flag: %d", data[offset])
	}
	offset++
	if haveReturns {
		if offset >= len(data) {
			return fmt.Errorf("invalid data length: %d", len(data))
		}
		f.Returns = &ProcedureReturn{}
		if err := f.Returns.UnmarshalBinary(data[offset:]); err != nil {
			return err
		}
		offset += f.Returns.SerializeSize()
	}

	if offset != f.SerializeSize() {
		return fmt.Errorf("invalid data length: %d", len(data))
	}

	return nil
}

func (f *ForeignProcedure) Clean() error {
	err := cleanIdent(&f.Name)
	if err != nil {
		return err
	}

	for _, param := range f.Parameters {
		err := param.Clean()
		if err != nil {
			return err
		}
	}

	if f.Returns != nil {
		err := f.Returns.Clean()
		if err != nil {
			return err
		}
	}

	return nil
}

// DataType is a data type.
// It includes both built-in types and user-defined types.
type DataType struct {
	// Name is the name of the type.
	Name string `json:"name"`
	// IsArray is true if the type is an array.
	IsArray bool `json:"is_array"`
	// Metadata is the metadata of the type.
	Metadata [2]uint16 `json:"metadata"`
}

func (c DataType) SerializeSize() int {
	// uint16 version + uint32 name length + name + uint8 is_array +
	//   2 x uint16 metadata
	return 2 + 4 + len(c.Name) + 1 + 4
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func (c DataType) MarshalBinary() ([]byte, error) {
	b := make([]byte, c.SerializeSize())
	const ver uint16 = 0
	binary.BigEndian.PutUint16(b, ver)
	offset := 2
	binary.BigEndian.PutUint32(b[offset:], uint32(len(c.Name)))
	offset += 4
	copy(b[offset:], c.Name)
	offset += len(c.Name)
	b[offset] = boolToByte(c.IsArray)
	offset++
	binary.BigEndian.PutUint16(b[offset:], c.Metadata[0])
	offset += 2
	binary.BigEndian.PutUint16(b[offset:], c.Metadata[1])
	return b, nil
}

func (c *DataType) UnmarshalBinary(data []byte) error {
	if len(data) < 6 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	ver := binary.BigEndian.Uint16(data)
	if ver != 0 {
		return fmt.Errorf("invalid tuple data, unknown version %d", ver)
	}
	offset := 2
	nameLen := int(binary.BigEndian.Uint32(data[offset:]))
	offset += 4
	if len(data) < offset+nameLen+1+2*2 {
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	c.Name = string(data[offset : offset+nameLen])
	offset += nameLen

	switch data[offset] {
	case 0:
	case 1:
		c.IsArray = true
	default:
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	offset++

	c.Metadata[0] = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	c.Metadata[1] = binary.BigEndian.Uint16(data[offset : offset+2])
	offset += 2
	if offset != c.SerializeSize() { // bug, must match
		return fmt.Errorf("invalid data length: %d", len(data))
	}
	return nil
}

// String returns the string representation of the type.
func (c *DataType) String() string {
	str := strings.Builder{}
	str.WriteString(c.Name)
	if c.IsArray {
		return str.String() + "[]"
	}

	if c.Name == DecimalStr {
		str.WriteString("(")
		str.WriteString(strconv.FormatUint(uint64(c.Metadata[0]), 10))
		str.WriteString(",")
		str.WriteString(strconv.FormatUint(uint64(c.Metadata[1]), 10))
		str.WriteString(")")
	}

	return str.String()
}

var ZeroMetadata = [2]uint16{}

// PGString returns the string representation of the type in Postgres.
func (c *DataType) PGString() (string, error) {
	var scalar string
	switch strings.ToLower(c.Name) {
	case intStr:
		scalar = "INT8"
	case textStr:
		scalar = "TEXT"
	case boolStr:
		scalar = "BOOL"
	case blobStr:
		scalar = "BYTEA"
	case uuidStr:
		scalar = "UUID"
	case uint256Str:
		scalar = "UINT256"
	case DecimalStr:
		if c.Metadata == ZeroMetadata {
			return "", errors.New("decimal type must have metadata")
		}

		scalar = fmt.Sprintf("NUMERIC(%d,%d)", c.Metadata[0], c.Metadata[1])
	case nullStr:
		return "", errors.New("cannot have null column type")
	case unknownStr:
		return "", errors.New("cannot have unknown column type")
	default:
		return "", fmt.Errorf("unknown column type: %s", c.Name)
	}

	if c.IsArray {
		return scalar + "[]", nil
	}

	return scalar, nil
}

func (c *DataType) Clean() error {
	lName := strings.ToLower(c.Name)
	switch lName {
	case intStr, textStr, boolStr, blobStr, uuidStr, uint256Str: // ok
		if c.Metadata != ZeroMetadata {
			return fmt.Errorf("type %s cannot have metadata", c.Name)
		}
	case DecimalStr:
		if c.Metadata == ZeroMetadata {
			return errors.New("decimal type must have metadata")
		}

		err := decimal.CheckPrecisionAndScale(c.Metadata[0], c.Metadata[1])
		if err != nil {
			return err
		}
	case nullStr, unknownStr:
		if c.IsArray {
			return fmt.Errorf("type %s cannot be an array", c.Name)
		}

		if c.Metadata != ZeroMetadata {
			return fmt.Errorf("type %s cannot have metadata", c.Name)
		}
	default:
		return fmt.Errorf("unknown type: %s", c.Name)
	}

	c.Name = lName

	return nil
}

// Copy returns a copy of the type.
func (c *DataType) Copy() *DataType {
	d := &DataType{
		Name:     c.Name,
		IsArray:  c.IsArray,
		Metadata: c.Metadata,
	}

	return d
}

// EqualsStrict returns true if the type is equal to the other type.
// The types must be exactly the same, including metadata.
func (c *DataType) EqualsStrict(other *DataType) bool {
	// if unknown, return true. unknown is a special case used
	// internally when type checking is disabled.
	if c.Name == unknownStr || other.Name == unknownStr {
		return true
	}

	if c.IsArray != other.IsArray {
		return false
	}

	if (c.Metadata == ZeroMetadata) != (other.Metadata == ZeroMetadata) {
		return false
	}
	if c.Metadata != ZeroMetadata {
		if c.Metadata[0] != other.Metadata[0] || c.Metadata[1] != other.Metadata[1] {
			return false
		}
	}

	return strings.EqualFold(c.Name, other.Name)
}

// Equals returns true if the type is equal to the other type, or if either type is null.
func (c *DataType) Equals(other *DataType) bool {
	if c.Name == nullStr || other.Name == nullStr {
		return true
	}

	return c.EqualsStrict(other)
}

func (c *DataType) IsNumeric() bool {
	if c.IsArray {
		return false
	}

	return c.Name == intStr || c.Name == DecimalStr || c.Name == uint256Str || c.Name == unknownStr
}

// declared DataType constants.
// We do not have one for fixed because fixed types require metadata.
var (
	IntType = &DataType{
		Name: intStr,
	}
	IntArrayType = ArrayType(IntType)
	TextType     = &DataType{
		Name: textStr,
	}
	TextArrayType = ArrayType(TextType)
	BoolType      = &DataType{
		Name: boolStr,
	}
	BoolArrayType = ArrayType(BoolType)
	BlobType      = &DataType{
		Name: blobStr,
	}
	BlobArrayType = ArrayType(BlobType)
	UUIDType      = &DataType{
		Name: uuidStr,
	}
	UUIDArrayType = ArrayType(UUIDType)
	// DecimalType contains 1,0 metadata.
	// For type detection, users should prefer compare a datatype
	// name with the DecimalStr constant.
	DecimalType = &DataType{
		Name:     DecimalStr,
		Metadata: [2]uint16{1, 0}, // the minimum precision and scale
	}
	DecimalArrayType = ArrayType(DecimalType)
	Uint256Type      = &DataType{
		Name: uint256Str,
	}
	Uint256ArrayType = ArrayType(Uint256Type)
	// NullType is a special type used internally
	NullType = &DataType{
		Name: nullStr,
	}
	// Unknown is a special type used internally
	// when a type is unknown until runtime.
	UnknownType = &DataType{
		Name: unknownStr,
	}
)

// ArrayType creates an array type of the given type.
// It panics if the type is already an array.
func ArrayType(t *DataType) *DataType {
	if t.IsArray {
		panic("cannot create an array of an array")
	}
	return &DataType{
		Name:     t.Name,
		IsArray:  true,
		Metadata: t.Metadata,
	}
}

const (
	textStr    = "text"
	intStr     = "int"
	boolStr    = "bool"
	blobStr    = "blob"
	uuidStr    = "uuid"
	uint256Str = "uint256"
	// DecimalStr is a fixed point number.
	DecimalStr = "decimal"
	nullStr    = "null"
	unknownStr = "unknown"
)

// NewDecimalType creates a new fixed point decimal type.
func NewDecimalType(precision, scale uint16) (*DataType, error) {
	err := decimal.CheckPrecisionAndScale(precision, scale)
	if err != nil {
		return nil, err
	}

	return &DataType{
		Name:     DecimalStr,
		Metadata: [2]uint16{precision, scale},
	}, nil
}
