package transactions

import (
	"github.com/kwilteam/kwil-db/core/types"
	"github.com/kwilteam/kwil-db/core/types/serialize"
)

var _ Payload = (*Schema)(nil)

func (s *Schema) MarshalBinary() (serialize.SerializedData, error) {
	return serialize.Encode(s)
}

func (s *Schema) UnmarshalBinary(b serialize.SerializedData) error {
	result, err := serialize.Decode[Schema](b)
	if err != nil {
		return err
	}

	*s = *result
	return nil
}

func (s *Schema) Type() PayloadType {
	return PayloadTypeDeploySchema
}

// Schema is a database schema that contains tables, procedures, and extensions.
type Schema struct {
	// Name is the name of the schema given by the deployer.
	Name string
	// Owner is the identifier (generally an address in bytes or public key) of the owner of the schema
	Owner             []byte
	Extensions        []*Extension
	Tables            []*Table
	Actions           []*Action
	Procedures        []*Procedure
	ForeignProcedures []*ForeignProcedure
}

// Table is a table in a database schema.
type Table struct {
	Name        string
	Columns     []*Column
	Indexes     []*Index
	ForeignKeys []*ForeignKey
}

// Column is a column in a table.
type Column struct {
	Name       string
	Type       *DataType
	Attributes []*Attribute
}

// Attribute is a column attribute.
// These are constraints and default values.
type Attribute struct {
	Type  string
	Value string
}

// Index is an index on a table.
type Index struct {
	Name    string
	Columns []string
	Type    string
}

// ForeignKey is a foreign key in a table.
type ForeignKey struct {
	// ChildKeys are the columns that are referencing another.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "a" is the child key
	ChildKeys []string

	// ParentKeys are the columns that are being referred to.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "b" is the parent key
	ParentKeys []string

	// ParentTable is the table that holds the parent columns.
	// For example, in FOREIGN KEY (a) REFERENCES tbl2(b), "tbl2" is the parent table
	ParentTable string

	// Do we need parent schema stored with meta data or should assume and
	// enforce same schema when creating the dataset with generated DDL.
	// ParentSchema string `json:"parent_schema"`

	// Action refers to what the foreign key should do when the parent is altered.
	// This is NOT the same as a database action;
	// however sqlite's docs refer to these as actions,
	// so we should be consistent with that.
	// For example, ON DELETE CASCADE is a foreign key action
	Actions []*ForeignKeyAction
}

// ForeignKeyAction is used to specify what should occur
// if a parent key is updated or deleted
type ForeignKeyAction struct {
	// On can be either "UPDATE" or "DELETE"
	On string

	// Do specifies what a foreign key action should do
	Do string
}

// Extension defines what extensions the schema uses, and how they are initialized.
type Extension struct {
	// Name is the name of the extension registered in the node
	Name string
	// Initialization is a list of key value pairs that are used to initialize the extension
	Initialization []*ExtensionConfig
	// Alias is the alias of the extension, which is how its instance is referred to in the schema
	Alias string
}

// ExtensionConfig is a key value pair that represents a configuration value for an extension
type ExtensionConfig struct {
	Key   string
	Value string
}

// Action is a procedure in a database schema.
// These are defined by Kuneiform's `action` keyword.
type Action struct {
	Name        string
	Annotations []string
	Parameters  []string
	Public      bool
	Modifiers   []string
	Body        string
}

type Procedure struct {
	// Name is the name of the procedure.
	// It should always be lower case.
	Name string

	// Parameters are the parameters of the procedure.
	Parameters []*NamedType

	// Public is true if the procedure is public.
	Public bool

	// Modifiers are the modifiers of the procedure.
	Modifiers []string

	// Body is the body of the procedure.
	Body string

	// Returns is the return type of the procedure.
	Returns *ProcedureReturn `rlp:"nil"`
	// Annotations are the annotations of the procedure.
	Annotations []string `rlp:"optional"`
}

// ProcedureReturn holds the return type of a procedure.
// Either one of the types can bet set, however, not both.
type ProcedureReturn struct {
	IsTable bool         `rlp:"optional"`
	Types   []*NamedType `rlp:"optional"`
}

