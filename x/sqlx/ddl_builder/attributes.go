package ddlbuilder

import (
	"fmt"
	"reflect"
	"strings"
)

type attribute struct {
	stmt  *strings.Builder
	schma string // including these for naming purposes
	tbl   string
	col   string
}

/*
	The attribute builder has the following flow:
	1. Schema
	2. Table
	3. Constraint / Attribute

	For example:
*/

func NewAttributeBuilder() schemaPicker {
	sb := &strings.Builder{}
	sb.WriteString("ALTER TABLE ")
	return &attribute{
		stmt: sb,
	}
}

type constraintPicker interface {
	PrimaryKey(column string) builder
	Default(column string, value any) builder
	NotNull(column string) builder
	Unique(column string) builder
	Min(column string, minimum int) builder
	Max(column string, maximum int) builder
	MinLength(column string, minimum int) builder
	MaxLength(column string, maximum int) builder
}

func (c *attribute) Schema(s string) tablePicker {
	c.write(s)
	c.write(".")
	c.schma = s
	return c
}

func (c *attribute) Table(t string) constraintPicker {
	c.write(t)
	c.write(" ")
	c.tbl = t
	return c
}

// Constraints / Attributes
func (c *attribute) PrimaryKey(col string) builder {
	c.col = col
	c.write("ADD PRIMARY KEY (")
	c.write(col)
	c.write(")")
	return c
}

func (c *attribute) Default(col string, value any) builder {
	c.col = col
	c.alterColumn(col)
	c.write("SET DEFAULT ")
	c.addAny(value)
	return c
}

func (c *attribute) NotNull(col string) builder {
	c.col = col
	c.alterColumn(col)
	c.write("SET NOT NULL")
	return c
}

func (c *attribute) Unique(col string) builder {
	c.col = col
	c.addConstraint("unique")
	c.write("UNIQUE (")
	c.write(col)
	c.write(")")
	return c
}

func (c *attribute) Min(col string, value int) builder {
	c.col = col
	c.addConstraint("min")
	c.write("CHECK (")
	c.write(col)
	c.write(" >= ")
	c.write(fmt.Sprint(value))
	c.write(")")
	return c
}

func (c *attribute) Max(col string, value int) builder {
	c.col = col
	c.addConstraint("max")
	c.write("CHECK (")
	c.write(col)
	c.write(" <= ")
	c.write(fmt.Sprint(value))
	c.write(")")
	return c
}

func (c *attribute) MinLength(col string, value int) builder {
	c.col = col
	c.addConstraint("min_length")
	c.write("CHECK (LENGTH(")
	c.write(col)
	c.write(") >= ")
	c.write(fmt.Sprint(value))
	c.write(")")
	return c
}

func (c *attribute) MaxLength(col string, value int) builder {
	c.col = col
	c.addConstraint("max_length")
	c.write("CHECK (LENGTH(")
	c.write(col)
	c.write(") <= ")
	c.write(fmt.Sprint(value))
	c.write(")")
	return c
}

// Build
func (c *attribute) Build() string {
	c.write(";")
	return c.stmt.String()
}

// internal methods

// will append the value to the string builder
// if the value is a string, it will be wrapped in single quotes
func (c *attribute) addAny(a any) *attribute {
	tp := reflect.TypeOf(a).Kind()
	switch tp {
	case reflect.String:
		c.write("'")
		c.write(a.(string))
		c.write("'")
	default:
		c.write(fmt.Sprint(a))
	}
	return c
}

func (c *attribute) alterColumn(col string) *attribute {
	c.write("ALTER COLUMN ")
	c.write(col)
	c.write(" ")
	return c
}

// will add constraint and name
func (c *attribute) addConstraint(name string) *attribute {

	c.write("ADD CONSTRAINT ")
	c.write(c.generateName(name))
	c.write(" ")
	return c
}

func (c *attribute) generateName(name string) string {
	// to generate a unique name, we can use the schema, table, column, and constraint type
	// we will hash the string and use the first 63 characters

	return generateName(c.schma, c.tbl, c.col, name)
}

func (c *attribute) write(s string) *attribute {
	c.stmt.WriteString(s)
	return c
}

/*
	I have included a few example of correct Postgres syntax below.

	Adding primary key:
		ALTER TABLE table_name
  		ADD PRIMARY KEY (id);

	Adding default:
		ALTER TABLE tbl_name
		ALTER COLUMN total_cents
		SET default 0;

	Adding not null:
		ALTER TABLE tbl_name
		ALTER COLUMN total_cents
		SET not null;

	Adding unique:
		ALTER TABLE the_table
		ADD CONSTRAINT constraint_name
		UNIQUE (thecolumn);

	Adding min:
		ALTER TABLE the_table
		ADD CONSTRAINT constraint_name
		CHECK (thecolumn >= min_value);

	Adding max:
		ALTER TABLE the_table
		ADD CONSTRAINT constraint_name
		CHECK (thecolumn <= max_value);

	Adding min length:
		ALTER TABLE the_table
		ADD CONSTRAINT constraint_name
		CHECK (LENGTH(thecolumn) >= min_length);

	Adding max length:
		ALTER TABLE the_table
		ADD CONSTRAINT constraint_name
		CHECK (LENGTH(thecolumn) <= max_length);

	Adding foreign key (not in this version):
		ALTER TABLE the_table
		ADD CONSTRAINT constraint_name
		FOREIGN KEY (thecolumn)
		REFERENCES other_table(other_column);
		- optionally (ON DELETE CASCADE)
*/
