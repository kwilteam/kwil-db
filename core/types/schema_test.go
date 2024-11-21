package types

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestDataTypeBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal valid data type", func(t *testing.T) {
		original := DataType{
			Name:     "test_type",
			IsArray:  true,
			Metadata: [2]uint16{42, 123},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded DataType
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
		if decoded.IsArray != original.IsArray {
			t.Errorf("got isArray %v, want %v", decoded.IsArray, original.IsArray)
		}
		if decoded.Metadata != original.Metadata {
			t.Errorf("got metadata %v, want %v", decoded.Metadata, original.Metadata)
		}
	})

	t.Run("unmarshal with insufficient data length", func(t *testing.T) {
		data := []byte{0, 0, 0, 0}
		var dt DataType
		err := dt.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for insufficient data length")
		}
	})

	t.Run("unmarshal with invalid version", func(t *testing.T) {
		data := []byte{0, 1, 0, 0, 0, 0}
		var dt DataType
		err := dt.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid version")
		}
	})

	t.Run("unmarshal with invalid name length", func(t *testing.T) {
		data := []byte{0, 0, 255, 255, 255, 255}
		var dt DataType
		err := dt.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid name length")
		}
	})

	t.Run("marshal empty name", func(t *testing.T) {
		original := DataType{
			Name:     "",
			IsArray:  false,
			Metadata: [2]uint16{0, 0},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded DataType
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded != original {
			t.Errorf("got %v, want %v", decoded, original)
		}
	})

	t.Run("marshal with maximum metadata values", func(t *testing.T) {
		original := DataType{
			Name:     "test",
			IsArray:  true,
			Metadata: [2]uint16{65535, 65535},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded DataType
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded != original {
			t.Errorf("got %v, want %v", decoded, original)
		}
	})
}

func TestAttributeBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal valid attribute", func(t *testing.T) {
		original := Attribute{
			Type:  "test_type",
			Value: "test_value",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Attribute
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Type != original.Type {
			t.Errorf("got type %s, want %s", decoded.Type, original.Type)
		}
		if decoded.Value != original.Value {
			t.Errorf("got value %s, want %s", decoded.Value, original.Value)
		}
	})

	t.Run("marshal and unmarshal empty attribute", func(t *testing.T) {
		original := Attribute{
			Type:  "",
			Value: "",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Attribute
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded != original {
			t.Errorf("got %v, want %v", decoded, original)
		}
	})

	t.Run("unmarshal with truncated version", func(t *testing.T) {
		data := []byte{0}
		var a Attribute
		err := a.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for truncated version")
		}
	})

	t.Run("unmarshal with truncated type length", func(t *testing.T) {
		data := []byte{0, 0, 0, 0}
		var a Attribute
		err := a.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for truncated type length")
		}
	})

	t.Run("unmarshal with truncated type data", func(t *testing.T) {
		data := []byte{0, 0, 0, 0, 0, 5, 't', 'e'}
		var a Attribute
		err := a.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for truncated type data")
		}
	})

	t.Run("unmarshal with truncated value length", func(t *testing.T) {
		data := []byte{0, 0, 0, 0, 0, 1, 'x'}
		var a Attribute
		err := a.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for truncated value length")
		}
	})

	t.Run("unmarshal with truncated value data", func(t *testing.T) {
		data := []byte{0, 0, 0, 0, 0, 1, 'x', 0, 0, 0, 5, 'v'}
		var a Attribute
		err := a.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for truncated value data")
		}
	})

	t.Run("verify serialize size calculation", func(t *testing.T) {
		a := Attribute{
			Type:  "test_type",
			Value: "test_value",
		}
		expected := 2 + 4 + len(a.Type) + 4 + len(a.Value)
		if size := a.SerializeSize(); size != expected {
			t.Errorf("got size %d, want %d", size, expected)
		}
	})
}

func TestColumnBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal valid column", func(t *testing.T) {
		original := Column{
			Name: "test_column",
			Type: &DataType{
				Name:     "string",
				IsArray:  false,
				Metadata: [2]uint16{1, 2},
			},
			Attributes: []*Attribute{
				{Type: "attr1", Value: "val1"},
				{Type: "attr2", Value: "val2"},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Column
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
		if decoded.Type.Name != original.Type.Name {
			t.Errorf("got type name %s, want %s", decoded.Type.Name, original.Type.Name)
		}
		if len(decoded.Attributes) != len(original.Attributes) {
			t.Errorf("got %d attributes, want %d", len(decoded.Attributes), len(original.Attributes))
		}
	})

	t.Run("marshal and unmarshal column with no attributes", func(t *testing.T) {
		original := Column{
			Name: "empty_attrs",
			Type: &DataType{
				Name:     "int",
				IsArray:  false,
				Metadata: [2]uint16{0, 0},
			},
			Attributes: []*Attribute{},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Column
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if len(decoded.Attributes) != 0 {
			t.Errorf("got %d attributes, want 0", len(decoded.Attributes))
		}
	})

	t.Run("unmarshal with insufficient data", func(t *testing.T) {
		data := []byte{0, 0, 0, 0}
		var c Column
		err := c.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for insufficient data")
		}
	})

	t.Run("unmarshal with invalid version", func(t *testing.T) {
		data := []byte{0, 1, 0, 0, 0, 0}
		var c Column
		err := c.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid version")
		}
	})

	t.Run("unmarshal with invalid name length", func(t *testing.T) {
		data := []byte{0, 0, 255, 255, 255, 255}
		var c Column
		err := c.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid name length")
		}
	})

	t.Run("verify serialize size calculation", func(t *testing.T) {
		c := Column{
			Name: "test",
			Type: &DataType{
				Name:     "int",
				IsArray:  false,
				Metadata: [2]uint16{1, 1},
			},
			Attributes: []*Attribute{
				{Type: "attr", Value: "val"},
			},
		}

		data, err := c.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != c.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), c.SerializeSize())
		}
	})
}

func TestIndexBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal valid index", func(t *testing.T) {
		original := Index{
			Name:    "test_index",
			Columns: []string{"col1", "col2", "col3"},
			Type:    "BTREE",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Index
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
		if len(decoded.Columns) != len(original.Columns) {
			t.Errorf("got %d columns, want %d", len(decoded.Columns), len(original.Columns))
		}
		if decoded.Type != original.Type {
			t.Errorf("got type %s, want %s", decoded.Type, original.Type)
		}
	})

	t.Run("marshal and unmarshal index with empty columns", func(t *testing.T) {
		original := Index{
			Name:    "empty_cols",
			Columns: []string{},
			Type:    "HASH",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Index
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if len(decoded.Columns) != 0 {
			t.Errorf("got %d columns, want 0", len(decoded.Columns))
		}
	})

	t.Run("unmarshal with truncated columns count", func(t *testing.T) {
		data := []byte{0, 0, 0, 0, 0, 4, 't', 'e', 's', 't'}
		var idx Index
		err := idx.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for truncated columns count")
		}
	})

	t.Run("unmarshal with invalid column length", func(t *testing.T) {
		data := []byte{0, 0, 0, 0, 0, 1, 'x', 0, 0, 0, 1, 255, 255, 255, 255}
		var idx Index
		err := idx.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid column length")
		}
	})

	t.Run("verify serialize size calculation", func(t *testing.T) {
		idx := Index{
			Name:    "test",
			Columns: []string{"col1", "col2"},
			Type:    "BTREE",
		}

		data, err := idx.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != idx.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), idx.SerializeSize())
		}
	})
}

func TestForeignKeyActionBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal valid foreign key action", func(t *testing.T) {
		original := ForeignKeyAction{
			On: "UPDATE",
			Do: "CASCADE",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ForeignKeyAction
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.On != original.On {
			t.Errorf("got On %s, want %s", decoded.On, original.On)
		}
		if decoded.Do != original.Do {
			t.Errorf("got Do %s, want %s", decoded.Do, original.Do)
		}
	})

	t.Run("marshal and unmarshal with empty values", func(t *testing.T) {
		original := ForeignKeyAction{
			On: "",
			Do: "",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ForeignKeyAction
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded != original {
			t.Errorf("got %v, want %v", decoded, original)
		}
	})

	t.Run("unmarshal with truncated on length", func(t *testing.T) {
		data := []byte{0, 0, 255, 255}
		var fka ForeignKeyAction
		err := fka.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for truncated on length")
		}
	})

	t.Run("unmarshal with truncated do length", func(t *testing.T) {
		data := []byte{0, 0, 0, 0, 0, 1, 'X', 255, 255}
		var fka ForeignKeyAction
		err := fka.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for truncated do length")
		}
	})

	t.Run("verify serialize size calculation", func(t *testing.T) {
		fka := ForeignKeyAction{
			On: "DELETE",
			Do: "SET NULL",
		}

		data, err := fka.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != fka.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), fka.SerializeSize())
		}
	})

	t.Run("unmarshal with invalid version", func(t *testing.T) {
		data := []byte{0, 1, 0, 0, 0, 0}
		var fka ForeignKeyAction
		err := fka.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid version")
		}
	})
}

func TestForeignKeyBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal valid foreign key", func(t *testing.T) {
		original := ForeignKey{
			ChildKeys:   []string{"id", "type"},
			ParentKeys:  []string{"parent_id", "parent_type"},
			ParentTable: "parent_table",
			Actions: []*ForeignKeyAction{
				{On: "DELETE", Do: "CASCADE"},
				{On: "UPDATE", Do: "SET NULL"},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ForeignKey
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(decoded.ChildKeys, original.ChildKeys) {
			t.Errorf("got child keys %v, want %v", decoded.ChildKeys, original.ChildKeys)
		}
		if !reflect.DeepEqual(decoded.ParentKeys, original.ParentKeys) {
			t.Errorf("got parent keys %v, want %v", decoded.ParentKeys, original.ParentKeys)
		}
		if decoded.ParentTable != original.ParentTable {
			t.Errorf("got parent table %s, want %s", decoded.ParentTable, original.ParentTable)
		}
		if !reflect.DeepEqual(decoded.Actions, original.Actions) {
			t.Errorf("got actions %v, want %v", decoded.Actions, original.Actions)
		}
	})

	t.Run("marshal and unmarshal empty foreign key", func(t *testing.T) {
		original := ForeignKey{
			ChildKeys:   []string{},
			ParentKeys:  []string{},
			ParentTable: "",
			Actions:     []*ForeignKeyAction{},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ForeignKey
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if len(decoded.ChildKeys) != 0 {
			t.Errorf("got %d child keys, want 0", len(decoded.ChildKeys))
		}
		if len(decoded.ParentKeys) != 0 {
			t.Errorf("got %d parent keys, want 0", len(decoded.ParentKeys))
		}
		if len(decoded.Actions) != 0 {
			t.Errorf("got %d actions, want 0", len(decoded.Actions))
		}
	})

	t.Run("unmarshal with truncated child keys count", func(t *testing.T) {
		data := []byte{0, 0, 255, 255}
		var fk ForeignKey
		err := fk.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for truncated child keys count")
		}
	})

	t.Run("unmarshal with invalid child key length", func(t *testing.T) {
		data := []byte{0, 0, 0, 0, 0, 1, 255, 255, 255, 255}
		var fk ForeignKey
		err := fk.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid child key length")
		}
	})

	t.Run("unmarshal with truncated parent keys count", func(t *testing.T) {
		data := []byte{0, 0, 0, 0, 0, 0}
		var fk ForeignKey
		err := fk.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for truncated parent keys count")
		}
	})

	t.Run("verify serialize size calculation", func(t *testing.T) {
		fk := ForeignKey{
			ChildKeys:   []string{"id"},
			ParentKeys:  []string{"parent_id"},
			ParentTable: "users",
			Actions: []*ForeignKeyAction{
				{On: "DELETE", Do: "CASCADE"},
			},
		}

		data, err := fk.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != fk.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), fk.SerializeSize())
		}
	})

	t.Run("unmarshal with mismatched data length", func(t *testing.T) {
		original := ForeignKey{
			ChildKeys:   []string{"id"},
			ParentKeys:  []string{"parent_id"},
			ParentTable: "users",
			Actions:     []*ForeignKeyAction{{On: "DELETE", Do: "CASCADE"}},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		truncatedData := data[:len(data)-1]
		var decoded ForeignKey
		err = decoded.UnmarshalBinary(truncatedData)
		if err == nil {
			t.Error("expected error for mismatched data length")
		}
	})
}

func TestTableBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal complex table", func(t *testing.T) {
		original := Table{
			Name: "users",
			Columns: []*Column{
				{
					Name: "id",
					Type: &DataType{
						Name:     "int",
						IsArray:  false,
						Metadata: [2]uint16{1, 0},
					},
					Attributes: []*Attribute{
						{Type: "primary_key", Value: "true"},
					},
				},
				{
					Name: "emails",
					Type: &DataType{
						Name:     "string",
						IsArray:  true,
						Metadata: [2]uint16{255, 0},
					},
				},
			},
			Indexes: []*Index{
				{
					Name:    "email_idx",
					Columns: []string{"emails"},
					Type:    "HASH",
				},
			},
			ForeignKeys: []*ForeignKey{
				{
					ChildKeys:   []string{"department_id"},
					ParentKeys:  []string{"id"},
					ParentTable: "departments",
					Actions: []*ForeignKeyAction{
						{On: "DELETE", Do: "SET NULL"},
					},
				},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Table
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
		if len(decoded.Columns) != len(original.Columns) {
			t.Errorf("got %d columns, want %d", len(decoded.Columns), len(original.Columns))
		}
		if len(decoded.Indexes) != len(original.Indexes) {
			t.Errorf("got %d indexes, want %d", len(decoded.Indexes), len(original.Indexes))
		}
		if len(decoded.ForeignKeys) != len(original.ForeignKeys) {
			t.Errorf("got %d foreign keys, want %d", len(decoded.ForeignKeys), len(original.ForeignKeys))
		}
	})

	t.Run("marshal and unmarshal empty table", func(t *testing.T) {
		original := Table{
			Name:        "empty",
			Columns:     []*Column{},
			Indexes:     []*Index{},
			ForeignKeys: []*ForeignKey{},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Table
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
		if len(decoded.Columns) != 0 {
			t.Errorf("got %d columns, want 0", len(decoded.Columns))
		}
		if len(decoded.Indexes) != 0 {
			t.Errorf("got %d indexes, want 0", len(decoded.Indexes))
		}
		if len(decoded.ForeignKeys) != 0 {
			t.Errorf("got %d foreign keys, want 0", len(decoded.ForeignKeys))
		}
	})

	t.Run("unmarshal with insufficient version data", func(t *testing.T) {
		data := []byte{0}
		var table Table
		err := table.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for insufficient version data")
		}
	})

	t.Run("unmarshal with invalid version", func(t *testing.T) {
		data := []byte{0, 1, 0, 0, 0, 0}
		var table Table
		err := table.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid version")
		}
	})

	t.Run("verify serialize size calculation", func(t *testing.T) {
		table := Table{
			Name: "test",
			Columns: []*Column{
				{
					Name: "id",
					Type: &DataType{
						Name:     "int",
						IsArray:  false,
						Metadata: [2]uint16{1, 0},
					},
				},
			},
			Indexes: []*Index{
				{
					Name:    "idx",
					Columns: []string{"id"},
					Type:    "BTREE",
				},
			},
			ForeignKeys: []*ForeignKey{
				{
					ChildKeys:   []string{"id"},
					ParentKeys:  []string{"id"},
					ParentTable: "parent",
					Actions:     []*ForeignKeyAction{{On: "DELETE", Do: "CASCADE"}},
				},
			},
		}

		data, err := table.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != table.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), table.SerializeSize())
		}
	})

	t.Run("unmarshal with truncated name length", func(t *testing.T) {
		data := []byte{0, 0, 255, 255}
		var table Table
		err := table.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for truncated name length")
		}
	})

	t.Run("unmarshal with invalid name data", func(t *testing.T) {
		data := []byte{0, 0, 0, 0, 0, 10, 't', 'e', 's', 't'}
		var table Table
		err := table.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid name data")
		}
	})
}

func TestActionBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal complex action", func(t *testing.T) {
		original := Action{
			Name:        "ComplexAction",
			Annotations: []string{"@Deprecated", "@Beta"},
			Parameters:  []string{"param1: string", "param2: int"},
			Public:      true,
			Modifiers:   []Modifier{"async", "final"},
			Body:        "return x + y;",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Action
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
		if !reflect.DeepEqual(decoded.Annotations, original.Annotations) {
			t.Errorf("got annotations %v, want %v", decoded.Annotations, original.Annotations)
		}
		if !reflect.DeepEqual(decoded.Parameters, original.Parameters) {
			t.Errorf("got parameters %v, want %v", decoded.Parameters, original.Parameters)
		}
		if decoded.Public != original.Public {
			t.Errorf("got public %v, want %v", decoded.Public, original.Public)
		}
		if !reflect.DeepEqual(decoded.Modifiers, original.Modifiers) {
			t.Errorf("got modifiers %v, want %v", decoded.Modifiers, original.Modifiers)
		}
		if decoded.Body != original.Body {
			t.Errorf("got body %s, want %s", decoded.Body, original.Body)
		}
	})

	t.Run("marshal and unmarshal minimal action", func(t *testing.T) {
		original := Action{
			Name:        "MinimalAction",
			Annotations: []string{},
			Parameters:  []string{},
			Public:      false,
			Modifiers:   []Modifier{},
			Body:        "",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Action
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(decoded, original) {
			t.Errorf("got %v, want %v", decoded, original)
		}
	})

	t.Run("verify serialize size with unicode characters", func(t *testing.T) {
		action := Action{
			Name:        "测试",
			Annotations: []string{"🔥"},
			Parameters:  []string{"param1: 字符串"},
			Public:      true,
			Modifiers:   []Modifier{"异步"},
			Body:        "返回;",
		}

		data, err := action.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != action.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), action.SerializeSize())
		}
	})

	t.Run("unmarshal with corrupted annotation length", func(t *testing.T) {
		original := Action{
			Name:        "Test",
			Annotations: []string{"test"},
		}
		data, _ := original.MarshalBinary()

		// Corrupt annotation length
		offset := 6 + len(original.Name)
		binary.BigEndian.PutUint32(data[offset:], uint32(255))

		var decoded Action
		err := decoded.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for corrupted annotation length")
		}
	})

	t.Run("unmarshal with corrupted parameter length", func(t *testing.T) {
		original := Action{
			Name:       "Test",
			Parameters: []string{"param"},
		}
		data, _ := original.MarshalBinary()

		// Corrupt parameter count
		offset := 10 + len(original.Name)
		binary.BigEndian.PutUint32(data[offset:], uint32(255))

		var decoded Action
		err := decoded.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for corrupted parameter length")
		}
	})

	t.Run("unmarshal with corrupted annotations length", func(t *testing.T) {
		original := Action{
			Name:        "Test",
			Annotations: []string{"test"},
		}
		data, _ := original.MarshalBinary()

		// Corrupt modifier count
		offset := 11 + len(original.Name)
		binary.BigEndian.PutUint32(data[offset:], uint32(255))

		var decoded Action
		err := decoded.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for corrupted modifier length")
		}
	})
}

func TestProcedureParameterBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal valid procedure parameter", func(t *testing.T) {
		original := ProcedureParameter{
			Name: "test_param",
			Type: &DataType{
				Name:     "varchar",
				IsArray:  false,
				Metadata: [2]uint16{100, 0},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ProcedureParameter
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
		if decoded.Type.Name != original.Type.Name {
			t.Errorf("got type name %s, want %s", decoded.Type.Name, original.Type.Name)
		}
		if decoded.Type.IsArray != original.Type.IsArray {
			t.Errorf("got type isArray %v, want %v", decoded.Type.IsArray, original.Type.IsArray)
		}
		if decoded.Type.Metadata != original.Type.Metadata {
			t.Errorf("got type metadata %v, want %v", decoded.Type.Metadata, original.Type.Metadata)
		}
	})

	t.Run("marshal and unmarshal with unicode name", func(t *testing.T) {
		original := ProcedureParameter{
			Name: "测试参数",
			Type: &DataType{
				Name:     "text",
				IsArray:  false,
				Metadata: [2]uint16{0, 0},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ProcedureParameter
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
	})

	t.Run("unmarshal with truncated type data", func(t *testing.T) {
		original := ProcedureParameter{
			Name: "test",
			Type: &DataType{
				Name:     "int",
				IsArray:  false,
				Metadata: [2]uint16{0, 0},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		truncatedData := data[:len(data)-1]
		var decoded ProcedureParameter
		err = decoded.UnmarshalBinary(truncatedData)
		if err == nil {
			t.Error("expected error for truncated type data")
		}
	})

	t.Run("verify serialize size calculation", func(t *testing.T) {
		param := ProcedureParameter{
			Name: "param",
			Type: &DataType{
				Name:     "decimal",
				IsArray:  true,
				Metadata: [2]uint16{10, 2},
			},
		}

		data, err := param.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != param.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), param.SerializeSize())
		}
	})

	t.Run("unmarshal with empty name", func(t *testing.T) {
		original := ProcedureParameter{
			Name: "",
			Type: &DataType{
				Name:     "bool",
				IsArray:  false,
				Metadata: [2]uint16{0, 0},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ProcedureParameter
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != "" {
			t.Errorf("got name %s, want empty string", decoded.Name)
		}
	})
}

func TestNamedTypeBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal named type", func(t *testing.T) {
		original := NamedType{
			Name: "CustomType",
			Type: &DataType{
				Name:     "int",
				IsArray:  true,
				Metadata: [2]uint16{8, 0},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded NamedType
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
		if decoded.Type.Name != original.Type.Name {
			t.Errorf("got type name %s, want %s", decoded.Type.Name, original.Type.Name)
		}
		if decoded.Type.IsArray != original.Type.IsArray {
			t.Errorf("got isArray %v, want %v", decoded.Type.IsArray, original.Type.IsArray)
		}
		if decoded.Type.Metadata != original.Type.Metadata {
			t.Errorf("got metadata %v, want %v", decoded.Type.Metadata, original.Type.Metadata)
		}
	})

	t.Run("verify serialize size matches actual size", func(t *testing.T) {
		nt := NamedType{
			Name: "MyType",
			Type: &DataType{
				Name:     "varchar",
				IsArray:  false,
				Metadata: [2]uint16{255, 0},
			},
		}

		data, err := nt.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != nt.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), nt.SerializeSize())
		}
	})

	t.Run("marshal and unmarshal with empty name", func(t *testing.T) {
		original := NamedType{
			Name: "",
			Type: &DataType{
				Name:     "bool",
				IsArray:  false,
				Metadata: [2]uint16{0, 0},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded NamedType
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != "" {
			t.Errorf("got name %s, want empty string", decoded.Name)
		}
	})

	t.Run("unmarshal with nil type", func(t *testing.T) {
		original := NamedType{
			Name: "Test",
			Type: nil,
		}

		_, err := original.MarshalBinary()
		if err == nil {
			t.Error("expected error for nil type")
		}
	})

	t.Run("unmarshal with invalid data length", func(t *testing.T) {
		data := []byte{0, 0}
		var nt NamedType
		err := nt.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid data length")
		}
	})
}

func TestProcedureReturnBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal complex procedure return", func(t *testing.T) {
		original := ProcedureReturn{
			IsTable: true,
			Fields: []*NamedType{
				{
					Name: "id",
					Type: &DataType{
						Name:     "int",
						IsArray:  false,
						Metadata: [2]uint16{4, 0},
					},
				},
				{
					Name: "names",
					Type: &DataType{
						Name:     "varchar",
						IsArray:  true,
						Metadata: [2]uint16{100, 0},
					},
				},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ProcedureReturn
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.IsTable != original.IsTable {
			t.Errorf("got isTable %v, want %v", decoded.IsTable, original.IsTable)
		}
		if len(decoded.Fields) != len(original.Fields) {
			t.Errorf("got %d fields, want %d", len(decoded.Fields), len(original.Fields))
		}
		for i, field := range decoded.Fields {
			if field.Name != original.Fields[i].Name {
				t.Errorf("field %d: got name %s, want %s", i, field.Name, original.Fields[i].Name)
			}
			if field.Type.Name != original.Fields[i].Type.Name {
				t.Errorf("field %d: got type %s, want %s", i, field.Type.Name, original.Fields[i].Type.Name)
			}
		}
	})

	t.Run("marshal and unmarshal scalar return", func(t *testing.T) {
		original := ProcedureReturn{
			IsTable: false,
			Fields: []*NamedType{
				{
					Name: "result",
					Type: &DataType{
						Name:     "boolean",
						IsArray:  false,
						Metadata: [2]uint16{0, 0},
					},
				},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ProcedureReturn
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.IsTable {
			t.Error("got isTable true, want false")
		}
		if len(decoded.Fields) != 1 {
			t.Errorf("got %d fields, want 1", len(decoded.Fields))
		}
	})

	t.Run("marshal and unmarshal empty fields", func(t *testing.T) {
		original := ProcedureReturn{
			IsTable: false,
			Fields:  []*NamedType{},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ProcedureReturn
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if len(decoded.Fields) != 0 {
			t.Errorf("got %d fields, want 0", len(decoded.Fields))
		}
	})

	t.Run("unmarshal with insufficient data", func(t *testing.T) {
		data := []byte{0, 0, 1}
		var pr ProcedureReturn
		err := pr.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for insufficient data")
		}
	})

	t.Run("unmarshal with invalid version", func(t *testing.T) {
		data := []byte{0, 1, 0, 0, 0, 0, 0}
		var pr ProcedureReturn
		err := pr.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid version")
		}
	})

	t.Run("verify serialize size calculation", func(t *testing.T) {
		pr := ProcedureReturn{
			IsTable: true,
			Fields: []*NamedType{
				{
					Name: "field1",
					Type: &DataType{
						Name:     "int",
						IsArray:  false,
						Metadata: [2]uint16{0, 0},
					},
				},
			},
		}

		data, err := pr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != pr.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), pr.SerializeSize())
		}
	})
}

func TestProcedureReturnBinaryMarshalingExtended(t *testing.T) {
	t.Run("marshal and unmarshal with max fields", func(t *testing.T) {
		original := ProcedureReturn{
			IsTable: true,
			Fields: []*NamedType{
				{
					Name: "field1",
					Type: &DataType{
						Name:     "int",
						IsArray:  false,
						Metadata: [2]uint16{0, 0},
					},
				},
				{
					Name: "field2",
					Type: &DataType{
						Name:     "varchar",
						IsArray:  true,
						Metadata: [2]uint16{255, 0},
					},
				},
				{
					Name: "field3",
					Type: &DataType{
						Name:     "decimal",
						IsArray:  false,
						Metadata: [2]uint16{10, 2},
					},
				},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ProcedureReturn
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if len(decoded.Fields) != len(original.Fields) {
			t.Errorf("got %d fields, want %d", len(decoded.Fields), len(original.Fields))
		}
		for i, field := range decoded.Fields {
			if !reflect.DeepEqual(field, original.Fields[i]) {
				t.Errorf("field %d mismatch: got %v, want %v", i, field, original.Fields[i])
			}
		}
	})

	t.Run("unmarshal with corrupted field data", func(t *testing.T) {
		original := ProcedureReturn{
			IsTable: true,
			Fields: []*NamedType{
				{
					Name: "test",
					Type: &DataType{
						Name:     "int",
						IsArray:  false,
						Metadata: [2]uint16{0, 0},
					},
				},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		// Corrupt the field data
		data = data[:len(data)-1]

		var decoded ProcedureReturn
		err = decoded.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for corrupted field data")
		}
	})

	t.Run("marshal with nil field type", func(t *testing.T) {
		pr := ProcedureReturn{
			IsTable: false,
			Fields: []*NamedType{
				{
					Name: "test",
					Type: nil,
				},
			},
		}

		_, err := pr.MarshalBinary()
		if err == nil {
			t.Error("expected error for nil field type")
		}
	})

	t.Run("verify size calculation with unicode field names", func(t *testing.T) {
		pr := ProcedureReturn{
			IsTable: true,
			Fields: []*NamedType{
				{
					Name: "测试字段",
					Type: &DataType{
						Name:     "text",
						IsArray:  false,
						Metadata: [2]uint16{0, 0},
					},
				},
			},
		}

		data, err := pr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != pr.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), pr.SerializeSize())
		}
	})
}

func TestExtensionConfigBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal with special characters", func(t *testing.T) {
		original := ExtensionConfig{
			Key:   "config.🔑",
			Value: "value.⚡",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ExtensionConfig
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Key != original.Key {
			t.Errorf("got key %s, want %s", decoded.Key, original.Key)
		}
		if decoded.Value != original.Value {
			t.Errorf("got value %s, want %s", decoded.Value, original.Value)
		}
	})

	t.Run("marshal and unmarshal with very long strings", func(t *testing.T) {
		key := strings.Repeat("k", 1000)
		value := strings.Repeat("v", 1000)
		original := ExtensionConfig{
			Key:   key,
			Value: value,
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ExtensionConfig
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Key != original.Key {
			t.Errorf("key length mismatch: got %d, want %d", len(decoded.Key), len(original.Key))
		}
		if decoded.Value != original.Value {
			t.Errorf("value length mismatch: got %d, want %d", len(decoded.Value), len(original.Value))
		}
	})

	t.Run("unmarshal with truncated value length", func(t *testing.T) {
		original := ExtensionConfig{
			Key:   "test",
			Value: "value",
		}
		data, _ := original.MarshalBinary()
		truncatedData := data[:len(data)-2]

		var decoded ExtensionConfig
		err := decoded.UnmarshalBinary(truncatedData)
		if err == nil {
			t.Error("expected error for truncated value length")
		}
	})

	t.Run("verify serialize size calculation", func(t *testing.T) {
		configs := []ExtensionConfig{
			{Key: "", Value: ""},
			{Key: "a", Value: "b"},
			{Key: "key", Value: strings.Repeat("v", 100)},
		}

		for _, config := range configs {
			data, err := config.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}

			if len(data) != config.SerializeSize() {
				t.Errorf("size mismatch for config %v: got %d, want %d",
					config, len(data), config.SerializeSize())
			}
		}
	})

	t.Run("unmarshal with invalid key length", func(t *testing.T) {
		data := []byte{0, 0, 255, 255, 255, 255}
		var config ExtensionConfig
		err := config.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid key length")
		}
	})

	t.Run("unmarshal with invalid value length", func(t *testing.T) {
		data := []byte{0, 0, 0, 0, 0, 1, 'x', 255, 255, 255, 255}
		var config ExtensionConfig
		err := config.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid value length")
		}
	})
}

func TestExtensionBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal complex extension", func(t *testing.T) {
		original := Extension{
			Name: "test_extension",
			Initialization: []*ExtensionConfig{
				{Key: "key1", Value: "value1"},
				{Key: "key2", Value: "value2"},
				{Key: "key3", Value: "value3"},
			},
			Alias: "test_alias",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Extension
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
		if decoded.Alias != original.Alias {
			t.Errorf("got alias %s, want %s", decoded.Alias, original.Alias)
		}
		if len(decoded.Initialization) != len(original.Initialization) {
			t.Errorf("got %d configs, want %d", len(decoded.Initialization), len(original.Initialization))
		}
		for i, config := range decoded.Initialization {
			if config.Key != original.Initialization[i].Key {
				t.Errorf("config %d: got key %s, want %s", i, config.Key, original.Initialization[i].Key)
			}
			if config.Value != original.Initialization[i].Value {
				t.Errorf("config %d: got value %s, want %s", i, config.Value, original.Initialization[i].Value)
			}
		}
	})

	t.Run("marshal and unmarshal empty extension", func(t *testing.T) {
		original := Extension{
			Name:           "",
			Initialization: []*ExtensionConfig{},
			Alias:          "",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Extension
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != "" {
			t.Error("expected empty name")
		}
		if len(decoded.Initialization) != 0 {
			t.Error("expected empty initialization")
		}
		if decoded.Alias != "" {
			t.Error("expected empty alias")
		}
	})

	t.Run("unmarshal with invalid initialization data", func(t *testing.T) {
		original := Extension{
			Name: "test",
			Initialization: []*ExtensionConfig{
				{Key: "key", Value: "value"},
			},
			Alias: "alias",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		// Corrupt the initialization data
		corrupted := make([]byte, len(data))
		copy(corrupted, data)
		offset := 6 + len(original.Name) + 4
		corrupted[offset] = 255

		var decoded Extension
		err = decoded.UnmarshalBinary(corrupted)
		if err == nil {
			t.Error("expected error for corrupted initialization data")
		}
	})

	t.Run("verify serialize size with large initialization", func(t *testing.T) {
		configs := make([]*ExtensionConfig, 100)
		for i := range configs {
			configs[i] = &ExtensionConfig{
				Key:   fmt.Sprintf("key%d", i),
				Value: fmt.Sprintf("value%d", i),
			}
		}

		ext := Extension{
			Name:           "large_test",
			Initialization: configs,
			Alias:          "large_alias",
		}

		data, err := ext.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != ext.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), ext.SerializeSize())
		}
	})

	t.Run("unmarshal with truncated alias length", func(t *testing.T) {
		original := Extension{
			Name:           "test",
			Initialization: []*ExtensionConfig{},
			Alias:          "alias",
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		truncated := data[:len(data)-2]
		var decoded Extension
		err = decoded.UnmarshalBinary(truncated)
		if err == nil {
			t.Error("expected error for truncated alias length")
		}
	})

	t.Run("unmarshal with invalid version", func(t *testing.T) {
		data := []byte{0, 1, 0, 0, 0, 0}
		var ext Extension
		err := ext.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid version")
		}
	})
}

func TestForeignProcedureBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal complex foreign procedure", func(t *testing.T) {
		original := ForeignProcedure{
			Name: "complex_proc",
			Parameters: []*DataType{
				{
					Name:     "param1",
					IsArray:  true,
					Metadata: [2]uint16{10, 2},
				},
				{
					Name:     "param2",
					IsArray:  false,
					Metadata: [2]uint16{0, 0},
				},
			},
			Returns: &ProcedureReturn{
				IsTable: true,
				Fields: []*NamedType{
					{
						Name: "result",
						Type: &DataType{
							Name:     "int",
							IsArray:  false,
							Metadata: [2]uint16{4, 0},
						},
					},
				},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ForeignProcedure
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
		if len(decoded.Parameters) != len(original.Parameters) {
			t.Errorf("got %d parameters, want %d", len(decoded.Parameters), len(original.Parameters))
		}
		if !reflect.DeepEqual(decoded.Returns, original.Returns) {
			t.Errorf("got returns %v, want %v", decoded.Returns, original.Returns)
		}
	})

	t.Run("marshal and unmarshal with no parameters", func(t *testing.T) {
		original := ForeignProcedure{
			Name:       "no_params",
			Parameters: []*DataType{},
			Returns: &ProcedureReturn{
				IsTable: false,
				Fields: []*NamedType{
					{
						Name: "result",
						Type: &DataType{
							Name:     "bool",
							IsArray:  false,
							Metadata: [2]uint16{0, 0},
						},
					},
				},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ForeignProcedure
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if len(decoded.Parameters) != 0 {
			t.Errorf("got %d parameters, want 0", len(decoded.Parameters))
		}
	})

	t.Run("unmarshal with invalid parameter data", func(t *testing.T) {
		original := ForeignProcedure{
			Name: "test",
			Parameters: []*DataType{
				{
					Name:     "param",
					IsArray:  false,
					Metadata: [2]uint16{0, 0},
				},
			},
			Returns: &ProcedureReturn{
				IsTable: false,
				Fields: []*NamedType{
					{
						Name: "result",
						Type: &DataType{
							Name:     "int",
							IsArray:  false,
							Metadata: [2]uint16{0, 0},
						},
					},
				},
			},
		}

		data, _ := original.MarshalBinary()
		corrupted := make([]byte, len(data))
		copy(corrupted, data)
		offset := 10 + len(original.Name)
		corrupted[offset] = 255

		var decoded ForeignProcedure
		err := decoded.UnmarshalBinary(corrupted)
		if err == nil {
			t.Error("expected error for corrupted parameter data")
		}
	})

	t.Run("verify serialize size calculation", func(t *testing.T) {
		fp := ForeignProcedure{
			Name: "size_test",
			Parameters: []*DataType{
				{
					Name:     "param",
					IsArray:  true,
					Metadata: [2]uint16{1, 1},
				},
			},
			Returns: &ProcedureReturn{
				IsTable: true,
				Fields: []*NamedType{
					{
						Name: "result",
						Type: &DataType{
							Name:     "varchar",
							IsArray:  false,
							Metadata: [2]uint16{100, 0},
						},
					},
				},
			},
		}

		data, err := fp.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != fp.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), fp.SerializeSize())
		}
	})

	t.Run("unmarshal with nil returns", func(t *testing.T) {
		fp := ForeignProcedure{
			Name: "test",
			Parameters: []*DataType{
				{
					Name:     "param",
					IsArray:  false,
					Metadata: [2]uint16{0, 0},
				},
			},
			Returns: nil,
		}

		data, err := fp.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded ForeignProcedure
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("unmarshal with invalid version", func(t *testing.T) {
		data := []byte{0, 1, 0, 0, 0, 0}
		var fp ForeignProcedure
		err := fp.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid version")
		}
	})
}

func TestSchemaBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal complex schema", func(t *testing.T) {
		original := Schema{
			Name:  "test_schema",
			Owner: []byte{0x1, 0x2, 0x3},
			Extensions: []*Extension{
				{
					Name: "ext1",
					Initialization: []*ExtensionConfig{
						{Key: "k1", Value: "v1"},
					},
					Alias: "e1",
				},
			},
			Tables: []*Table{
				{
					Name: "table1",
					Columns: []*Column{
						{
							Name: "col1",
							Type: &DataType{
								Name:     "int",
								IsArray:  false,
								Metadata: [2]uint16{4, 0},
							},
						},
					},
				},
			},
			Actions: []*Action{
				{
					Name:        "action1",
					Annotations: []string{"@test"},
					Parameters:  []string{"p1"},
					Public:      true,
					Body:        "body",
				},
			},
			Procedures: []*Procedure{
				{
					Name: "proc1",
					Parameters: []*ProcedureParameter{
						{
							Name: "param1",
							Type: &DataType{
								Name:     "text",
								IsArray:  false,
								Metadata: [2]uint16{0, 0},
							},
						},
					},
				},
			},
			ForeignProcedures: []*ForeignProcedure{
				{
					Name: "fp1",
					Parameters: []*DataType{
						{
							Name:     "param1",
							IsArray:  false,
							Metadata: [2]uint16{0, 0},
						},
					},
					Returns: &ProcedureReturn{
						IsTable: false,
						Fields: []*NamedType{
							{
								Name: "result",
								Type: &DataType{
									Name:     "bool",
									IsArray:  false,
									Metadata: [2]uint16{0, 0},
								},
							},
						},
					},
				},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Schema
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != original.Name {
			t.Errorf("got name %s, want %s", decoded.Name, original.Name)
		}
		if !reflect.DeepEqual(decoded.Owner, original.Owner) {
			t.Errorf("got owner %v, want %v", decoded.Owner, original.Owner)
		}
		if len(decoded.Extensions) != len(original.Extensions) {
			t.Errorf("got %d extensions, want %d", len(decoded.Extensions), len(original.Extensions))
		}
		if len(decoded.Tables) != len(original.Tables) {
			t.Errorf("got %d tables, want %d", len(decoded.Tables), len(original.Tables))
		}
		if len(decoded.Actions) != len(original.Actions) {
			t.Errorf("got %d actions, want %d", len(decoded.Actions), len(original.Actions))
		}
		if len(decoded.Procedures) != len(original.Procedures) {
			t.Errorf("got %d procedures, want %d", len(decoded.Procedures), len(original.Procedures))
		}
		if len(decoded.ForeignProcedures) != len(original.ForeignProcedures) {
			t.Errorf("got %d foreign procedures, want %d", len(decoded.ForeignProcedures), len(original.ForeignProcedures))
		}
	})

	t.Run("marshal and unmarshal empty schema", func(t *testing.T) {
		original := Schema{
			Name:              "",
			Owner:             []byte{},
			Extensions:        []*Extension{},
			Tables:            []*Table{},
			Actions:           []*Action{},
			Procedures:        []*Procedure{},
			ForeignProcedures: []*ForeignProcedure{},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Schema
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Name != "" {
			t.Error("expected empty name")
		}
		if len(decoded.Owner) != 0 {
			t.Error("expected empty owner")
		}
		if len(decoded.Extensions) != 0 {
			t.Error("expected empty extensions")
		}
		if len(decoded.Tables) != 0 {
			t.Error("expected empty tables")
		}
		if len(decoded.Actions) != 0 {
			t.Error("expected empty actions")
		}
		if len(decoded.Procedures) != 0 {
			t.Error("expected empty procedures")
		}
		if len(decoded.ForeignProcedures) != 0 {
			t.Error("expected empty foreign procedures")
		}
	})

	t.Run("unmarshal with insufficient data", func(t *testing.T) {
		data := []byte{0, 0}
		var s Schema
		err := s.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for insufficient data")
		}
	})

	t.Run("unmarshal with invalid version", func(t *testing.T) {
		data := []byte{0, 1, 0, 0, 0, 0}
		var s Schema
		err := s.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid version")
		}
	})

	t.Run("unmarshal with invalid name length", func(t *testing.T) {
		data := []byte{0, 0, 255, 255, 255, 255}
		var s Schema
		err := s.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid name length")
		}
	})

	t.Run("verify serialize size calculation", func(t *testing.T) {
		schema := Schema{
			Name:  "test",
			Owner: []byte{0x1},
			Extensions: []*Extension{
				{
					Name:  "ext",
					Alias: "e",
				},
			},
			Tables: []*Table{
				{
					Name: "table",
				},
			},
			Actions: []*Action{
				{
					Name: "action",
					Body: "body",
				},
			},
			Procedures: []*Procedure{
				{
					Name: "proc",
				},
			},
			ForeignProcedures: []*ForeignProcedure{
				{
					Name: "fp",
					Returns: &ProcedureReturn{
						IsTable: false,
						Fields: []*NamedType{
							{
								Name: "result",
								Type: &DataType{
									Name: "void",
								},
							},
						},
					},
				},
			},
		}

		data, err := schema.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != schema.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), schema.SerializeSize())
		}
	})
}