// NamedType is a field of a composite type.
type NamedType struct {
	// Name is the name of the parameter.
	// It should always be lower case.
	Name string
	// Type is the type of the parameter.
	Type *DataType
}

// Import is an import statement in a schema.
type Import struct {
	// Target is the database ID or name of the schema being imported.
	Target string
	// Alias is the alias of the schema being imported.
	Alias string
}

// DataType is a data type.
// It includes both built-in types and user-defined types.
type DataType struct {
	// Name is the name of the type.
	Name string
	// IsArray is true if the type is an array.
	IsArray bool
	// Metadata is the metadata of the type.
	Metadata [2]uint16 `rlp:"optional"`
}

// ForeignProcedure is a foreign procedure call in a database
// schema. These are defined by Kuneiform's `call` keyword.
type ForeignProcedure struct {
	// Name is the name of the procedure.
	Name string
	// Parameters are the parameters of the procedure.
	Parameters []*DataType
	// Returns is the return type of the procedure.
	Returns *ProcedureReturn `rlp:"nil"`
}

// ToTypes converts the type to the core type

func (s *Schema) ToTypes() (s2 *types.Schema, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()

	s2 = &types.Schema{
		Name:  s.Name,
		Owner: s.Owner,
	}

	for _, table := range s.Tables {
		s2.Tables = append(s2.Tables, table.toTypes())
	}

	for _, action := range s.Actions {
		s2.Actions = append(s2.Actions, action.toTypes())
	}

	for _, procedure := range s.Procedures {
		s2.Procedures = append(s2.Procedures, procedure.toTypes())
	}

	for _, extension := range s.Extensions {
		s2.Extensions = append(s2.Extensions, extension.toTypes())
	}

	for _, foreignProcedure := range s.ForeignProcedures {
		s2.ForeignProcedures = append(s2.ForeignProcedures, foreignProcedure.toTypes())
	}

	return s2, s2.Clean()
}

// transactions converts the type to an RLP serializable type.

func (t *Table) toTypes() *types.Table {
	t2 := &types.Table{
		Name: t.Name,
	}

	for _, col := range t.Columns {
		t2.Columns = append(t2.Columns, col.toTypes())
	}

	for _, idx := range t.Indexes {
		t2.Indexes = append(t2.Indexes, idx.toTypes())
	}

	for _, fk := range t.ForeignKeys {
		t2.ForeignKeys = append(t2.ForeignKeys, fk.toTypes())
	}

	return t2
}

func (c *Column) toTypes() *types.Column {
	c2 := &types.Column{
		Name: c.Name,
		Type: c.Type.toTypes(),
	}

	for _, attr := range c.Attributes {
		c2.Attributes = append(c2.Attributes, attr.toTypes())
	}

	return c2
}

func (a *Attribute) toTypes() *types.Attribute {
	return &types.Attribute{
		Type:  types.AttributeType(a.Type),
		Value: a.Value,
	}
}

func (i *Index) toTypes() *types.Index {
	return &types.Index{
		Name:    i.Name,
		Columns: i.Columns,
		Type:    types.IndexType(i.Type),
	}
}

func (f *ForeignKey) toTypes() *types.ForeignKey {
	actions := make([]*types.ForeignKeyAction, len(f.Actions))
	for i, action := range f.Actions {
		actions[i] = action.toTypes()
	}

	return &types.ForeignKey{
		ChildKeys:   f.ChildKeys,
		ParentKeys:  f.ParentKeys,
		ParentTable: f.ParentTable,
		Actions:     actions,
	}
}

func (f *ForeignKeyAction) toTypes() *types.ForeignKeyAction {
	return &types.ForeignKeyAction{
		On: types.ForeignKeyActionOn(f.On),
		Do: types.ForeignKeyActionDo(f.Do),
	}
}

func (e *Extension) toTypes() *types.Extension {
	e2 := &types.Extension{
		Name:  e.Name,
		Alias: e.Alias,
	}

	for _, config := range e.Initialization {
		e2.Initialization = append(e2.Initialization, config.toTypes())
	}

	return e2
}

