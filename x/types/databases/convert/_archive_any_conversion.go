package convert

import (
	"fmt"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
)

/*databases. this file contains methods to convert a database's any type
I will provide conversions for:
- any <---> KwilAny ---> string

Since Golang generics fucking suck and don't (always) work with methods,
the conversion functions will be held in a "Convert" namespace.

The Convert namespace will have 3 sub-namespaces:
- KwilAny
- Any
- Bytes

The "KwilAny" namespace is meant to be an intermediary namespace for conversions.

For example, in order to convert any to bytes, we need to convert any to KwilAny first.

Supported conversions:
Namespace: KwilAny
- anytype.KwilAny ---> any
- anytype.KwilAny ---> string
- anytype.KwilAny ---> bytes

Namespace: Any
- any ---> anytype.KwilAny
- any ---> string
- any ---> anytype.KwilAny ---> bytes

Namespace: Bytes
- bytes ---> anytype.KwilAny
- bytes ---> anytype.KwilAny ---> string
- bytes ---> anytype.KwilAny ---> any

We do not provide a string namespace, since string -> anytype.KwilAny is would result in all values being strings.

Most of this code is extremely repetitive and written by CoPilot.
*/

type conversions struct {
	Any     anyConversion
	KwilAny kwilAnyConversion
	Bytes   bytesConversion
}

type anyConversion struct{}

type kwilAnyConversion struct{}

type bytesConversion struct{}

var (
	Convert = conversions{
		Any:     anyConversion{},
		KwilAny: kwilAnyConversion{},
		Bytes:   bytesConversion{},
	}
)

/*databases.
####################################################################################################
"AnyType" namespace conversions
####################################################################################################
*/

// anytype.KwilAny ---> any

// DatabaseToAny converts a Database[anytype.KwilAny] to Database[any]
func (kwilAnyConversion) DatabaseToAny(d *databases.Database[anytype.KwilAny]) (*databases.Database[any], error) {
	// convert tables
	tables := make([]*databases.Table[any], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := Convert.KwilAny.TableToAny(t)
		if err != nil {
			return nil, err
		}
		tables[i] = tbl
	}

	return &Database[any]{
		Owner:      d.Owner,
		Name:       d.Name,
		Tables:     tables,
		Roles:      d.Roles,
		SQLQueries: d.SQLQueries,
		Indexes:    d.Indexes,
	}, nil
}

// TableToAny converts a Table[anytype.KwilAny] to Table[any]
func (kwilAnyConversion) TableToAny(t *databases.Table[anytype.KwilAny]) (*databases.Table[any], error) {
	// convert columns
	columns := make([]*databases.Column[any], len(t.Columns))
	for i, c := range t.Columns {
		col, err := Convert.KwilAny.ColumnToAny(c)
		if err != nil {
			return nil, err
		}
		columns[i] = col
	}

	return &Table[any]{
		Name:    t.Name,
		Columns: columns,
	}, nil
}

// ColumnToAny converts a Column[anytype.KwilAny] to Column[any]
func (kwilAnyConversion) ColumnToAny(c *databases.Column[anytype.KwilAny]) (*databases.Column[any], error) {
	// convert attributes
	attributes := make([]*databases.Attribute[any], len(c.Attributes))
	for i, a := range c.Attributes {
		value, err := a.Value.Deserialize()
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize attribute value: %w", err)
		}

		attributes[i] = &Attribute[any]{
			Type:  a.Type,
			Value: value,
		}
	}

	return &Column[any]{
		Name:       c.Name,
		Type:       c.Type,
		Attributes: attributes,
	}, nil
}

// anytype.KwilAny ---> string

// DatabaseToString converts a Database[anytype.KwilAny] to Database[string]
func (kwilAnyConversion) DatabaseToString(d *databases.Database[anytype.KwilAny]) (*databases.Database[string], error) {
	// convert tables
	tables := make([]*databases.Table[string], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := Convert.KwilAny.TableToString(t)
		if err != nil {
			return nil, err
		}
		tables[i] = tbl
	}

	return &Database[string]{
		Owner:      d.Owner,
		Name:       d.Name,
		Tables:     tables,
		Roles:      d.Roles,
		SQLQueries: d.SQLQueries,
		Indexes:    d.Indexes,
	}, nil
}

