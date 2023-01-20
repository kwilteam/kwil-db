package convert

import (
	"fmt"
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/databases"
)

type kwilAnyConversion struct{}

// DatabaseToAny converts a Database[anytype.KwilAny] to Database[any]
func (k *kwilAnyConversion) DatabaseToAny(d *databases.Database[anytype.KwilAny]) (*databases.Database[any], error) {
	// convert tables
	tables := make([]*databases.Table[any], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := k.TableToAny(t)
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
func (k *kwilAnyConversion) TableToAny(t *databases.Table[anytype.KwilAny]) (*databases.Table[any], error) {
	// convert columns
	columns := make([]*databases.Column[any], len(t.Columns))
	for i, c := range t.Columns {
		col, err := k.ColumnToAny(c)
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
func (k *kwilAnyConversion) ColumnToAny(c *databases.Column[anytype.KwilAny]) (*databases.Column[any], error) {
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
func (k *kwilAnyConversion) DatabaseToString(d *databases.Database[anytype.KwilAny]) (*databases.Database[string], error) {
	// convert tables
	tables := make([]*databases.Table[string], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := k.TableToString(t)
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
func (k *kwilAnyConversion) TableToString(t *databases.Table[anytype.KwilAny]) (*databases.Table[string], error) {
	// convert columns
	columns := make([]*databases.Column[string], len(t.Columns))
	for i, c := range t.Columns {
		col, err := k.ColumnToString(c)
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
func (k *kwilAnyConversion) ColumnToString(c *databases.Column[anytype.KwilAny]) (*databases.Column[string], error) {
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
func (k *kwilAnyConversion) DatabaseToBytes(d *databases.Database[anytype.KwilAny]) (*databases.Database[[]byte], error) {
	// convert tables
	tables := make([]*databases.Table[[]byte], len(d.Tables))
	for i, t := range d.Tables {
		tbl, err := k.TableToBytes(t)
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
func (k *kwilAnyConversion) TableToBytes(t *databases.Table[anytype.KwilAny]) (*databases.Table[[]byte], error) {
	// convert columns
	columns := make([]*databases.Column[[]byte], len(t.Columns))
	for i, c := range t.Columns {
		col, err := k.ColumnToBytes(c)
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
func (k *kwilAnyConversion) ColumnToBytes(c *databases.Column[anytype.KwilAny]) (*databases.Column[[]byte], error) {
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