func (e *ExtensionConfig) toTypes() *types.ExtensionConfig {
	return &types.ExtensionConfig{
		Key:   e.Key,
		Value: e.Value,
	}
}

func (p *Action) toTypes() *types.Action {
	t := &types.Action{
		Name:        p.Name,
		Annotations: p.Annotations,
		Parameters:  p.Parameters,
		Public:      p.Public,
		Body:        p.Body,
	}

	for _, m := range p.Modifiers {
		t.Modifiers = append(t.Modifiers, types.Modifier(m))
	}

	return t
}

func (p *Procedure) toTypes() *types.Procedure {
	t := &types.Procedure{
		Name:        p.Name,
		Annotations: p.Annotations,
		Public:      p.Public,
		Body:        p.Body,
	}

	for _, m := range p.Modifiers {
		t.Modifiers = append(t.Modifiers, types.Modifier(m))
	}

	for _, param := range p.Parameters {
		t.Parameters = append(t.Parameters, param.parameterType())
	}

	if p.Returns != nil {
		t.Returns = p.Returns.toTypes()
	}

	return t
}

func (p *ProcedureReturn) toTypes() *types.ProcedureReturn {
	tps := make([]*types.NamedType, len(p.Types))
	for i, t := range p.Types {
		tps[i] = t.toTypes()
	}

	return &types.ProcedureReturn{
		IsTable: p.IsTable,
		Fields:  tps,
	}
}

func (p *NamedType) toTypes() *types.NamedType {
	return &types.NamedType{
		Name: p.Name,
		Type: p.Type.toTypes(),
	}
}

func (p *NamedType) parameterType() *types.ProcedureParameter {
	return &types.ProcedureParameter{
		Name: p.Name,
		Type: p.Type.toTypes(),
	}
}

func (c *DataType) toTypes() *types.DataType {
	return &types.DataType{
		Name:     c.Name,
		IsArray:  c.IsArray,
		Metadata: c.Metadata,
	}
}

func (f *ForeignProcedure) toTypes() *types.ForeignProcedure {
	params := make([]*types.DataType, len(f.Parameters))
	for i, param := range f.Parameters {
		params[i] = param.toTypes()
	}

	var returns *types.ProcedureReturn
	if f.Returns != nil {
		returns = f.Returns.toTypes()
	}

	return &types.ForeignProcedure{
		Name:       f.Name,
		Parameters: params,
		Returns:    returns,
	}
}

// fromTypes converts the core type to the RLP serializable type.

func (s *Schema) FromTypes(s2 *types.Schema) {
	s.Name = s2.Name
	s.Owner = s2.Owner

	for _, table := range s2.Tables {
		t := &Table{}
		t.fromTypes(table)
		s.Tables = append(s.Tables, t)
	}

	for _, action := range s2.Actions {
		a := &Action{}
		a.fromTypes(action)
		s.Actions = append(s.Actions, a)
	}

	for _, procedure := range s2.Procedures {
		p := &Procedure{}
		p.fromTypes(procedure)
		s.Procedures = append(s.Procedures, p)
	}

	for _, extension := range s2.Extensions {
		e := &Extension{}
		e.fromTypes(extension)
		s.Extensions = append(s.Extensions, e)
	}

	for _, foreignProcedure := range s2.ForeignProcedures {
		f := &ForeignProcedure{}
		f.fromTypes(foreignProcedure)
		s.ForeignProcedures = append(s.ForeignProcedures, f)
	}
}

func (t *Table) fromTypes(t2 *types.Table) {
	t.Name = t2.Name

	for _, col := range t2.Columns {
		c := &Column{}
		c.fromTypes(col)
		t.Columns = append(t.Columns, c)
	}

	for _, idx := range t2.Indexes {
		i := &Index{}
		i.fromTypes(idx)
		t.Indexes = append(t.Indexes, i)
	}

	for _, fk := range t2.ForeignKeys {
		f := &ForeignKey{}
		f.fromTypes(fk)
		t.ForeignKeys = append(t.ForeignKeys, f)
	}
}

