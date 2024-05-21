package types

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/kwilteam/kwil-db/core/types/decimal"
	"github.com/kwilteam/kwil-db/core/types/validation"
	"github.com/kwilteam/kwil-db/core/utils"
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

// Clean validates rules about the data in the struct (naming conventions, syntax, etc.).
func (a *Attribute) Clean(col *Column) error {
	switch a.Type {
	case MIN, MAX:
		if !col.Type.EqualsStrict(IntType) {
			return fmt.Errorf("attribute %s is only valid for int columns", a.Type)
		}
	case MIN_LENGTH, MAX_LENGTH:
		if !col.Type.EqualsStrict(TextType) {
			return fmt.Errorf("attribute %s is only valid for text columns", a.Type)
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

// Clean runs a set of validations and cleans the foreign key
func (f *ForeignKey) Clean(currentTable *Table, allTables []*Table) error {
	if len(f.ChildKeys) != len(f.ParentKeys) {
		return fmt.Errorf("foreign key must have same number of child and parent keys")
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

	return errors.Join(
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

func cleanIdent(ident *string) error {
	if ident == nil {
		return fmt.Errorf("ident cannot be nil")
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
		return fmt.Errorf("parameter cannot be empty")
	}

	if len(*input) > validation.MAX_IDENT_NAME_LENGTH {
		return fmt.Errorf("parameter cannot be longer than %d characters", validation.MAX_IDENT_NAME_LENGTH)
	}

	if !strings.HasPrefix(*input, "$") {
		return fmt.Errorf("parameter must start with $")
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

	// Returns is the return type of the procedure.
	Returns *ProcedureReturn `json:"return_types"`
	// Annotations are the annotations of the procedure.
	Annotations []string `json:"annotations"`
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
// EITHER the Type field is set, OR the Table field is set.
type ProcedureReturn struct {
	IsTable bool         `json:"is_table"`
	Fields  []*NamedType `json:"fields"`
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

// ProcedureParameter is a parameter in a procedure.
type ProcedureParameter struct {
	// Name is the name of the parameter.
	// It should always be lower case.
	Name string `json:"name"`
	// Type is the type of the parameter.
	Type *DataType `json:"type"`
}

func (c *ProcedureParameter) Clean() error {
	return errors.Join(
		cleanParameter(&c.Name),
		c.Type.Clean(),
	)
}

// NamedType is a single column in a
// RETURN TABLE(...) statement in a procedure.
type NamedType struct {
	// Name is the name of the column.
	Name string `json:"name"`
	// Type is the type of the column.
	Type *DataType `json:"type"`
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
	Returns *ProcedureReturn `json:"returns"`
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
	Metadata any `json:"metadata"`
}

// String returns the string representation of the type.
func (c *DataType) String() string {
	str := strings.Builder{}
	str.WriteString(c.Name)
	if c.IsArray {
		return str.String() + "[]"
	}

	if c.Name == DecimalStr {
		data, ok := c.Metadata.([2]uint16)
		if ok {
			str.WriteString("(")
			str.WriteString(fmt.Sprint(data[0]))
			str.WriteString(",")
			str.WriteString(fmt.Sprint(data[1]))
			str.WriteString(")")
		}
	}

	return str.String()
}

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
		data, ok := c.Metadata.([2]uint16)
		if !ok {
			// should never happen, since Clean() should have caught this
			return "", fmt.Errorf("fixed type must have metadata of type [2]uint8")
		}

		scalar = fmt.Sprintf("NUMERIC(%d,%d)", data[0], data[1])
	case nullStr:
		return "", fmt.Errorf("cannot have null column type")
	case unknownStr:
		return "", fmt.Errorf("cannot have unknown column type")
	default:
		return "", fmt.Errorf("unknown column type: %s", c.Name)
	}

	if c.IsArray {
		return scalar + "[]", nil
	}

	return scalar, nil
}

func (c *DataType) Clean() error {
	c.Name = strings.ToLower(c.Name)
	switch c.Name {
	case intStr, textStr, boolStr, blobStr, uuidStr, uint256Str: // ok
		if c.Metadata != nil {
			return fmt.Errorf("type %s cannot have metadata", c.Name)
		}

		return nil
	case DecimalStr:
		data, ok := c.Metadata.([2]uint16)
		if !ok {
			return fmt.Errorf("fixed type must have metadata of type [2]uint8")
		}

		err := decimal.CheckPrecisionAndScale(data[0], data[1])
		if err != nil {
			return err
		}

		return nil
	case nullStr, unknownStr:
		if c.IsArray {
			return fmt.Errorf("type %s cannot be an array", c.Name)
		}

		if c.Metadata != nil {
			return fmt.Errorf("type %s cannot have metadata", c.Name)
		}

		return nil
	default:
		return fmt.Errorf("unknown type: %s", c.Name)
	}
}

// Copy returns a copy of the type.
func (c *DataType) Copy() *DataType {
	return &DataType{
		Name:     c.Name,
		IsArray:  c.IsArray,
		Metadata: c.Metadata,
	}
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

	if c.Metadata != other.Metadata {
		return false
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
	return c.Name == intStr || c.Name == DecimalStr || c.Name == uint256Str || c.Name == unknownStr
}

// declared DataType constants.
// We do not have one for fixed because fixed types require metadata.
var (
	IntType = &DataType{
		Name: intStr,
	}
	TextType = &DataType{
		Name: textStr,
	}
	BoolType = &DataType{
		Name: boolStr,
	}
	BlobType = &DataType{
		Name: blobStr,
	}
	UUIDType = &DataType{
		Name: uuidStr,
	}
	Uint256Type = &DataType{
		Name: uint256Str,
	}
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
	DecimalStr = "fixed"
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
