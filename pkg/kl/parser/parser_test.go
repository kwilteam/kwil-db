package parser_test

import (
	"kwil/pkg/kl/ast"
	"kwil/pkg/kl/parser"
	"kwil/pkg/kl/token"
	"testing"
)

func TestParser_DatabaseDeclaration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantDB      string
		wantTables  []ast.TableDefinition
		wantActions []ast.ActionDefinition
	}{
		{
			name:   "empty database",
			input:  "database test",
			wantDB: "test",
		},
		{
			name:   "empty tables",
			input:  `database test; table user{} table order{}`,
			wantDB: "test",
			wantTables: []ast.TableDefinition{
				{Name: "user"},
				{Name: "order"}},
		},
		{
			name:   "table with multiple columns and attributes",
			input:  `database demo; table user{user_id int notnull,username string null,gender bool}`,
			wantDB: "demo",
			wantTables: []ast.TableDefinition{
				{
					Name: "user",
					Columns: []ast.ColumnDefinition{
						{Name: "user_id", Type: "int", Attrs: []ast.Attribute{{AType: token.NOTNULL.ToInt()}}},
						{Name: "username", Type: "string", Attrs: []ast.Attribute{{AType: token.NULL.ToInt()}}},
						{Name: "gender", Type: "bool", Attrs: []ast.Attribute{}},
					},
				},
			},
		},
		{
			name:   "table with one column and attributes(with parameters)",
			input:  `database demo; table user{age int min(18) max(30), email string maxlen(50) minlen(10)}`,
			wantDB: "demo",
			wantTables: []ast.TableDefinition{
				{
					Name: "user",
					Columns: []ast.ColumnDefinition{
						{Name: "age", Type: "int", Attrs: []ast.Attribute{
							{AType: token.MIN.ToInt(), Value: "18"}, {AType: token.MAX.ToInt(), Value: "30"}}},
						{Name: "email", Type: "string", Attrs: []ast.Attribute{
							{AType: token.MAXLEN.ToInt(), Value: "50"}, {AType: token.MINLEN.ToInt(), Value: "10"}}},
					},
				},
			},
		},
		{
			name:   "table with index",
			input:  `database demo; table user{name string, age int, email string, uname unique(name, email), im index(email)}`,
			wantDB: "demo",
			wantTables: []ast.TableDefinition{
				{
					Name: "user",
					Columns: []ast.ColumnDefinition{
						{Name: "name", Type: "string", Attrs: []ast.Attribute{}},
						{Name: "age", Type: "int", Attrs: []ast.Attribute{}},
						{Name: "email", Type: "string", Attrs: []ast.Attribute{}},
					},
					Indexes: []ast.IndexDefinition{
						{Name: "uname", Type: "unique", Columns: []string{"name", "email"}},
						{Name: "im", Type: "index", Columns: []string{"email"}},
					},
				},
			},
		},
		{
			name: "table with action insert",
			input: `database demo;
                        table user{name string, age int, email string}
                        action create_user(name, age) public {insert into user(name, age) values (name, age)}`,
			wantDB: "demo",
			wantTables: []ast.TableDefinition{
				{
					Name: "user",
					Columns: []ast.ColumnDefinition{
						{Name: "name", Type: "string", Attrs: []ast.Attribute{}},
						{Name: "age", Type: "int", Attrs: []ast.Attribute{}},
						{Name: "email", Type: "string", Attrs: []ast.Attribute{}},
					},
				},
			},
			wantActions: []ast.ActionDefinition{
				{
					Name:   "create_user",
					Params: []string{"name", "age"},
					Public: true,
					Ops: []ast.SQLOP{
						{Op: "insert", Args: []string{"user"}},
						{Op: "columns", Args: []string{"name", "age"}},
						//{Op: "values", Args: []string{"name", "age", "a@b.com"}},
					},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			a, err := parser.Parse([]byte(c.input), parser.WithTraceOff())

			if err != nil {
				t.Errorf("Parse() got error: %s", err)
			}

			if a.Name.Name != c.wantDB {
				t.Errorf("Parse() got database name %s, want %s", a.Name, c.wantDB)
			}

			if len(a.Decls) != len(c.wantTables)+len(c.wantActions) {
				t.Errorf("Parse() got %d declarations, want %d", len(a.Decls), len(c.wantTables)+len(c.wantActions))
			}

			ti := 0
			ai := 0
			for _, decl := range a.Decls {
				switch d := decl.(type) {
				case *ast.TableDecl:
					if !testTableDeclaration(t, d, c.wantTables[ti]) {
						return
					}
					ti++
				case *ast.ActionDecl:
					if !testActionDeclaration(t, d, c.wantActions[ai]) {
						return
					}
					ai++
				}
			}
		})
	}
}