func (c *Column) fromTypes(c2 *types.Column) {
	c.Name = c2.Name
	c.Type = &DataType{}
	c.Type.fromTypes(c2.Type)

	for _, attr := range c2.Attributes {
		a := &Attribute{}
		a.fromTypes(attr)
		c.Attributes = append(c.Attributes, a)
	}
}

func (a *Attribute) fromTypes(a2 *types.Attribute) {
	a.Type = a2.Type.String()
	a.Value = a2.Value
}

func (i *Index) fromTypes(i2 *types.Index) {
	i.Name = i2.Name
	i.Columns = i2.Columns
	i.Type = i2.Type.String()
}

func (f *ForeignKey) fromTypes(f2 *types.ForeignKey) {
	f.ChildKeys = f2.ChildKeys
	f.ParentKeys = f2.ParentKeys
	f.ParentTable = f2.ParentTable

	for _, action := range f2.Actions {
		a := &ForeignKeyAction{}
		a.fromTypes(action)
		f.Actions = append(f.Actions, a)
	}
}

func (f *ForeignKeyAction) fromTypes(f2 *types.ForeignKeyAction) {
	f.On = f2.On.String()
	f.Do = f2.Do.String()
}

func (e *Extension) fromTypes(e2 *types.Extension) {
	e.Name = e2.Name
	e.Alias = e2.Alias

	for _, config := range e2.Initialization {
		c := &ExtensionConfig{}
		c.fromTypes(config)
		e.Initialization = append(e.Initialization, c)
	}
}

func (e *ExtensionConfig) fromTypes(e2 *types.ExtensionConfig) {
	e.Key = e2.Key
	e.Value = e2.Value
}

func (p *Action) fromTypes(p2 *types.Action) {
	p.Name = p2.Name
	p.Annotations = p2.Annotations
	p.Parameters = p2.Parameters
	p.Public = p2.Public
	p.Body = p2.Body

	for _, m := range p2.Modifiers {
		p.Modifiers = append(p.Modifiers, m.String())
	}
}

func (p *Procedure) fromTypes(p2 *types.Procedure) {
	p.Name = p2.Name
	p.Annotations = p2.Annotations
	p.Public = p2.Public
	p.Body = p2.Body

	for _, m := range p2.Modifiers {
		p.Modifiers = append(p.Modifiers, m.String())
	}

	for _, param := range p2.Parameters {
		pp := &NamedType{}
		pp.fromParameter(param)
		p.Parameters = append(p.Parameters, pp)
	}

	if p2.Returns != nil {
		p.Returns = &ProcedureReturn{}
		p.Returns.fromTypes(p2.Returns)
	}
}

func (p *NamedType) fromTypes(p2 *types.NamedType) {
	p.Name = p2.Name
	p.Type = &DataType{}
	p.Type.fromTypes(p2.Type)
}

func (p *NamedType) fromParameter(p2 *types.ProcedureParameter) {
	p.Name = p2.Name
	p.Type = &DataType{}
	p.Type.fromTypes(p2.Type)
}

func (p *ProcedureReturn) fromTypes(p2 *types.ProcedureReturn) {
	for _, col := range p2.Fields {
		n := &NamedType{}
		n.fromTypes(col)
		p.Types = append(p.Types, n)
	}

	p.IsTable = p2.IsTable
}

func (c *DataType) fromTypes(c2 *types.DataType) {
	c.Name = c2.Name
	c.IsArray = c2.IsArray
	c.Metadata = c2.Metadata
}

func (f *ForeignProcedure) fromTypes(f2 *types.ForeignProcedure) {
	f.Name = f2.Name

	for _, param := range f2.Parameters {
		p := &DataType{}
		p.fromTypes(param)
		f.Parameters = append(f.Parameters, p)
	}

	if f2.Returns != nil {
		f.Returns = &ProcedureReturn{}
		f.Returns.fromTypes(f2.Returns)
	}
}
