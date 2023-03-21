package parser_test

import (
	"bytes"
	"kwil/pkg/engine/models"
	"kwil/pkg/engine/types"
	"kwil/pkg/kl/ast"
	"kwil/pkg/kl/parser"
	"strings"
	"testing"
)

func TestParser_DatabaseDeclaration(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantDB      string
		wantTables  []models.Table
		wantActions []models.Action
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
			wantTables: []models.Table{
				{Name: "user"},
				{Name: "order"}},
		},
		{
			name:   "table with multiple columns and attributes",
			input:  `database demo; table user{user_id int32 notnull,username string,gender bool}`,
			wantDB: "demo",
			wantTables: []models.Table{
				{
					Name: "user",
					Columns: []*models.Column{
						{Name: "user_id", Type: types.INT32, Attributes: []*models.Attribute{{Type: types.NOT_NULL}}},
						{Name: "username", Type: types.TEXT, Attributes: []*models.Attribute{}},
						{Name: "gender", Type: types.BOOLEAN, Attributes: []*models.Attribute{}},
					},
				},
			},
		},
		{
			name:   "table with one column and attributes(with parameters)",
			input:  `database demo; table user{age int32 min(18) max(30), email string maxlen(50) minlen(10)}`,
			wantDB: "demo",
			wantTables: []models.Table{
				{
					Name: "user",
					Columns: []*models.Column{
						{Name: "age", Type: types.INT32, Attributes: []*models.Attribute{
							{Type: types.MIN, Value: []byte("18")}, {Type: types.MAX, Value: []byte("30")}}},
						{Name: "email", Type: types.TEXT, Attributes: []*models.Attribute{
							{Type: types.MAX_LENGTH, Value: []byte("50")}, {Type: types.MIN_LENGTH, Value: []byte("10")}}},
					},
				},
			},
		},
		{
			name:   "table with index",
			input:  `database demo; table user{name string, age int64, email string, uname unique(name, email), im index(email)}`,
			wantDB: "demo",
			wantTables: []models.Table{
				{
					Name: "user",
					Columns: []*models.Column{
						{Name: "name", Type: types.TEXT, Attributes: []*models.Attribute{}},
						{Name: "age", Type: types.INT64, Attributes: []*models.Attribute{}},
						{Name: "email", Type: types.TEXT, Attributes: []*models.Attribute{}},
					},
					Indexes: []*models.Index{
						{Name: "uname", Type: types.UNIQUE_BTREE, Columns: []string{"name", "email"}},
						{Name: "im", Type: types.BTREE, Columns: []string{"email"}},
					},
				},
			},
		},
		{
			name: "table with action insert",
			input: `database demo;
                    table user{name string, age int64, email string}
                    action create_user(name, age) public {
insert into user (name, age) values (name, age);
insert into user (name, email) values ("test_name", "test_email@a.com");
}`,
			wantDB: "demo",
			wantTables: []models.Table{
				{
					Name: "user",
					Columns: []*models.Column{
						{Name: "name", Type: types.TEXT, Attributes: []*models.Attribute{}},
						{Name: "age", Type: types.INT64, Attributes: []*models.Attribute{}},
						{Name: "email", Type: types.TEXT, Attributes: []*models.Attribute{}},
					},
				},
			},
			wantActions: []models.Action{
				{
					Name:   "create_user",
					Inputs: []string{"name", "age"},
					Public: true,
					Statements: []string{
						"insert into user (name, age) values (name, age)",
						"insert into user (name, email) values (\"test_name\", 'test_email@a.com')"},
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
					if !testTableDeclaration(t, d, &c.wantTables[ti]) {
						return
					}
					ti++
				case *ast.ActionDecl:
					if !testActionDeclaration(t, d, &c.wantActions[ai]) {
						return
					}
					ai++
				}
			}
		})
	}
}

func testTableBody(t *testing.T, col *ast.ColumnDef, want *models.Column) bool {
	if col.Name.Name != want.Name {
		t.Errorf("columnDef.Name is not '%s'. got=%s", want.Name, col.Name.Name)
		return false
	}

	if ast.GetMappedColumnType(col.Type.Name) != want.Type {
		t.Errorf("columnDef.Name.Type is not '%s'. got=%s", want.Type, col.Type)
		return false
	}

	if len(col.Attrs) != len(want.Attributes) {
		t.Errorf("columnDef.Name.Attrs length is not %d. got=%d", len(want.Attributes), len(col.Attrs))
		return false
	}

	for j, attr := range col.Attrs {
		at := ast.GetMappedAttributeType(attr.Type)
		if at != want.Attributes[j].Type {
			t.Errorf("columnDef.Name.Attrs[%d].Atype is not '%d'. got=%d", j, want.Attributes[j].Type, at)
			return false
		}

		if attr.Param == nil {
			continue
		}

		var v string
		switch t := attr.Param.(type) {
		case *ast.BasicLit:
			v = t.Value
		case *ast.Ident:
			v = t.Name
		}

		if !bytes.Equal([]byte(v), want.Attributes[j].Value) {
			t.Errorf("columnDef.Name.Attrs[%d].Param is not '%s'. got=%s", j, want.Attributes[j].Value, v)
			return false
		}
	}

	return true
}

func testTableIndex(t *testing.T, idx *ast.IndexDef, want *models.Index) bool {
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
		switch c := col.(type) {
		case *ast.Ident:
			name = c.String()
		case *ast.SelectorExpr:
			name = c.String()
		}

		if name != want.Columns[j] {
			t.Errorf("indexDef.Columns[%d] is not '%s'. got=%s", j, want.Columns[j], name)
			return false
		}
	}

	return true
}

func testTableDeclaration(t *testing.T, d ast.Decl, want *models.Table) bool {
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

func testSQLStatement(t *testing.T, s ast.Stmt, want string) bool {
	sqlStmt, ok := s.(*ast.SQLStmt)
	if !ok {
		t.Errorf("statement is not *ast.SQLStmt. got=%T", s)
		return false
	}

	gotSql := strings.ReplaceAll(sqlStmt.SQL, " ", "")
	wantSql := strings.ReplaceAll(want, " ", "")
	if gotSql != wantSql {
		t.Errorf("sqlStmt.SQL is not '%s'. got=%s", wantSql, gotSql)
		return false
	}

	return false
}

func testActionDeclaration(t *testing.T, d ast.Decl, want *models.Action) bool {
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

	if len(actionDecl.Params) != len(want.Inputs) {
		t.Errorf("actionDecl.Body length is not %d. got=%d", len(want.Inputs), len(actionDecl.Params))
		return false
	}

	// by actionDecl.Type ?
	si := 0
	for _, stmt := range actionDecl.Body.Statements {
		switch st := stmt.(type) {
		case *ast.SQLStmt:
			if !testSQLStatement(t, st, want.Statements[si]) {
				return false
			}
			si++

			//case *ast.InsertStmt:
			//	if !testInsertStatement(t, st, want.Ops) {
			//		return false
			//	}
		}
	}

	return true
}