// TableToString converts a Table[anytype.KwilAny] to Table[string]
func (kwilAnyConversion) TableToString(t *databases.Table[anytype.KwilAny]) (*databases.Table[string], error) {
	// convert columns
	columns := make([]*databases.Column[string], len(t.Columns))
	for i, c := range t.Columns {
		col, err := Convert.KwilAny.ColumnToString(c)
		if err != nil {
			return nil, err
		}
		columns[i] = col
	}

	return &Table[string]{
		Name:    t.Name,
		Columns: columns,
	}, nil
}

// ColumnToString converts a Column[anytype.KwilAny] to Column[string]
func (kwilAnyConversion) ColumnToString(c *databases.Column[anytype.KwilAny]) (*databases.Column[string], error) {
	// convert attributes
	attributes := make([]*databases.Attribute[string], len(c.Attributes))
	for i, a := range c.Attributes {
		value, err := a.Value.String()
		if err != nil {
			return nil, fmt.Errorf("failed to deserialize attribute value and convert to string: %w", err)
		}

		attributes[i] = &Attribute[string]{
			Type:  a.Type,
			Value: value,
		}
	}

	return &Column[string]{
		Name:       c.Name,
		Type:       c.Type,
		Attributes: attributes,
	}, nil
}

// anytype.KwilAny ---> []byte

// DatabaseToBytes converts a Database[anytype.KwilAny] to Database[[]byte]
func (kwilAnyConversion) DatabaseToBytes(d *databases.Database[anytype.KwilAny]) (*databases.Database[[]byte], error) {
	// convert tables
	tables := make([]*databases.Table[[]byte], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := Convert.KwilAny.TableToBytes(t)
		if err != nil {
			return nil, err
		}
		tables[i] = tbl
	}

	return &Database[[]byte]{
		Owner:      d.Owner,
		Name:       d.Name,
		Tables:     tables,
		Roles:      d.Roles,
		SQLQueries: d.SQLQueries,
		Indexes:    d.Indexes,
	}, nil
}

// TableToBytes converts a Table[anytype.KwilAny] to Table[[]byte]
func (kwilAnyConversion) TableToBytes(t *databases.Table[anytype.KwilAny]) (*databases.Table[[]byte], error) {
	// convert columns
	columns := make([]*databases.Column[[]byte], len(t.Columns))
	for i, c := range t.Columns {
		col, err := Convert.KwilAny.ColumnToBytes(c)
		if err != nil {
			return nil, err
		}
		columns[i] = col
	}

	return &Table[[]byte]{
		Name:    t.Name,
		Columns: columns,
	}, nil
}

// ColumnToBytes converts a Column[anytype.KwilAny] to Column[[]byte]
func (kwilAnyConversion) ColumnToBytes(c *databases.Column[anytype.KwilAny]) (*databases.Column[[]byte], error) {
	// convert attributes
	attributes := make([]*databases.Attribute[[]byte], len(c.Attributes))
	for i, a := range c.Attributes {
		value, err := a.Value.Serialize()
		if err != nil {
			return nil, fmt.Errorf("failed to serialize attribute value: %w", err)
		}

		attributes[i] = &Attribute[[]byte]{
			Type:  a.Type,
			Value: value,
		}
	}

	return &Column[[]byte]{
		Name:       c.Name,
		Type:       c.Type,
		Attributes: attributes,
	}, nil
}

/*
	####################################################################################################
	"Any" namespace conversions
	####################################################################################################
*/

// any ---> anytype.KwilAny

// ToAnyType converts the database's values T type to anytype.KwilAny
func (anyConversion) DatabaseToAnyType(d *databases.Database[any]) (*databases.Database[anytype.KwilAny], error) {
	// convert tables
	tables := make([]*databases.Table[anytype.KwilAny], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := Convert.Any.TableToAnyType(t)
		if err != nil {
			return nil, err
		}
		tables[i] = tbl
	}

	return &Database[anytype.KwilAny]{
		Owner:      d.Owner,
		Name:       d.Name,
		Tables:     tables,
		Roles:      d.Roles,
		SQLQueries: d.SQLQueries,
		Indexes:    d.Indexes,
	}, nil
}

func (anyConversion) TableToAnyType(t *databases.Table[any]) (*databases.Table[anytype.KwilAny], error) {
	// convert columns
	columns := make([]*databases.Column[anytype.KwilAny], len(t.Columns))
	for i, c := range t.Columns {
		col, err := Convert.Any.ColumnToAnyType(c)
		if err != nil {
			return nil, err
		}
		columns[i] = col
	}

	return &Table[anytype.KwilAny]{
		Name:    t.Name,
		Columns: columns,
	}, nil
}