func testTableBody(t *testing.T, col *ast.ColumnDef, want ast.ColumnDefinition) bool {
	if col.Name.Name != want.Name {
		t.Errorf("columnDef.Name is not '%s'. got=%s", want.Name, col.Name.Name)
		return false
	}

	if col.Type.Name != want.Type {
		t.Errorf("columnDef.Name.Type is not '%s'. got=%s", want.Type, col.Type)
		return false
	}

	if len(col.Attrs) != len(want.Attrs) {
		t.Errorf("columnDef.Name.Attrs length is not %d. got=%d", len(want.Attrs), len(col.Attrs))
		return false
	}

	for j, attr := range col.Attrs {
		at := attr.Type.ToInt()
		if at != want.Attrs[j].AType {
			t.Errorf("columnDef.Name.Attrs[%d].Atype is not '%d'. got=%d", j, want.Attrs[j].AType, at)
			return false
		}

		if attr.Param == nil {
			continue
		}

		var v string
		switch attr.Param.(type) {
		case *ast.BasicLit:
			v = attr.Param.(*ast.BasicLit).Value
		case *ast.Ident:
			v = attr.Param.(*ast.Ident).Name
		}

		if v != want.Attrs[j].Value {
			t.Errorf("columnDef.Name.Attrs[%d].Param is not '%s'. got=%s", j, want.Attrs[j].Value, v)
			return false
		}
	}

	return true
}

func testTableIndex(t *testing.T, idx *ast.IndexDef, want ast.IndexDefinition) bool {
	if idx.Name.Name != want.Name {
		t.Errorf("indexDef.Name is not '%s'. got=%s", want.Name, idx.Name.Name)
		return false
	}

	if len(idx.Columns) != len(want.Columns) {
		t.Errorf("indexDef.Columns length is not %d. got=%d", len(want.Columns), len(idx.Columns))
		return false
	}

	for j, col := range idx.Columns {
		var name string
		switch col.(type) {
		case *ast.Ident:
			name = col.(*ast.Ident).String()
		case *ast.SelectorExpr:
			name = col.(*ast.SelectorExpr).String()
		}

		if name != want.Columns[j] {
			t.Errorf("indexDef.Columns[%d] is not '%s'. got=%s", j, want.Columns[j], name)
			return false
		}
	}

	return true
}

func testTableDeclaration(t *testing.T, d ast.Decl, want ast.TableDefinition) bool {
	tableDecl, ok := d.(*ast.TableDecl)
	if !ok {
		t.Errorf("statement is not *ast.TableDecl. got=%T", d)
		return false
	}

	if tableDecl.Name.Name != want.Name {
		t.Errorf("tableDecl.Name is not '%s'. got=%s", want.Name, tableDecl.Name.Name)
		return false
	}

	if len(tableDecl.Body) != len(want.Columns) {
		t.Errorf("tableDecl.Body length is not %d. got=%d", len(want.Columns), len(tableDecl.Body))
		return false
	}

	for i, column := range tableDecl.Body {
		col := column.(*ast.ColumnDef)
		if !testTableBody(t, col, want.Columns[i]) {
			return false
		}
	}

	for i, index := range tableDecl.Idx {
		idx := index.(*ast.IndexDef)
		if !testTableIndex(t, idx, want.Indexes[i]) {
			return false
		}
	}

	return true
}

func testInsertStatement(t *testing.T, s ast.Stmt, want []ast.SQLOP) bool {
	insertStmt, ok := s.(*ast.InsertStmt)
	if !ok {
		t.Errorf("statement is not *ast.InsertStmt. got=%T", s)
		return false
	}

	for _, w := range want {
		switch w.Op {
		case "insert":
			if insertStmt.Table.Name != w.Args[0] {
				t.Errorf("insertStmt.Table.Name is not '%s'. got=%s", w.Args[0], insertStmt.Table.Name)
				return false
			}
		case "values":
			if len(insertStmt.Values) != len(w.Args) {
				t.Errorf("insertStmt.Values length is not %d. got=%d", len(w.Args), len(insertStmt.Values))
				return false
			}

			for j, value := range insertStmt.Values {
				v := value.String()
				if v != w.Args[j] {
					t.Errorf("insertStmt.Values[%d] is not '%s'. got=%s", j, w.Args[j], v)
					return false
				}
			}

		case "columns":
			if len(insertStmt.Columns) != len(w.Args) {
				t.Errorf("insertStmt.Columns length is not %d. got=%d", len(w.Args), len(insertStmt.Columns))
				return false
			}

			for j, col := range insertStmt.Columns {
				name := col.String()
				if name != w.Args[j] {
					t.Errorf("insertStmt.Columns[%d] is not '%s'. got=%s", j, w.Args[j], name)
					return false
				}
			}
		}
	}

	return true
}

func testActionDeclaration(t *testing.T, d ast.Decl, want ast.ActionDefinition) bool {
	actionDecl, ok := d.(*ast.ActionDecl)
	if !ok {
		t.Errorf("statement is not *ast.ActionDecl. got=%T", d)
		return false
	}

	if actionDecl.Name.Name != want.Name {
		t.Errorf("actionDecl.Name is not '%s'. got=%s", want.Name, actionDecl.Name.Name)
		return false
	}

	if actionDecl.Public != want.Public {
		t.Errorf("actionDecl.Public is not '%t'. got=%t", want.Public, actionDecl.Public)
		return false
	}

	if len(actionDecl.Params) != len(want.Params) {
		t.Errorf("actionDecl.Body length is not %d. got=%d", len(want.Params), len(actionDecl.Params))
		return false
	}

	// by actionDecl.Type ?
	for _, stmt := range actionDecl.Body.Statements {
		switch st := stmt.(type) {
		case *ast.InsertStmt:
			if !testInsertStatement(t, st, want.Ops) {
				return false
			}
		}
	}

	return true
}
