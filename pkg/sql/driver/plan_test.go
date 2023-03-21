package driver_test

import (
	"fmt"
	"kwil/pkg/sql/driver"
	"math/big"
	"testing"
)

// this test will test query plan generation
func Test_Plan(t *testing.T) {
	db, err := createTest3DB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer db.ReleaseLock()

	// generate plan for insert into users
	plan, err := db.Plan(`WITH latest_comments AS (
		SELECT comments.post_id, MAX(comments.created_at) AS latest_comment_at
		FROM comments
		GROUP BY comments.post_id
	  ),
	  top_posts AS (
		SELECT posts.id, posts.title, posts.created_at, users.username,
			   (SELECT COUNT(comments.id) FROM comments WHERE comments.post_id = posts.id AND comments.created_at >= DATE(posts.created_at)) AS comment_count,
			   (SELECT COUNT(DISTINCT post_tags.tag_id) FROM post_tags WHERE post_tags.post_id = posts.id) AS tag_count,
			   (SELECT COUNT(*) FROM latest_comments WHERE latest_comments.post_id = posts.id) AS recent_comment_count
		FROM posts
		INNER JOIN users ON posts.user_id = users.id
		WHERE posts.created_at >= DATE('2022-01-01')
	  )
	  SELECT top_posts.id, top_posts.title, top_posts.created_at, top_posts.username,
			 top_posts.comment_count, top_posts.tag_count, top_posts.recent_comment_count,
			 GROUP_CONCAT(DISTINCT tags.name) AS tag_names
	  FROM top_posts
	  LEFT JOIN post_tags ON top_posts.id = post_tags.post_id
	  LEFT JOIN tags ON post_tags.tag_id = tags.id
	  GROUP BY top_posts.id
	  HAVING top_posts.comment_count >= 5 AND top_posts.tag_count >= 2 AND top_posts.recent_comment_count >= 1
	  ORDER BY top_posts.recent_comment_count DESC, top_posts.comment_count DESC, top_posts.created_at DESC;`)
	if err != nil {
		t.Fatal(err)
	}

	poly := plan.Polynomial()

	if poly.String() != "((((((((((((1 * (2 * posts)) * log_2((1 * users))) * log_2((1 * post_tags))) * log_2((1 * tags))) * (1 + (1 * (2 * comments)))) * (1 + ((1 * 1) * (2 * post_tags)))) * (1 + ((1 * 1) * (2 * latest_comments)))) * (1 + ((1 * 1) * (2 * post_tags)))) * (1 + ((1 * 1) * (2 * latest_comments)))) * (1 + (1 * (2 * comments)))) * 1) * 1)" {
		t.Error("polynomial is wrong")
	}

	res, err := poly.Evaluate(
		map[string]*big.Float{
			"posts":           big.NewFloat(100000),
			"users":           big.NewFloat(100000),
			"post_tags":       big.NewFloat(100000),
			"tags":            big.NewFloat(100000),
			"comments":        big.NewFloat(100000),
			"latest_comments": big.NewFloat(1),
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	val, _ := res.Int64()
	if val != 9223372036854775807 {
		t.Error("result is wrong, should be overflow")
	}
}

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