func (anyConversion) ColumnToAnyType(c *databases.Column[any]) (*databases.Column[anytype.KwilAny], error) {
	// convert attributes
	attributes := make([]*databases.Attribute[anytype.KwilAny], len(c.Attributes))
	for i, a := range c.Attributes {
		value, err := anytype.New(a)
		if err != nil {
			return nil, err
		}

		attributes[i] = &Attribute[anytype.KwilAny]{
			Type:  a.Type,
			Value: value,
		}
	}

	return &Column[anytype.KwilAny]{
		Name:       c.Name,
		Type:       c.Type,
		Attributes: attributes,
	}, nil
}

// any ---> string

// DatabaseToString converts a Database[any] to Database[string]
func (anyConversion) DatabaseToString(d *databases.Database[any]) (*databases.Database[string], error) {
	// convert tables
	tables := make([]*databases.Table[string], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := Convert.Any.TableToString(t)
		if err != nil {
			return nil, err
		}
		tables[i] = tbl
	}

	return &Database[string]{
		Owner:      d.Owner,
		Name:       d.Name,
		Tables:     tables,
		Roles:      d.Roles,
		SQLQueries: d.SQLQueries,
		Indexes:    d.Indexes,
	}, nil
}

// TableToString converts a Table[any] to Table[string]
func (anyConversion) TableToString(t *databases.Table[any]) (*databases.Table[string], error) {
	// convert columns
	columns := make([]*databases.Column[string], len(t.Columns))
	for i, c := range t.Columns {
		col, err := Convert.Any.ColumnToString(c)
		if err != nil {
			return nil, err
		}
		columns[i] = col
	}

	return &Table[string]{
		Name:    t.Name,
		Columns: columns,
	}, nil
}

// ColumnToString converts a Column[any] to Column[string]
func (anyConversion) ColumnToString(c *databases.Column[any]) (*databases.Column[string], error) {
	// convert attributes
	attributes := make([]*databases.Attribute[string], len(c.Attributes))
	for i, a := range c.Attributes {
		attributes[i] = &Attribute[string]{
			Type:  a.Type,
			Value: fmt.Sprintf("%v", a.Value),
		}
	}

	return &Column[string]{
		Name:       c.Name,
		Type:       c.Type,
		Attributes: attributes,
	}, nil
}

// any ---> anytype.KwilAny ---> []byte

// DatabaseToBytes converts a Database[any] to Database[[]byte]
func (anyConversion) DatabaseToBytes(d *databases.Database[any]) (*databases.Database[[]byte], error) {
	// convert tables
	tables := make([]*databases.Table[[]byte], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := Convert.Any.TableToBytes(t)
		if err != nil {
			return nil, err
		}
		tables[i] = tbl
	}

	return &Database[[]byte]{
		Owner:      d.Owner,
		Name:       d.Name,
		Tables:     tables,
		Roles:      d.Roles,
		SQLQueries: d.SQLQueries,
		Indexes:    d.Indexes,
	}, nil
}

// TableToBytes converts a Table[any] to Table[[]byte]
func (anyConversion) TableToBytes(t *databases.Table[any]) (*databases.Table[[]byte], error) {
	// convert columns
	columns := make([]*databases.Column[[]byte], len(t.Columns))
	for i, c := range t.Columns {
		col, err := Convert.Any.ColumnToBytes(c)
		if err != nil {
			return nil, err
		}
		columns[i] = col
	}

	return &Table[[]byte]{
		Name:    t.Name,
		Columns: columns,
	}, nil
}

// ColumnToBytes converts a Column[any] to Column[[]byte]
func (anyConversion) ColumnToBytes(c *databases.Column[any]) (*databases.Column[[]byte], error) {
	// convert attributes
	attributes := make([]*databases.Attribute[[]byte], len(c.Attributes))
	for i, a := range c.Attributes {
		kwilAny, err := anytype.New(a)
		if err != nil {
			return nil, fmt.Errorf("failed to convert attribute value to kwilAny: %w", err)
		}

		bytes, err := kwilAny.Bytes()
		if err != nil {
			return nil, fmt.Errorf("failed to convert kwilAny to bytes: %w", err)
		}

		attributes[i] = &Attribute[[]byte]{
			Type:  a.Type,
			Value: bytes,
		}
	}

	return &Column[[]byte]{
		Name:       c.Name,
		Type:       c.Type,
		Attributes: attributes,
	}, nil
}

