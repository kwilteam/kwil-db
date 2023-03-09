package parser_test

import (
	"kwil/pkg/kl/ast"
	"kwil/pkg/kl/parser"
	"kwil/pkg/kl/token"
	"testing"
)

func TestParser_DatabaseDeclaration(t *testing.T) {
	cases := []struct {
		input      string
		wantDB     string
		wantTables []ast.TableDefinition
	}{
		{
			input:  `database test{table user{} table order{}}`,
			wantDB: "test",
			wantTables: []ast.TableDefinition{
				{Name: "user"},
				{Name: "order"}},
		},
		{
			input:  `database demo{table user{user_id int notnull,username string null,gender bool}}`,
			wantDB: "demo",
			wantTables: []ast.TableDefinition{
				{
					Name: "user",
					Columns: []ast.ColumnDefinition{
						{Name: "user_id", Type: "int", Attrs: []ast.AttributeType{{AType: token.NOTNULL.ToInt(), Value: token.NOTNULL.String()}}},
						{Name: "username", Type: "string", Attrs: []ast.AttributeType{{AType: token.NULL.ToInt(), Value: token.NULL.String()}}},
						{Name: "gender", Type: "bool", Attrs: []ast.AttributeType{}},
					},
				},
			},
		},
	}

	for _, c := range cases {
		a, err := parser.Parse([]byte(c.input), parser.WithTraceOff())

		if err != nil {
			t.Errorf("Parse() got error: %s", err)
		}

		if len(a.Statements) != 1 {
			t.Errorf("Parse() got %d statements, want 1", len(a.Statements))
		}

		if !testDatabaseDeclaration(t, a.Statements[0], c.wantDB, c.wantTables) {
			return
		}
	}
}

func testDatabaseDeclaration(t *testing.T, s ast.Stmt, wantDB string, wantTables []ast.TableDefinition) bool {
	databaseDecl, ok := s.(*ast.DatabaseDecl)
	if !ok {
		t.Errorf("statement is not *ast.DatabaseDecl. got=%T", s)
		return false
	}

	if databaseDecl.Name.Value != wantDB {
		t.Errorf("databaseDecl.Name is not '%s'. got=%s", wantDB, databaseDecl.Name)
		return false
	}

	if len(databaseDecl.Body.Statements) != len(wantTables) {
		t.Errorf("databaseDecl.Tables is not 1. got=%d", len(databaseDecl.Body.Statements))
		return false
	}

	for i, table := range databaseDecl.Body.Statements {
		if !testTableDeclaration(t, table, wantTables[i]) {
			return false
		}
	}

	return true
}

func testTableDeclaration(t *testing.T, s ast.Stmt, want ast.TableDefinition) bool {
	tableDecl, ok := s.(*ast.TableDecl)
	if !ok {
		t.Errorf("statement is not *ast.TableDecl. got=%T", s)
		return false
	}

	if tableDecl.Name.Value != want.Name {
		t.Errorf("tableDecl.Name is not '%s'. got=%s", want.Name, tableDecl.Name.Value)
		return false
	}

	if len(tableDecl.Body) != len(want.Columns) {
		t.Errorf("tableDecl.Body length is not %d. got=%d", len(want.Columns), len(tableDecl.Body))
		return false
	}

	for i, col := range tableDecl.Body {
		if col.Name.Value != want.Columns[i].Name {
			t.Errorf("tableDecl.Body[%d].Name is not '%s'. got=%s", i, want.Columns[i].Name, col.Name.Value)
			return false
		}

		if col.Type.Value != want.Columns[i].Type {
			t.Errorf("tableDecl.Body[%d].Type is not '%s'. got=%s", i, want.Columns[i].Type, col.Type)
			return false
		}

		if len(col.Attrs) != len(want.Columns[i].Attrs) {
			t.Errorf("tableDecl.Body[%d].Attrs length is not %d. got=%d", i, len(want.Columns[i].Attrs), len(col.Attrs))
			return false
		}

		for j, attr := range col.Attrs {
			if attr.Type.Token.Literal != want.Columns[i].Attrs[j].Value {
				t.Errorf("tableDecl.Body[%d].Attrs[%d] is not '%s'. got=%s", i, j, want.Columns[i].Attrs[j].Value, attr.Type.Token.Literal)
				return false
			}
		}
	}

	return true
}
