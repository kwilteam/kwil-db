package driver_test

import (
	"fmt"
	"kwil/pkg/sql/driver"
	"math/big"
	"testing"
)

func createTest3DB() (*driver.Connection, error) {
	conn, err := driver.OpenConn("kwil_test_3")
	if err != nil {
		return nil, err
	}

	err = conn.AcquireLock()
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func Test_Plan2(t *testing.T) {
	db, err := createTestDB2()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer db.ReleaseLock()

	plan, err := db.Plan(`WITH product_sales AS (
		SELECT
		  p.id AS product_id,
		  p.name AS product_name,
		  SUM(o.quantity) AS total_quantity
		FROM
		  products p
		  INNER JOIN orders o ON p.id = o.product_id
		GROUP BY
		  p.id,
		  p.name
	  ), max_sales AS (
		SELECT
		  MAX(total_quantity) AS max_quantity
		FROM
		  product_sales
	  )
	  SELECT
		product_id,
		product_name,
		total_quantity
	  FROM
		product_sales
		INNER JOIN max_sales ON product_sales.total_quantity = max_sales.max_quantity;`)
	if err != nil {
		t.Fatal(err)
	}

	printForest(plan, "   ")

	poly := plan.Polynomial()
	fmt.Println("Expression:", poly)

	res, err := poly.Evaluate(map[string]*big.Float{
		"orders":        big.NewFloat(14321),
		"users":         big.NewFloat(14321),
		"products":      big.NewFloat(14321),
		"product_sales": big.NewFloat(1),
		"max_sales":     big.NewFloat(1),
		"o":             big.NewFloat(14321),
		"u":             big.NewFloat(14321),
		"p":             big.NewFloat(14321),
	})
	if err != nil {
		t.Fatal(err)
	}

	val, _ := res.Int64()

	if val != 1581712 {
		t.Error("result is wrong, should be 1581712")
	}
}

func createTestDB2() (*driver.Connection, error) {
	conn, err := driver.OpenConn("kwil_test_DBHJWBVHJEBVHJEFVHRVH")
	if err != nil {
		return nil, err
	}

	err = conn.AcquireLock()
	if err != nil {
		return nil, err
	}

	err = conn.Execute(`DROP TABLE IF EXISTS users;`)
	if err != nil {
		return nil, err
	}

	err = conn.Execute(`DROP TABLE IF EXISTS products;`)
	if err != nil {
		return nil, err
	}

	err = conn.Execute(`DROP TABLE IF EXISTS orders;`)
	if err != nil {
		return nil, err
	}

	err = conn.Execute(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return nil, err
	}

	err = conn.Execute(`CREATE TABLE products (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		description TEXT,
		price REAL NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		return nil, err
	}

	err = conn.Execute(`CREATE TABLE orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		product_id INTEGER NOT NULL,
		quantity INTEGER NOT NULL CHECK (quantity > 0),
		order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users (id),
		FOREIGN KEY (product_id) REFERENCES products (id)
	);`)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func printForest(forest driver.QueryPlan, indent string) {
	for _, node := range forest {
		printNode(node, indent)
	}
}

func printNode(node *driver.QueryPlanNode, indent string) {
	fmt.Printf("%sId: %d, Parent: %d, NotUsed: %d, Detail: %s\n", indent, node.Id, node.Parent, node.NotUsed, node.Detail)

	childIndent := indent + "  "
	for _, child := range node.Children {
		printNode(child, childIndent)
	}
}