/*
	####################################################################################################
	"Bytes" namespace conversions
	####################################################################################################
*/

// bytes ---> anytype.KwilAny

// DatabaseToKwilAny converts a Database[[]byte] to Database[anytype.KwilAny]
func (bytesConversion) DatabaseToKwilAny(d *databases.Database[[]byte]) (*databases.Database[anytype.KwilAny], error) {
	// convert tables
	tables := make([]*databases.Table[anytype.KwilAny], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := Convert.Bytes.TableToKwilAny(t)
		if err != nil {
			return nil, err
		}
		tables[i] = tbl
	}

	return &Database[anytype.KwilAny]{
		Owner:      d.Owner,
		Name:       d.Name,
		Tables:     tables,
		Roles:      d.Roles,
		SQLQueries: d.SQLQueries,
		Indexes:    d.Indexes,
	}, nil
}

// TableToKwilAny converts a Table[[]byte] to Table[anytype.KwilAny]
func (bytesConversion) TableToKwilAny(t *databases.Table[[]byte]) (*databases.Table[anytype.KwilAny], error) {
	// convert columns
	columns := make([]*databases.Column[anytype.KwilAny], len(t.Columns))
	for i, c := range t.Columns {
		col, err := Convert.Bytes.ColumnToKwilAny(c)
		if err != nil {
			return nil, err
		}
		columns[i] = col
	}

	return &Table[anytype.KwilAny]{
		Name:    t.Name,
		Columns: columns,
	}, nil
}

// ColumnToKwilAny converts a Column[[]byte] to Column[anytype.KwilAny]
func (bytesConversion) ColumnToKwilAny(c *databases.Column[[]byte]) (*databases.Column[anytype.KwilAny], error) {
	// convert attributes
	attributes := make([]*databases.Attribute[anytype.KwilAny], len(c.Attributes))
	for i, a := range c.Attributes {
		kwilAny, err := anytype.New(a)
		if err != nil {
			return nil, fmt.Errorf("failed to convert attribute value to kwilAny: %w", err)
		}

		attributes[i] = &Attribute[anytype.KwilAny]{
			Type:  a.Type,
			Value: kwilAny,
		}
	}

	return &Column[anytype.KwilAny]{
		Name:       c.Name,
		Type:       c.Type,
		Attributes: attributes,
	}, nil
}

// bytes ---> anytype.KwilAny ---> string

// DatabaseToString converts a Database[[]byte] to Database[string]
func (bytesConversion) DatabaseToString(d *databases.Database[[]byte]) (*databases.Database[string], error) {
	// convert tables
	tables := make([]*databases.Table[string], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := Convert.Bytes.TableToString(t)
		if err != nil {
			return nil, err
		}
		tables[i] = tbl
	}

	return &Database[string]{
		Owner:      d.Owner,
		Name:       d.Name,
		Tables:     tables,
		Roles:      d.Roles,
		SQLQueries: d.SQLQueries,
		Indexes:    d.Indexes,
	}, nil
}

// TableToString converts a Table[[]byte] to Table[string]
func (bytesConversion) TableToString(t *databases.Table[[]byte]) (*databases.Table[string], error) {
	// convert columns
	columns := make([]*databases.Column[string], len(t.Columns))
	for i, c := range t.Columns {
		col, err := Convert.Bytes.ColumnToString(c)
		if err != nil {
			return nil, err
		}
		columns[i] = col
	}

	return &Table[string]{
		Name:    t.Name,
		Columns: columns,
	}, nil
}

// ColumnToString converts a Column[[]byte] to Column[string]
func (bytesConversion) ColumnToString(c *databases.Column[[]byte]) (*databases.Column[string], error) {
	// convert attributes
	attributes := make([]*databases.Attribute[string], len(c.Attributes))
	for i, a := range c.Attributes {
		kwilAny, err := anytype.New(a)
		if err != nil {
			return nil, fmt.Errorf("failed to convert attribute value to kwilAny: %w", err)
		}

		str, err := kwilAny.String()
		if err != nil {
			return nil, fmt.Errorf("failed to convert kwilAny to string: %w", err)
		}

		attributes[i] = &Attribute[string]{
			Type:  a.Type,
			Value: str,
		}
	}

	return &Column[string]{
		Name:       c.Name,
		Type:       c.Type,
		Attributes: attributes,
	}, nil
}
