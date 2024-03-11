package common

import (
	"errors"
	"fmt"
	"strings"

	"github.com/kwilteam/kwil-db/common/validation"
	"github.com/kwilteam/kwil-db/core/utils"
)

// Schema is a database schema that contains tables, procedures, and extensions.
type Schema struct {
	// Name is the name of the schema given by the deployer.
	Name string `json:"name"`
	// Owner is the identifier (generally an address in bytes or public key) of the owner of the schema
	Owner      []byte       `json:"owner"`
	Extensions []*Extension `json:"extensions"`
	Tables     []*Table     `json:"tables"`
	Procedures []*Procedure `json:"procedures"`
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (s *Schema) Clean() error {
	err := cleanIdent(&s.Name)
	if err != nil {
		return err
	}

	tableSet := make(map[string]struct{})
	for _, table := range s.Tables {
		err := table.Clean()
		if err != nil {
			return err
		}

		_, ok := tableSet[table.Name]
		if ok {
			return fmt.Errorf(`duplicate table name: "%s"`, table.Name)
		}

		tableSet[table.Name] = struct{}{}
	}

	procedureSet := make(map[string]struct{})
	for _, action := range s.Procedures {
		err := action.Clean()
		if err != nil {
			return err
		}

		_, ok := procedureSet[action.Name]
		if ok {
			return fmt.Errorf(`duplicate procedure name: "%s"`, action.Name)
		}

		procedureSet[action.Name] = struct{}{}
	}

	extensionSet := make(map[string]struct{})
	for _, extension := range s.Extensions {
		err := extension.Clean()
		if err != nil {
			return err
		}

		_, ok := extensionSet[extension.Alias]
		if ok {
			return fmt.Errorf(`duplicate extension alias: "%s"`, extension.Alias)
		}

		extensionSet[extension.Alias] = struct{}{}
	}

	return nil
}

func (s *Schema) DBID() string {
	return utils.GenerateDBID(s.Name, s.Owner)
}

// Table is a table in a database schema.
type Table struct {
	Name        string        `json:"name"`
	Columns     []*Column     `json:"columns"`
	Indexes     []*Index      `json:"indexes,omitempty"`
	ForeignKeys []*ForeignKey `json:"foreign_keys"`
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (t *Table) Clean() error {
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
	for _, idx := range t.Indexes {
		if err := idx.Clean(); err != nil {
			return err
		}

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

	return runCleans(
		cleanIdent(&t.Name),
	)
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

// Column is a column in a table.
type Column struct {
	Name       string       `json:"name"`
	Type       DataType     `json:"type"`
	Attributes []*Attribute `json:"attributes,omitempty"`
}

func (c *Column) Clean() error {
	for _, attr := range c.Attributes {
		if err := attr.Clean(); err != nil {
			return err
		}
	}

	return runCleans(
		cleanIdent(&c.Name),
		c.Type.Clean(),
	)
}

// Copy returns a copy of the column
func (c *Column) Copy() *Column {
	res := &Column{
		Name: c.Name,
		Type: c.Type,
	}

	for _, attr := range c.Attributes {
		res.Attributes = append(res.Attributes, attr.Copy())
	}

	return res
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
	Value string        `json:"value,omitempty"`
}

func (a *Attribute) Clean() error {
	return runCleans(
		a.Type.Clean(),
	)
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

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (i *Index) Clean() error {
	return runCleans(
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

// Clean runs a set of validations and cleans the foreign key
func (f *ForeignKey) Clean() error {
	if len(f.ChildKeys) != len(f.ParentKeys) {
		return fmt.Errorf("foreign key must have same number of child and parent keys")
	}

	for _, action := range f.Actions {
		err := action.Clean()
		if err != nil {
			return err
		}
	}

	return runCleans(
		cleanIdents(&f.ChildKeys),
		cleanIdents(&f.ParentKeys),
		// cleanIdent(&f.ParentSchema),
		cleanIdent(&f.ParentTable),
	)
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

// Clean runs a set of validations and cleans the attributes in ForeignKeyAction
func (f *ForeignKeyAction) Clean() error {
	return runCleans(
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
		return fmt.Errorf("invalid ForeginKeyActionOn. received: %s", f.String())
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

	return runCleans(
		cleanIdent(&e.Name),
		cleanIdent(&e.Alias),
	)
}

// CleanMap returns a map of the config values for the extension.
// Since the Kueiform parser parses all values as strings, it cleans
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

// runCleans runs a series of clean functions and returns the first error.
func runCleans(errs ...error) error {
	return errors.Join(errs...)
}

func cleanIdent(ident *string) error {
	err := cleanString(ident)
	if err != nil {
		return err
	}

	err = validation.ValidateIdentifier(*ident)
	if err != nil {
		return err
	}

	return nil
}

func cleanDBID(dbid *string) error {
	err := cleanString(dbid)
	if err != nil {
		return err
	}

	err = validation.ValidateDBID(*dbid)
	if err != nil {
		return err
	}

	return nil
}

// cleanString cleans a string by trimming whitespace and making it lowercase.
// It returns an error if the string is nil.
func cleanString(str *string) error {
	if str == nil {
		return fmt.Errorf("string cannot be nil")
	}

	*str = strings.TrimSpace(*str)
	*str = strings.ToLower(*str)

	return nil
}

func cleanIdents(idents *[]string) error {
	if idents == nil {
		return fmt.Errorf("identifiers cannot be nil")
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
		err := cleanActionParameter(&(*inputs)[i])
		if err != nil {
			return err
		}
	}

	return nil
}

// cleanActionParameter applies only to the unparsed instructions/statements.
func cleanActionParameter(input *string) error {
	if len(*input) == 0 {
		return fmt.Errorf("action parameter cannot be empty")
	}

	if len(*input) > validation.MAX_IDENT_NAME_LENGTH {
		return fmt.Errorf("action parameter cannot be longer than %d characters", validation.MAX_IDENT_NAME_LENGTH)
	}

	if !strings.HasPrefix(*input, "$") {
		return fmt.Errorf("action parameter must start with $")
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

// DataType is a type of data (e.g. NULL, TEXT, INT, BLOB, BOOLEAN)
type DataType string

// Data types
const (
	NULL DataType = "NULL"
	TEXT DataType = "TEXT"
	INT  DataType = "INT"
	BLOB DataType = "BLOB"
	BOOL DataType = "BOOLEAN"
)

func (d DataType) String() string {
	return string(d)
}

func (d *DataType) IsNumeric() bool {
	return *d == INT
}

func (d *DataType) IsValid() bool {
	upper := strings.ToUpper(d.String())

	return upper == NULL.String() ||
		upper == TEXT.String() ||
		upper == INT.String() ||
		upper == BOOL.String() ||
		upper == BLOB.String()

}

// will check if the data type is a text type
func (d *DataType) IsText() bool {
	return *d == TEXT
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (d *DataType) Clean() error {
	if !d.IsValid() {
		return fmt.Errorf("invalid data type: %s", d.String())
	}

	*d = DataType(strings.ToUpper(d.String()))

	return nil
}

// Procedure is a procedure in a database schema.
// These are defined by Kuneiform's `action` keyword.
type Procedure struct {
	Name        string     `json:"name"`
	Annotations []string   `json:"annotations,omitempty"`
	Args        []string   `json:"inputs"`
	Public      bool       `json:"public"`
	Modifiers   []Modifier `json:"modifiers"`
	Statements  []string   `json:"statements"`
}

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (p *Procedure) Clean() error {
	for _, m := range p.Modifiers {
		if err := m.Clean(); err != nil {
			return err
		}
	}

	return runCleans(
		cleanIdent(&p.Name),
		cleanActionParameters(&p.Args),
	)
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
