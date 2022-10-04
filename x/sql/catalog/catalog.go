package catalog

import (
	"fmt"
	"kwil/x/sql/sqlerr"
	"kwil/x/sql/sqlparse/ast"
	"strings"
)

// Catalog describes a database instance consisting of metadata in which database objects are defined
type Catalog struct {
	Comment       string
	DefaultSchema string
	Name          string
	Schemas       []*Schema
	SearchPath    []string
	LoadExtension func(string) *Schema

	extensions map[string]struct{}
}

// New creates a new catalog
func New(defaultSchema string) *Catalog {
	newCatalog := &Catalog{
		DefaultSchema: defaultSchema,
		Schemas:       make([]*Schema, 0),
		extensions:    make(map[string]struct{}),
	}

	if newCatalog.DefaultSchema != "" {
		newCatalog.Schemas = append(newCatalog.Schemas, &Schema{Name: defaultSchema})
	}

	return newCatalog
}

func (c *Catalog) Table(schemaName, tableName string) (*Table, bool) {
	if schema, ok := c.Schema(schemaName); ok {
		return schema.Table(tableName)
	}
	return nil, false
}

func (c *Catalog) Schema(schemaName string) (*Schema, bool) {
	if schemaName == "" {
		schemaName = c.DefaultSchema
	}

	for _, schema := range c.Schemas {
		if schema.Name == schemaName {
			return schema, true
		}
	}
	return nil, false
}

func (s *Schema) Table(tableName string) (*Table, bool) {
	for _, table := range s.Tables {
		if table.QualName.Name == tableName {
			return table, true
		}
	}
	return nil, false
}

func (c *Catalog) schemasToSearch(ns string) []string {
	if ns == "" {
		ns = c.DefaultSchema
	}
	return append(c.SearchPath, ns)
}

func (c *Catalog) Funcs(schemaName, funcName string) ([]Function, error) {
	var funcs []Function
	lowered := strings.ToLower(funcName)
	for _, ns := range c.schemasToSearch(schemaName) {
		s, ok := c.Schema(ns)
		if !ok {
			return nil, fmt.Errorf("schema %s not found", ns)
		}
		for i := range s.Funcs {
			if strings.ToLower(s.Funcs[i].Name) == lowered {
				funcs = append(funcs, *s.Funcs[i])
			}
		}
	}
	return funcs, nil
}
func (c *Catalog) Func(schemaName, funcName string, tns ...*QualName) (*Function, bool) {
	ns := schemaName
	if ns == "" {
		ns = c.DefaultSchema
	}
	s, ok := c.Schema(ns)
	if !ok {
		return nil, false
	}
	return s.Func(funcName, tns...)
}

func (c *Catalog) ResolveFuncCall(call *ast.FuncCall) (*Function, error) {
	// Do not validate unknown functions
	funs, err := c.Funcs(call.Func.Schema, call.Func.Name)
	if err != nil || len(funs) == 0 {
		return nil, sqlerr.FunctionNotFound(call.Func.Name)
	}

	// https://www.postgresql.org/docs/current/sql-syntax-calling-funcs.html
	var positional []ast.Node
	var named []*ast.NamedArgExpr

	if call.Args != nil {
		for _, arg := range call.Args.Items {
			if narg, ok := arg.(*ast.NamedArgExpr); ok {
				named = append(named, narg)
			} else {
				// The mixed notation combines positional and named notation.
				// However, as already mentioned, named arguments cannot precede
				// positional arguments.
				if len(named) > 0 {
					return nil, &sqlerr.Error{
						Code:     "",
						Message:  "positional argument cannot follow named argument",
						Location: call.Pos(),
					}
				}
				positional = append(positional, arg)
			}
		}
	}

	for _, fun := range funs {
		args := fun.InArgs()
		var defaults int
		var variadic bool
		known := map[string]struct{}{}
		for _, arg := range args {
			if arg.HasDefault {
				defaults += 1
			}
			if arg.Mode == FuncParamVariadic {
				variadic = true
				defaults += 1
			}
			if arg.Name != "" {
				known[arg.Name] = struct{}{}
			}
		}

		if variadic {
			if (len(named) + len(positional)) < (len(args) - defaults) {
				continue
			}
		} else {
			if (len(named) + len(positional)) > len(args) {
				continue
			}
			if (len(named) + len(positional)) < (len(args) - defaults) {
				continue
			}
		}

		// Validate that the provided named arguments exist in the function
		var unknownArgName bool
		for _, expr := range named {
			if expr.Name != nil {
				if _, found := known[*expr.Name]; !found {
					unknownArgName = true
				}
			}
		}
		if unknownArgName {
			continue
		}

		return &fun, nil
	}

	var sig []string
	for range call.Args.Items {
		sig = append(sig, "unknown")
	}

	return nil, &sqlerr.Error{
		Code:     "42883",
		Message:  fmt.Sprintf("function %s(%s) does not exist", call.Func.Name, strings.Join(sig, ", ")),
		Location: call.Pos(),
	}
}

func (s *Schema) Func(name string, tns ...*QualName) (*Function, bool) {
	for i := range s.Funcs {
		if !strings.EqualFold(s.Funcs[i].Name, name) {
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
		return s.Funcs[i], true
	}
	return nil, false
}
