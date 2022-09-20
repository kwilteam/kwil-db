package dbml

import (
	"strings"
	"testing"
)

func p(str string) *Parser {
	r := strings.NewReader(str)
	s := NewScanner(r)
	parser := NewParser(s)
	return parser
}

func TestIllegalSyntax(t *testing.T) {
	parser := p(`Project test { abc , xyz`)
	_, err := parser.Parse()
	if err == nil {
		t.Fail()
	}
}

func TestParseSimple(t *testing.T) {
	parser := p(`
	Project test {
		note: 'just test note'
	}
	table users {
		id int [pk, note: 'just test column note']
	}
	table float_number {

	}
	`)
	dbml, err := parser.Parse()
	if err != nil {
		t.Fail()
	}
	if dbml.Project.Name != "test" {
		t.Fail()
	}

	if dbml.Project.Note != "just test note" {
		t.Fail()
	}

	usersTable := dbml.Tables[0]
	if usersTable.Name != "users" {
		t.Fail()
	}
	idColumn := usersTable.Columns[0]
	if idColumn.Name != "id" {
		t.Fail()
	}
	if !idColumn.Settings.PK {
		t.Fail()
	}
	if idColumn.Settings.Note != "just test column note" {
		t.Fail()
	}
}

func TestParseTableName(t *testing.T) {
	parser := p(`
	Table int {
		id int
	}
	`)
	dbml, err := parser.Parse()
	if err != nil {
		t.Fail()
	}
	table := dbml.Tables[0]
	if table.Name != "int" {
		t.Fatalf("table name should be 'int'")
	}
}

func TestParseTableWithType(t *testing.T) {
	parser := p(`
	Table int {
		type int
	}
	`)
	dbml, err := parser.Parse()
	if err != nil {
		t.Fail()
	}
	table := dbml.Tables[0]
	if table.Columns[0].Name != "type" {
		t.Fatalf("column name should be 'type'")
	}
}

func TestParseTableWithNoteColumn(t *testing.T) {
	parser := p(`
	Table int {
		note int
	}
	`)
	dbml, err := parser.Parse()

	//t.Log(err)
	if err != nil {
		t.Fatalf("%v", err)
	}

	table := dbml.Tables[0]
	if table.Columns[0].Name != "note" {
		t.Fatalf("column name should be 'note'")
	}
}

func TestAllowKeywordsAsTable(t *testing.T) {
	parser := p(`
	Table project {
		note int
	}
	`)
	dbml, err := parser.Parse()

	//t.Log(err)
	if err != nil {
		t.Fatalf("%v", err)
	}

	table := dbml.Tables[0]
	if table.Name != "project" {
		t.Fatalf("table name should be 'project'")
	}
}

func TestAllowKeywordsAsEnum(t *testing.T) {
	parser := p(`
	Enum project {
		key
	}
	`)
	dbml, err := parser.Parse()

	//t.Log(err)
	if err != nil {
		t.Fatalf("%v", err)
	}

	enum := dbml.Enums[0]
	if enum.Name != "project" {
		t.Fatalf("enum name should be 'project'")
	}

	if enum.Values[0].Name != "key" {
		t.Fatalf("enum value should be 'key'")
	}
}

func TestQueryParse(t *testing.T) {
	markup := "query get_user_by_id: `\n\tselect * from users where id = $1\n`"

	dbml, err := ParseString(markup)
	if err != nil {
		t.Fail()
	}
	_ = dbml
}

func TestStuff(t *testing.T) {
	dbml, err := ParseFile("/Users/bryan/Downloads/dbml-go-master/test.dbml")
	if err != nil {
		t.Fail()
	}
	_ = dbml
}

func TestParserStuff(t *testing.T) {
	markup := `
	table ecommerce.merchants {
		id int
		country_code int
		merchant_name varchar

		"created at" varchar
		admin_id int [ref: > U.id]
		indexes {
			(id, country_code) [pk]
		}
	}

	table users as U {
		id int [pk, increment]
		full_name varchar
		created_at timestamp
		country_code int
	}

	table countries {
		code int [pk]
		name varchar(1024)
		continent_name varchar
	}

	table ecommerce.order_items {
		order_id int [ref: > ecommerce.orders.id]
		product_id int
		quantity int [default: 1]
	}

	table ecommerce.orders {
		id int [pk]
		user_id int [not null, unique]
		status varchar
		created_at varchar [note: 'When order created']
	}

	enum ecommerce.products_status {
		out_of_stock
		in_stock
		running_low [note: 'less than 20']
	}

	table ecommerce.products {
		id int [pk]
		name varchar
		merchant_id int [not null]
		price int
		status ecommerce.products_status
		created_at datetime [default: 'now()']

		indexes {
			(merchant_id, status) [name:'product_status']
			id [unique]
		}
	}

	table ecommerce.product_tags {
		id int [pk]
		name varchar(1024)
	}

	table ecommerce.merchant_periods {
		id int [pk]
		merchant_id int
		country_code int
		start_date datetime
		end_date datetime
	}

	ref: U.country_code > countries.code
	ref: ecommerce.merchants.country_code > countries.code
	ref: ecommerce.order_items.product_id > ecommerce.products.id
	ref: ecommerce.products.merchant_id > ecommerce.merchants.id
	ref: ecommerce.product_tags.id <> ecommerce.products.id
	ref: ecommerce.merchant_periods.(merchant_id, country_code) > ecommerce.merchants.(id, country_code)
	`

	dbml, err := ParseString(markup)
	if err != nil {
		t.Fail()
	}
	_ = dbml
}
