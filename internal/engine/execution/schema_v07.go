package execution

import (
	"encoding/json"

	"github.com/kwilteam/kwil-db/core/types"
)

// this file contains the schema from Kwil v0.7, which is used for proper migration to
// the new schema in v0.8

// convertV07Schema converts a v0.7 schema to a v0.8 schema
func convertV07Schema(bts []byte) (*types.Schema, error) {
	s := &v07Schema{}
	err := json.Unmarshal(bts, s)
	if err != nil {
		return nil, err
	}

	tables := make([]*types.Table, len(s.Tables))
	for i, t := range s.Tables {
		columns := make([]*types.Column, len(t.Columns))
		for j, c := range t.Columns {
			attrs := make([]*types.Attribute, len(c.Attributes))
			for k, a := range c.Attributes {
				attrs[k] = &types.Attribute{
					Type:  types.AttributeType(a.Type),
					Value: a.Value,
				}
			}
			columns[j] = &types.Column{
				Name: c.Name,
				Type: &types.DataType{
					Name: string(c.Type),
				},
				Attributes: attrs,
			}
		}

		indexes := make([]*types.Index, len(t.Indexes))
		for j, idx := range t.Indexes {
			indexes[j] = &types.Index{
				Name:    idx.Name,
				Columns: idx.Columns,
				Type:    types.IndexType(idx.Type),
			}
		}

		foreignKeys := make([]*types.ForeignKey, len(t.ForeignKeys))
		for j, fk := range t.ForeignKeys {
			fkActions := make([]*types.ForeignKeyAction, len(fk.Actions))
			for k, fkAct := range fk.Actions {
				fkActions[k] = &types.ForeignKeyAction{
					On: types.ForeignKeyActionOn(fkAct.On),
					Do: types.ForeignKeyActionDo(fkAct.Do),
				}
			}
			foreignKeys[j] = &types.ForeignKey{
				ChildKeys:   fk.ChildKeys,
				ParentKeys:  fk.ParentKeys,
				ParentTable: fk.ParentTable,
				Actions:     fkActions,
			}
		}

		tables[i] = &types.Table{
			Name:        t.Name,
			Columns:     columns,
			Indexes:     indexes,
			ForeignKeys: foreignKeys,
		}
	}

	extensions := make([]*types.Extension, len(s.Extensions))
	for i, e := range s.Extensions {
		init := make([]*types.ExtensionConfig, len(e.Initialization))
		for j, i := range e.Initialization {
			init[j] = &types.ExtensionConfig{
				Key:   i.Key,
				Value: i.Value,
			}
		}
		extensions[i] = &types.Extension{
			Name:           e.Name,
			Initialization: init,
			Alias:          e.Alias,
		}
	}

	actions := make([]*types.Action, len(s.Procedures))
	for i, p := range s.Procedures {
		actions[i] = &types.Action{
			Name:        p.Name,
			Annotations: p.Annotations,
			Parameters:  p.Args,
			Public:      p.Public,
		}

		for _, m := range p.Modifiers {
			actions[i].Modifiers = append(actions[i].Modifiers, types.Modifier(m))
		}

		var body string
		for _, s := range p.Statements {
			body += s + "\n"
		}
		actions[i].Body = body
	}

	return &types.Schema{
		Name:       s.Name,
		Owner:      s.Owner,
		Extensions: extensions,
		Tables:     tables,
		Actions:    actions,
	}, nil

}

// v07Schema is a database schema that contains tables, procedures, and extensions.
type v07Schema struct {
	// Name is the name of the schema given by the deployer.
	Name string `json:"name"`
	// Owner is the identifier (generally an address in bytes or public key) of the owner of the schema
	Owner      []byte          `json:"owner"`
	Extensions []*v07Extension `json:"extensions"`
	Tables     []*v07Table     `json:"tables"`
	Procedures []*v07Procedure `json:"procedures"`
}

// v07Table is a table in a database schema.
type v07Table struct {
	Name        string           `json:"name"`
	Columns     []*v07Column     `json:"columns"`
	Indexes     []*v07Index      `json:"indexes,omitempty"`
	ForeignKeys []*v07ForeignKey `json:"foreign_keys"`
}

// v07Column is a column in a table.
type v07Column struct {
	Name       string          `json:"name"`
	Type       v07DataType     `json:"type"`
	Attributes []*v07Attribute `json:"attributes,omitempty"`
}

// v07Attribute is a column attribute.
// These are constraints and default values.
type v07Attribute struct {
	Type  string `json:"type"`
	Value string `json:"value,omitempty"`
}

// v07Index is an index on a table.
type v07Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Type    string   `json:"type"`
}

// v07ForeignKey is a foreign key in a table.
type v07ForeignKey struct {
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
	Actions []*v07ForeignKeyAction `json:"actions"`
}

// v07ForeignKeyAction is used to specify what should occur
// if a parent key is updated or deleted
type v07ForeignKeyAction struct {
	// On can be either "UPDATE" or "DELETE"
	On string `json:"on"`

	// Do specifies what a foreign key action should do
	Do string `json:"do"`
}

// v07Extension defines what extensions the schema uses, and how they are initialized.
type v07Extension struct {
	// Name is the name of the extension registered in the node
	Name string `json:"name"`
	// Initialization is a list of key value pairs that are used to initialize the extension
	Initialization []*v07ExtensionConfig `json:"initialization"`
	// Alias is the alias of the extension, which is how its instance is referred to in the schema
	Alias string `json:"alias"`
}

// v07ExtensionConfig is a key value pair that represents a configuration value for an extension
type v07ExtensionConfig struct {
	Key   string `json:"name"`
	Value string `json:"value"`
}

// v07DataType is a type of data (e.g. NULL, TEXT, INT, BLOB, BOOLEAN)
type v07DataType string

// v07Procedure is a procedure in a database schema.
// These are defined by Kuneiform's `action` keyword.
type v07Procedure struct {
	Name        string   `json:"name"`
	Annotations []string `json:"annotations,omitempty"`
	Args        []string `json:"inputs"`
	Public      bool     `json:"public"`
	Modifiers   []string `json:"modifiers"`
	Statements  []string `json:"statements"`
}