func TestProcedureBinaryMarshaling(t *testing.T) {
	t.Run("marshal and unmarshal with all fields populated", func(t *testing.T) {
		original := Procedure{
			Name: "test_procedure",
			Parameters: []*ProcedureParameter{
				{
					Name: "param1",
					Type: &DataType{
						Name:     "int",
						IsArray:  false,
						Metadata: [2]uint16{4, 0},
					},
				},
				{
					Name: "param2",
					Type: &DataType{
						Name:     "varchar",
						IsArray:  true,
						Metadata: [2]uint16{255, 0},
					},
				},
			},
			Public: true,
			Modifiers: []Modifier{
				"IMMUTABLE",
				"STRICT",
			},
			Body: "SELECT * FROM table WHERE id = $1",
			Returns: &ProcedureReturn{
				IsTable: true,
				Fields: []*NamedType{
					{
						Name: "id",
						Type: &DataType{
							Name:     "int",
							IsArray:  false,
							Metadata: [2]uint16{4, 0},
						},
					},
				},
			},
			Annotations: []string{
				"@deprecated",
				"@returns(int)",
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Procedure
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(decoded, original) {
			t.Errorf("decoded procedure does not match original")
		}
	})

	t.Run("marshal and unmarshal with no returns", func(t *testing.T) {
		original := Procedure{
			Name: "void_proc",
			Parameters: []*ProcedureParameter{
				{
					Name: "param",
					Type: &DataType{
						Name:     "text",
						IsArray:  false,
						Metadata: [2]uint16{0, 0},
					},
				},
			},
			Public:      true,
			Modifiers:   []Modifier{"VOLATILE"},
			Body:        "INSERT INTO logs(message) VALUES ($1)",
			Returns:     nil,
			Annotations: []string{},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		var decoded Procedure
		err = decoded.UnmarshalBinary(data)
		if err != nil {
			t.Fatal(err)
		}

		if decoded.Returns != nil {
			t.Error("expected nil returns")
		}
		if !reflect.DeepEqual(decoded, original) {
			t.Errorf("decoded procedure does not match original")
		}
	})

	t.Run("unmarshal with truncated modifier length", func(t *testing.T) {
		original := Procedure{
			Name:      "test",
			Public:    false,
			Modifiers: []Modifier{"TEST"},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		truncatedData := data[:len(data)-2]
		var decoded Procedure
		err = decoded.UnmarshalBinary(truncatedData)
		if err == nil {
			t.Error("expected error for truncated modifier length")
		}
	})

	t.Run("verify serialize size with unicode characters", func(t *testing.T) {
		proc := Procedure{
			Name: "测试过程",
			Parameters: []*ProcedureParameter{
				{
					Name: "参数",
					Type: &DataType{
						Name:     "text",
						IsArray:  false,
						Metadata: [2]uint16{0, 0},
					},
				},
			},
			Public:      true,
			Modifiers:   []Modifier{"异步"},
			Body:        "返回 TRUE;",
			Returns:     nil,
			Annotations: []string{"@测试"},
		}

		data, err := proc.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		if len(data) != proc.SerializeSize() {
			t.Errorf("got size %d, want %d", len(data), proc.SerializeSize())
		}
	})

	t.Run("unmarshal with invalid parameter data", func(t *testing.T) {
		data := []byte{
			0, 0, // version
			0, 0, 0, 4, 't', 'e', 's', 't', // name
			0, 0, 0, 1, // parameter count
			255, 255, // invalid parameter data
		}
		var proc Procedure
		err := proc.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid parameter data")
		}
	})

	t.Run("unmarshal with invalid return data", func(t *testing.T) {
		original := Procedure{
			Name: "test",
			Returns: &ProcedureReturn{
				IsTable: false,
				Fields: []*NamedType{
					{
						Name: "result",
						Type: &DataType{
							Name:     "int",
							IsArray:  false,
							Metadata: [2]uint16{0, 0},
						},
					},
				},
			},
		}

		data, err := original.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}

		// Corrupt return data
		offset := 6 + len(original.Name) + 4 + 1 + 4 + 4
		data[offset] = 255

		var decoded Procedure
		err = decoded.UnmarshalBinary(data)
		if err == nil {
			t.Error("expected error for invalid return data")
		}
	})
}
