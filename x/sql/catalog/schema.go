package catalog

import (
	"fmt"
	"strings"

	"kwil/x/sql/ast"
	"kwil/x/sql/sqlerr"
)

// Schema describes how the data in a relational database may relate to other tables or other data models
type Schema struct {
	Name   string
	Tables []*Table
	Types  []Type
	Funcs  []*Function

	Comment string
}

func (s *Schema) getFunc(rel *ast.FuncName, tns []*ast.TypeName) (*Function, int, error) {
	for i := range s.Funcs {
		if !strings.EqualFold(s.Funcs[i].Name, rel.Name) {
			continue
		}

		args := s.Funcs[i].InArgs()
		if len(args) != len(tns) {
			continue
		}
		found := true
		for j := range args {
			if !sameType(s.Funcs[i].Args[j].Type, tns[j]) {
				found = false
				break
			}
		}
		if !found {
			continue
		}
		return s.Funcs[i], i, nil
	}
	return nil, -1, sqlerr.RelationNotFound(rel.Name)
}

func (s *Schema) getFuncByName(rel *ast.FuncName) (*Function, int, error) {
	idx := -1
	name := strings.ToLower(rel.Name)
	for i := range s.Funcs {
		lowered := strings.ToLower(s.Funcs[i].Name)
		if lowered == name && idx >= 0 {
			return nil, -1, sqlerr.FunctionNotUnique(rel.Name)
		}
		if lowered == name {
			idx = i
		}
	}
	if idx < 0 {
		return nil, -1, sqlerr.RelationNotFound(rel.Name)
	}
	return s.Funcs[idx], idx, nil
}

func (s *Schema) GetTable(name string) (*Table, int, error) {
	for i := range s.Tables {
		if s.Tables[i].Rel.Name == name {
			return s.Tables[i], i, nil
		}
	}
	return nil, -1, sqlerr.RelationNotFound(name)
}

func (s *Schema) getType(rel *ast.TypeName) (Type, int, error) {
	for i := range s.Types {
		switch typ := s.Types[i].(type) {
		case *Enum:
			if typ.Name == rel.Name {
				return s.Types[i], i, nil
			}
		}
	}
	return nil, -1, sqlerr.TypeNotFound(rel.Name)
}

func (c *Catalog) GetSchema(name string) (*Schema, error) {
	for i := range c.Schemas {
		if c.Schemas[i].Name == name {
			return c.Schemas[i], nil
		}
	}
	return nil, sqlerr.SchemaNotFound(name)
}

func (c *Catalog) createSchemaAST(stmt *ast.CreateSchemaStmt) error {
	if stmt.Name == nil {
		return fmt.Errorf("create schema: empty name")
	}
	if _, err := c.GetSchema(*stmt.Name); err == nil {
		if !stmt.IfNotExists {
			return sqlerr.SchemaExists(*stmt.Name)
		}
	}
	c.Schemas = append(c.Schemas, &Schema{Name: *stmt.Name})
	return nil
}

func (c *Catalog) dropSchemaAST(stmt *ast.DropSchemaStmt) error {
	// TODO: n^2 in the worst-case
	for _, name := range stmt.Schemas {
		idx := -1
		for i := range c.Schemas {
			if c.Schemas[i].Name == name.Str {
				idx = i
			}
		}
		if idx == -1 {
			if stmt.MissingOk {
				continue
			}
			return sqlerr.SchemaNotFound(name.Str)
		}
		c.Schemas = append(c.Schemas[:idx], c.Schemas[idx+1:]...)
	}
	return nil
}
